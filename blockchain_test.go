package main

import "testing"

func TestNewBlockchain(t *testing.T) {
	blockChannel := make(chan string)
	bc := NewBlockchain(blockChannel)

	if string(bc.blocks[0].Data[:]) != "Genesis Block" {
		t.Error("Blockchain not initialize with new genesis block.")
	}
}

func TestAddBlock(t *testing.T) {
	blockChannel := make(chan string)
	bc := NewBlockchain(blockChannel)
	bc.AddBlock("First Block")

	if string(bc.blocks[1].Data[:]) != "First Block" {
		t.Error("Data of the new block added to blockchain do not match data used as argument.")
	}
}

func TestReadBlockChannel(t *testing.T) {
	blockChannel := make(chan string)
	bc := NewBlockchain(blockChannel)

	go func() {
		bc.blockChannel <- "First Block"
	}()
	bc.ReadBlockChannel()

	if string(bc.blocks[1].Data[:]) != "First Block" {
		t.Error("Data of the new block added to blockchain do not match data used as argument.")
	}
}
