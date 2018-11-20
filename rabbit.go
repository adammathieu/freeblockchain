package main

import (
	"errors"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/streadway/amqp"
)

// Rabbit defines the structure of rabbitmq
type Rabbit struct {
	url              string
	connection       *amqp.Connection
	rabbitCloseError chan *amqp.Error
	// channel                *amqp.Channel
	// undeliverablesExchange string
	// undeliverablesQueue    string
}

// Connect try to connect to url and return amqp connection
func Connect(url string) (*amqp.Connection, error) {
	if url == "" {
		return nil, errors.New("Could not establish connection to an empty url")
	}
	log.Printf("Connecting to rabbitmq on %s\n", url)
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("Could not establish connection to: %s, %v", url, err)
	}
	return conn, nil
}

// RabbitReconnector check if connection is alive (listener on *amqp.Error NotifyClose channel) and try to reconnect if necessary
// To be runned on a separate coroutine. To be tested with bats integration test
func (r *Rabbit) RabbitReconnector(url string) {
	var rabbitErr *amqp.Error
	var conn *amqp.Connection
	var err error
	retryTime := 1
	for {
		//Check if we get a CloseError from the NotifyClose channel
		rabbitErr = <-r.rabbitCloseError
		if rabbitErr != nil {
			//Wait a certain among of time before trying to reconnect
			time.Sleep(time.Duration(15+rand.Intn(30)+2*retryTime) * time.Second) //TODO: make duration configurable
			logError("Reconnecting after connection closed", rabbitErr)
			//Try to reconnect to rabbitmq server
			log.Printf("Reconnecting to rabbitmq on %s\n", r.url)
			conn, err = Connect(r.url)
			if err != nil {
				logError("Reconnecting failed", err)
				retryTime++
				log.Printf("Reconnecting attempt: %d\n", retryTime)
			} else {
				r.connection = conn
				log.Printf("Reconnected to rabbitmq on %s\n", r.url)
				//Reattach a listener to CloseError channel to trigger reconnection
				r.rabbitCloseError = make(chan *amqp.Error)
				r.connection.NotifyClose(r.rabbitCloseError)
			}
		}
	}
}

// NewRabbit create a new Rabbitmq structure
func NewRabbit(host string, port string, login string, password string) (*Rabbit, error) {
	r := new(Rabbit)
	r.url = "amqp://" + login + ":" + password + "@" + host + ":" + port
	conn, err := Connect(r.url)
	if err != nil {
		return nil, fmt.Errorf("Could not establish connection: %v", err)
	}
	r.connection = conn
	r.rabbitCloseError = make(chan *amqp.Error)
	return r, nil
}

func logError(message string, err error) {
	if err != nil {
		log.Printf("%s: %s", message, err)
	}
}
