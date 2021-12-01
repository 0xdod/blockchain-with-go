package main

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"strconv"
	"strings"
	"time"
)

type Block struct {
	Timestamp     int64
	Data          []byte
	PrevBlockHash []byte
	Hash          []byte
	Nonce         int
}

func NewBlock(data string, prevBlockHash []byte) *Block {
	block := Block{
		Timestamp:     time.Now().Unix(),
		Data:          []byte(data),
		PrevBlockHash: prevBlockHash,
	}

	pow := NewProofOfWork(&block)
	nonce, hash := pow.Run()
	block.Hash = hash[:]
	block.Nonce = nonce

	return &block
}

func NewGenesisBlock() *Block {
	return NewBlock("Genesis block", []byte{})
}

func (b Block) String() string {
	sb := strings.Builder{}
	sb.WriteString(fmt.Sprintf("Prev. hash: %x\n", b.PrevBlockHash))
	sb.WriteString(fmt.Sprintf("Data: %s\n", b.Data))
	sb.WriteString(fmt.Sprintf("Hash: %x\n", b.Hash))
	pow := NewProofOfWork(&b)
	sb.WriteString(fmt.Sprintf("PoW: %s\n", strconv.FormatBool(pow.Validate())))
	sb.WriteString("\n")
	return sb.String()
}

func (b Block) Serialize() []byte {
	buf := &bytes.Buffer{}
	_ = gob.NewEncoder(buf).Encode(b)

	return buf.Bytes()
}

func Deserialize(p []byte) *Block {
	block := Block{}
	_ = gob.NewDecoder(bytes.NewReader(p)).Decode(&block)

	return &block
}
