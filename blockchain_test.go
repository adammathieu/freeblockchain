package main

import "testing"

func TestNewBlockchain(t *testing.T) {
	bc := NewBlockchain()

	if string(bc.blocks[0].Data[:]) != "Genesis Block" {
		t.Error("Blockchain not initialize with new genesis block.")
	}
}

func TestAddBlock(t *testing.T) {
	bc := NewBlockchain()
	bc.AddBlock("First Block")

	if string(bc.blocks[1].Data[:]) != "First Block" {
		t.Error("Data of the new block added to blockchain do not match data used as argument.")
	}
}
