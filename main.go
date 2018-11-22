package main

import (
	"log"
	"os"
)

func main() {
	// blockChannel := make(chan string)
	// bc := NewBlockchain(blockChannel)

	// go func() {
	// 	bc.blockChannel <- "Premier Bloc"
	// 	bc.blockChannel <- "Second Bloc"
	// }()

	// bc.ReadBlockChannel()
	// bc.ReadBlockChannel()

	// for _, block := range bc.blocks {
	// 	fmt.Printf("Previous Hash: %x\n", block.PrevBlockHash)
	// 	fmt.Printf("Data: %s\n", block.Data)
	// 	fmt.Printf("Hash: %x\n", block.Hash)
	// }

	lapin, err := NewRabbit("localhost", "5672", "guest", "guest", "errQueue", "errX")
	if err != nil {
		log.Printf("%v", err)
		os.Exit(-1)
	}
	go lapin.RabbitReconnector()
	go lapin.Publisher()
	for {
		lapin.internal <- Message{"100", "json", "xXx", "CeMatin", "Me", 1, 1, []byte{}}
		//time.Sleep(time.Duration(10) * time.Second)
	}
}
