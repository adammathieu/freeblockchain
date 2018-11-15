package main

import "fmt"

func main() {
	bc := NewBlockchain()

	bc.AddBlock("Premier Bloc")
	bc.AddBlock("Second Bloc")

	for _, block := range bc.blocks {
		fmt.Printf("Previous Hash: %x\n", block.PrevBlockHash)
		fmt.Printf("Data: %s\n", block.Data)
		fmt.Printf("Hash: %x\n", block.Hash)
	}
}
