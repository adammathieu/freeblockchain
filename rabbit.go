package main

import (
	"errors"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/streadway/amqp"
)

type Message struct {
	CorrelationId string `protobuf:"bytes,1,opt,name=correlationId,proto3" json:"correlationId,omitempty"`
	ContentType   string `protobuf:"bytes,2,opt,name=contentType,proto3" json:"contentType,omitempty"`
	Exchange      string `protobuf:"bytes,3,opt,name=exchange,proto3" json:"exchange,omitempty"`
	RoutingKey    string `protobuf:"bytes,4,opt,name=routingKey,proto3" json:"routingKey,omitempty"`
	ReplyTo       string `protobuf:"bytes,5,opt,name=replyTo,proto3" json:"replyTo,omitempty"`
	Ttl           int32  `protobuf:"varint,6,opt,name=ttl,proto3" json:"ttl,omitempty"`
	RetryCount    int32  `protobuf:"varint,7,opt,name=retryCount,proto3" json:"retryCount,omitempty"`
	Content       []byte `protobuf:"bytes,15,opt,name=content,proto3" json:"content,omitempty"`
}

// Rabbit defines the structure of rabbitmq
type Rabbit struct {
	url                       string
	connection                *amqp.Connection
	channel                   *amqp.Channel
	internal                  chan interface{}
	semInternal               chan int64
	rabbitCloseError          chan *amqp.Error
	rabbitUndelivarablesError chan amqp.Return
	undeliverablesExchange    string
	undeliverablesQueue       string
}

// Connect try to connect to url and return amqp connection
func Connect(url string) (*amqp.Connection, error) {
	log.Printf("Connecting to rabbitmq on %s\n", url)
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("Could not establish connection to: %s\n %v", url, err)
	}
	return conn, nil
}

// OpenChannel try to create a channel on an establised connection and return amqp channel
func OpenChannel(conn *amqp.Connection, url string) (*amqp.Channel, error) {
	if conn == nil {
		return nil, errors.New("Could not create channel on a nil connection")
	}
	log.Printf("Creating channel for Connection to rabbitmq on %s\n", url)
	channel, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("Could not create channel on established connection to: %s\n %v", url, err)
	}
	return channel, nil
}

// DeclareUndeliverablesQueue creates a queue for Undeliverables message if it doesn't already exist
func (r *Rabbit) DeclareUndeliverablesQueue() error {
	err := r.channel.ExchangeDeclare(
		r.undeliverablesExchange,
		"direct",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("Error while declaring undelivarables Exchange\n %v", err)
	}
	_, err = r.channel.QueueDeclare(
		r.undeliverablesQueue,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("Error while declaring undelivarables Queue\n %v", err)
	}
	err = r.channel.QueueBind(
		r.undeliverablesQueue,    // queue name
		r.undeliverablesQueue,    // routing key
		r.undeliverablesExchange, // exchange
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("Error while binding undelivarables Queue and Exchange\n %v", err)
	}
	return nil
}

// RabbitReconnector check if connection is alive (listener NotifyClose on *amqp.Error channel) and try to reconnect if necessary
// To be runned on a separate coroutine. To be tested with bats integration test
func (r *Rabbit) RabbitReconnector() {
	var rabbitErr *amqp.Error
	var conn *amqp.Connection
	var channel *amqp.Channel
	var err error
	retryTime := 1
	for {
		// Check if we get a CloseError from the NotifyClose channel
		rabbitErr = <-r.rabbitCloseError
		if rabbitErr != nil {
			// Wait a certain among of time before trying to reconnect
			time.Sleep(time.Duration(15+rand.Intn(30)+2*retryTime) * time.Second) //TODO: make duration configurable
			logError("Reconnecting after connection closed", rabbitErr)
			// Try to reconnect to rabbitmq server
			log.Printf("Reconnecting to rabbitmq on %s\n", r.url)
			conn, err = Connect(r.url)
			if err != nil {
				logError("Reconnecting failed", err)
				retryTime++
				log.Printf("Reconnecting attempt: %d\n", retryTime)
			} else {
				r.connection = conn
				log.Printf("Reconnected to rabbitmq on %s\n", r.url)
				// Reattach a listener to CloseError channel to trigger reconnection
				r.rabbitCloseError = make(chan *amqp.Error)
				r.connection.NotifyClose(r.rabbitCloseError)
				// Re open channel
				log.Printf("Reopening channel for connection to rabbitmq on %s\n", r.url)
				channel, err = OpenChannel(r.connection, r.url)
				if err != nil {
					logError("Reopening channel failed", err)
				} else {
					r.channel = channel
				}
			}
		}
	}
}

// RabbitUndelivarablesHandler check if published messages are not undelivarables (listener NotifyReturn on amqp.Return channel) and try to publish it on undelivarables queue.
// To be runned on a separate goroutine. To be tested with bats integration test
func (r *Rabbit) RabbitUndelivarablesHandler(chReturn <-chan amqp.Return) {
	for undeliveredMsg := range chReturn {
		log.Printf("Could not deliver %s with routing key %s", undeliveredMsg.CorrelationId, undeliveredMsg.RoutingKey)
		msg := Message{"100", "json", "xXx", "CeMatin", "Me", 1, 1, []byte{}} //TODO replace msg with value from Return
		ProcessMsg(msg)
	}
}

// ProcessMsg handle the message to be published  on rabbitmq
func ProcessMsg(msg Message) {
	fmt.Println(msg) //TODO use amqp.Publish()
}

// GetInternalIPC2RabbitChannel return rabbit isntance internal queue
func (r *Rabbit) GetInternalIPC2RabbitChannel() chan interface{} {
	return r.internal
}

// Publisher read IPC2Rabbit internal channel for messages to be published on rabbitmq
func (r *Rabbit) Publisher() {
	for {
		for msg := range r.internal {
			r.semInternal <- 1
			switch msg.(type) {
			case Message:
				msg := msg // Create new instance of msg for the goroutine.
				go func() {
					ProcessMsg(msg.(Message))
					<-r.semInternal
				}()
			}
		}
	}
}

// NewRabbit create a new Rabbitmq instance
func NewRabbit(host string, port string, login string, password string, undeliverablesQueueName string, undeliverablesExchangeName string) (*Rabbit, error) {
	// Create a new Rabbit instance
	r := new(Rabbit)
	// Store url for connection
	r.url = "amqp://" + login + ":" + password + "@" + host + ":" + port
	// Try to connect to rabbitmq server
	conn, err := Connect(r.url)
	if err != nil {
		return nil, err
	}
	r.connection = conn
	// Try to open a channel
	channel, err := OpenChannel(r.connection, r.url)
	if err != nil {
		return nil, err
	}
	r.channel = channel
	// Create the rabbit error channel to monitor closing connection/channel error
	r.rabbitCloseError = make(chan *amqp.Error)
	r.connection.NotifyClose(r.rabbitCloseError)
	// Declare undelivarbles queue and exchange
	r.undeliverablesQueue = undeliverablesQueueName
	r.undeliverablesExchange = undeliverablesExchangeName
	err = r.DeclareUndeliverablesQueue()
	if err != nil {
		return nil, err
	}
	// Create IPC internal channel for publishing
	r.internal = make(chan interface{})
	r.semInternal = make(chan int64, 4)
	// Create the rabbit return channel to monitor undelivarables publish error
	r.rabbitUndelivarablesError = make(chan amqp.Return)
	r.channel.NotifyReturn(r.rabbitUndelivarablesError)
	// Return Rabbit instance
	return r, nil
}

func logError(message string, err error) {
	if err != nil {
		log.Printf("%s: %s", message, err)
	}
}
