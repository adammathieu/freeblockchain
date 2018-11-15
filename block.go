package main

import (
	"bytes"
	"crypto/sha256"
	"strconv"
	"time"
)

// Block defines the structure of a block
type Block struct {
	Timestamp     int64
	Data          []byte
	PrevBlockHash []byte
	Hash          []byte
	Counter       int64
	HashCounter   []byte
}

// SetHash calculate and set the Hash of a block
func (b *Block) SetHash() {
	timestamp := []byte(strconv.FormatInt(b.Timestamp, 10))
	headers := bytes.Join([][]byte{b.PrevBlockHash, b.Data, timestamp}, []byte{})
	hash := sha256.Sum256(headers)

	b.Hash = hash[:]
}

// SetCounterHash increment the check counter and generate a new Hash Counter
func (b *Block) SetCounterHash() {
	b.Counter++
	timestamp := []byte(strconv.FormatInt(b.Timestamp, 10))
	counter := []byte(strconv.FormatInt(b.Counter, 10))
	headers := bytes.Join([][]byte{counter, b.Hash, b.PrevBlockHash, b.Data, timestamp}, []byte{})
	hash := sha256.Sum256(headers)

	b.HashCounter = hash[:]
}

// NewBlock create a new bloc with data and previous bloc hash as arguments.
func NewBlock(data string, prevBlockHash []byte) *Block {
	block := &Block{time.Now().Unix(), []byte(data), prevBlockHash, []byte{}, 0, []byte{}}
	block.SetHash()
	block.SetCounterHash()
	return block
}
