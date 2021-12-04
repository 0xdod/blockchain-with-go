package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"fmt"
)

const subsidy = 10

type Transaction struct {
	ID      []byte
	Inputs  []TXInput
	Outputs []TXOutput
}

type Transactions []*Transaction

type TXOutput struct {
	Value        int
	ScriptPubKey string
}

type TXInput struct {
	TXID      []byte
	Vout      int // value of output referenced
	ScriptSig string
}

func NewCoinbaseTX(to, data string) *Transaction {
	if data == "" {
		data = fmt.Sprintf("Reward to %q", to)
	}

	txIn := TXInput{[]byte{}, -1, data}
	txOut := TXOutput{subsidy, to}
	tx := Transaction{nil, []TXInput{txIn}, []TXOutput{txOut}}
	tx.SetID()

	return &tx
}

func (tx Transaction) IsCoinbase() bool {
	return len(tx.Inputs) == 1 && len(tx.Inputs[0].TXID) == 0
}

func (tx *Transaction) SetID() {
	buf := new(bytes.Buffer)
	var hash [32]byte
	if err := gob.NewEncoder(buf).Encode(tx); err != nil {
		panic(err)
	}
	hash = sha256.Sum256(buf.Bytes())
	tx.ID = hash[:]
}

func (txi TXInput) CanUnlockOutputWith(unlockingData string) bool {
	return txi.ScriptSig == unlockingData
}

func (txo TXOutput) CanBeUnlockedWith(unlockingData string) bool {
	return txo.ScriptPubKey == unlockingData
}

func (tx Transaction) String() string {
	return string(tx.ID)
}

func (txs Transactions) String() string {
	return "hello"
}
