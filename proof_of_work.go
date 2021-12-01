package main

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"math"
	"math/big"
)

const (
	// in bitcoin `target bits` is the block header storing the difficulty at which the block was mined.
	// 24 is an arbitary number. The bigger the targetBit, the smaller our target, the more difficult the work.
	targetBits = 24
	// our goal is to have a target that takes less than 256 bits in memory
	maxTargetBits = 256
	maxNonce      = math.MaxInt64
)

type ProofOfWork struct {
	block *Block

	// target is another name for requirement, we will compare this with the hash we get
	// from our computation to verify our work
	target *big.Int
}

func NewProofOfWork(b *Block) *ProofOfWork {
	target := big.NewInt(1)
	target.Lsh(target, uint(maxTargetBits-targetBits))
	pow := ProofOfWork{b, target}
	return &pow
}

func (pow *ProofOfWork) prepareData(nonce int) []byte {
	data := bytes.Join(
		[][]byte{
			pow.block.PrevBlockHash,
			pow.block.Data,
			IntToHex(pow.block.Timestamp),
			IntToHex(int64(targetBits)),
			IntToHex(int64(nonce)),
		},
		[]byte{},
	)
	return data
}

func (pow *ProofOfWork) Run() (int, []byte) {
	var hashInt big.Int
	var hash [32]byte
	nonce := 0

	fmt.Printf("Mining block containing %q\n", pow.block.Data)
	for nonce < maxNonce {
		data := pow.prepareData(nonce)
		hash = sha256.Sum256(data)
		fmt.Printf("\r%x", hash)
		hashInt.SetBytes(hash[:])

		if hashInt.Cmp(pow.target) == -1 {
			break
		} else {
			nonce++
		}
	}

	fmt.Print("\n\n")

	return nonce, hash[:]
}

func (pow *ProofOfWork) Validate() bool {
	var hashInt big.Int

	data := pow.prepareData(pow.block.Nonce)
	hash := sha256.Sum256(data)
	hashInt.SetBytes(hash[:])
	isValid := hashInt.Cmp(pow.target) == -1

	return isValid
}

func IntToHex(i int64) []byte {
	return []byte(fmt.Sprintf("%x", i))
}
