package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"fmt"
	"strconv"
	"strings"
	"time"
)

type Block struct {
	Timestamp     int64
	Transactions  Transactions
	PrevBlockHash []byte
	Hash          []byte
	Nonce         int
}

func NewBlock(txs []*Transaction, prevBlockHash []byte) *Block {
	block := Block{
		Timestamp:     time.Now().Unix(),
		Transactions:  txs,
		PrevBlockHash: prevBlockHash,
	}

	pow := NewProofOfWork(&block)
	nonce, hash := pow.Run()
	block.Hash = hash[:]
	block.Nonce = nonce

	return &block
}

func NewGenesisBlock(coinbase *Transaction) *Block {
	return NewBlock([]*Transaction{coinbase}, []byte{})
}

func (b Block) String() string {
	sb := strings.Builder{}
	sb.WriteString(fmt.Sprintf("Prev. hash: %x\n", b.PrevBlockHash))
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

func (b Block) HashTransactions() []byte {
	var txHashes [][]byte
	var txHash [32]byte

	for _, tx := range b.Transactions {
		txHashes = append(txHashes, tx.ID)
	}

	txHash = sha256.Sum256(bytes.Join(txHashes, []byte{}))

	return txHash[:]
}

func Deserialize(p []byte) *Block {
	block := Block{}
	_ = gob.NewDecoder(bytes.NewReader(p)).Decode(&block)

	return &block
}

func (b *Block) IsGenesis() bool {
	if b != nil && len(b.PrevBlockHash) == 0 {
		return true
	}

	return false
}
