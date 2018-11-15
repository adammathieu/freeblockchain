package main

import "testing"

func TestNewBlock(t *testing.T) {
	prevBlockHash := []byte{'e', '7', '9', '5', 'b', '3', 'c', '1', '4', 'b', '4', '6', 'd', '5', 'f', '6', '2', '4', 'c', 'a', '9', 'd', '5', 'd', '2', 'c', 'a', '0', '0', 'f', '6', '3', 'd', '7', 'b', '2', '2', 'c', '7', '1', '5', '5', '1', '3', 'c', '2', 'b', 'a', '6', '9', 'e', 'b', '1', 'a', 'a', 'd', '7', 'f', 'a', '9', 'b', '5', '1', 'f'}
	var data = "test block"
	b := NewBlock(data, prevBlockHash)

	if string(b.Data[:]) != data {
		t.Error("Block not initialize with data passed as argument.")
	}
	if b.Counter != 1 {
		t.Error("Block counter haven't been increment when generating Counter Hash.")
	}
	if b.Hash == nil {
		t.Error("Block hash should not be null.")
	}
	if b.HashCounter == nil {
		t.Error("Block counter hash should not be null.")
	}
}
