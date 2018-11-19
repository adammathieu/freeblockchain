package main

import (
	"fmt"
)

func main() {
	blockChannel := make(chan string)
	bc := NewBlockchain(blockChannel)

	go func() {
		bc.blockChannel <- "Premier Bloc"
		bc.blockChannel <- "Second Bloc"
	}()

	bc.ReadBlockChannel()
	bc.ReadBlockChannel()

	for _, block := range bc.blocks {
		fmt.Printf("Previous Hash: %x\n", block.PrevBlockHash)
		fmt.Printf("Data: %s\n", block.Data)
		fmt.Printf("Hash: %x\n", block.Hash)
	}
}
