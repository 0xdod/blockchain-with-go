package main

import (
	"encoding/hex"
	"errors"
	"io/fs"
	"log"
	"os"

	"github.com/boltdb/bolt"
)

const (
	blocksBucket        = "blocks"
	dbFile              = "./database.boltdb"
	genesisCoinbaseData = "The Times 03/Jan/2009 Chancellor on brink of second bailout for banks"
)

var (
	lastHashKey     = []byte("l")
	blocksBucketKey = []byte(blocksBucket)
)

type Blockchain struct {
	tip []byte
	db  *bolt.DB
}

type BlockchainIterator struct {
	currentHash []byte
	db          *bolt.DB
	block       *Block
}

func (bc *Blockchain) MineBlock(transactions []*Transaction) {
	lastHash := make([]byte, 0)

	err := bc.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(blocksBucketKey)
		lastHash = bucket.Get(lastHashKey)

		return nil
	})

	if err != nil {
		panic(err)
	}

	newBlock := NewBlock(transactions, lastHash)

	err = bc.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(blocksBucketKey)
		err = bucket.Put(newBlock.Hash, newBlock.Serialize())

		if err != nil {
			return err
		}

		err = bucket.Put(lastHashKey, newBlock.Hash)

		if err != nil {
			return err
		}

		bc.tip = newBlock.Hash

		return nil
	})

	if err != nil {
		panic(err)
	}

}

func (bc *Blockchain) Iterator() *BlockchainIterator {
	return &BlockchainIterator{bc.tip, bc.db, nil}
}

func NewBlockchain() *Blockchain {
	if !dbExists() {
		log.Fatal("no exisiting blockchain found, create one first.")
	}

	tip := make([]byte, 0)
	db, err := bolt.Open(dbFile, 0600, nil)

	if err != nil {
		panic(err)
	}

	err = db.Update(func(tx *bolt.Tx) error {

		bucket := tx.Bucket([]byte(blocksBucket))
		tip = bucket.Get(lastHashKey)

		return nil
	})

	if err != nil {
		panic(err)
	}

	return &Blockchain{tip, db}
}

func dbExists() bool {
	_, err := os.Stat(dbFile)
	return !errors.Is(err, fs.ErrNotExist)
}

func CreateBlockchain(address string) *Blockchain {
	if dbExists() {
		log.Fatalln("blockchain exists")
	}

	var tip []byte
	db, err := bolt.Open(dbFile, 0600, nil)

	if err != nil {
		panic(err)
	}

	err = db.Update(func(tx *bolt.Tx) error {
		cbtx := NewCoinbaseTX(address, genesisCoinbaseData)
		genesis := NewGenesisBlock(cbtx)

		bucket, err := tx.CreateBucket(blocksBucketKey)

		if err != nil {
			return err
		}

		err = bucket.Put(genesis.Hash, genesis.Serialize())

		if err != nil {
			return err
		}

		err = bucket.Put(lastHashKey, genesis.Hash)

		if err != nil {
			return err
		}

		tip = genesis.Hash

		return nil
	})

	if err != nil {
		panic(err)
	}

	return &Blockchain{tip, db}
}

func (bci *BlockchainIterator) Scan() bool {
	if bci.block.IsGenesis() {
		return false
	}

	var block *Block

	err := bci.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(blocksBucketKey)
		encodedBlock := bucket.Get(bci.currentHash)
		block = Deserialize(encodedBlock)

		return nil
	})

	if err != nil {
		panic(err)
	}

	bci.block = block
	bci.currentHash = block.PrevBlockHash

	return true
}

func (bci *BlockchainIterator) Next() *Block {
	return bci.block
}

func (bc *Blockchain) FindUnspentTransactions(address string) []Transaction {
	var unspentTxs []Transaction
	spentTXOs := make(map[string][]int)
	bci := bc.Iterator()

	for bci.Scan() {
		block := bci.Next()

		for _, tx := range block.Transactions {
			txID := string(tx.ID)

		Outputs:
			for outIndex, out := range tx.Outputs {
				if spentTXOs[txID] != nil {
					for _, spentOutput := range spentTXOs[txID] {
						if spentOutput == outIndex {
							continue Outputs
						}
					}
				}

				if out.CanBeUnlockedWith(address) {
					unspentTxs = append(unspentTxs, *tx)
				}
			}

			if !tx.IsCoinbase() {
				for _, inp := range tx.Inputs {
					if inp.CanUnlockOutputWith(address) {
						inTxID := string(inp.TXID)
						spentTXOs[inTxID] = append(spentTXOs[inTxID], inp.Vout)
					}
				}
			}
		}
	}

	return unspentTxs
}

func (bc *Blockchain) FindUTXO(address string) []TXOutput {
	var UTXOs []TXOutput
	unspentTxs := bc.FindUnspentTransactions(address)

	for _, tx := range unspentTxs {
		for _, output := range tx.Outputs {
			if output.CanBeUnlockedWith(address) {
				UTXOs = append(UTXOs, output)
			}
		}
	}

	return UTXOs
}

func NewUTXOTransaction(from, to string, amount int, bc *Blockchain) *Transaction {
	inputs, outputs := make([]TXInput, 0), make([]TXOutput, 0)

	acc, validOutputs := bc.FindSpendableOutputs(from, amount)

	if acc < amount {
		log.Fatal("ERROR: Not enough funds")
	}

	for txid, outs := range validOutputs {
		txID, err := hex.DecodeString(txid)

		if err != nil {
			panic(err)
		}

		for _, out := range outs {
			input := TXInput{txID, out, from}
			inputs = append(inputs, input)
		}
	}

	outputs = append(outputs, TXOutput{amount, to})

	if acc > amount {
		outputs = append(outputs, TXOutput{acc - amount, from})
	}

	tx := Transaction{nil, inputs, outputs}
	tx.SetID()

	return &tx
}

func (bc *Blockchain) FindSpendableOutputs(address string, amount int) (int, map[string][]int) {
	unspentOutputs := make(map[string][]int)
	unspentTxs := bc.FindUnspentTransactions(address)
	accumulated := 0

Work:
	for _, tx := range unspentTxs {
		txID := hex.EncodeToString(tx.ID)

		for outIndex, output := range tx.Outputs {
			if output.CanBeUnlockedWith(address) && accumulated < amount {
				accumulated += output.Value
				unspentOutputs[txID] = append(unspentOutputs[txID], outIndex)

				if accumulated >= amount {
					break Work
				}
			}
		}
	}
	return accumulated, unspentOutputs
}
