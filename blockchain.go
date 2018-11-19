package main

// Blockchain defines the structure of a blockchain
type Blockchain struct {
	blocks       []*Block
	blockChannel chan string
}

// ReadBlockChannel read data from a channel and add a block to the blockchain
func (bc *Blockchain) ReadBlockChannel() {
	select {
	case data := <-bc.blockChannel:
		bc.AddBlock(data)
	}
}

// AddBlock add a block to the blockchain
func (bc *Blockchain) AddBlock(data string) {
	previousBlock := bc.blocks[len(bc.blocks)-1]
	newBlock := NewBlock(data, previousBlock.Hash)
	bc.blocks = append(bc.blocks, newBlock)
}

// NewGenesisBlock create a new genesis block
func NewGenesisBlock() *Block {
	return NewBlock("Genesis Block", []byte{})
}

// NewBlockchain create a new blockchain
func NewBlockchain(channel chan string) *Blockchain {
	return &Blockchain{[]*Block{NewGenesisBlock()}, channel}
}
