package main

import "github.com/boltdb/bolt"

const (
	blocksBucket = "blocks"
	dbFile       = "./database.boltdb"
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
}

func (bc *Blockchain) AddBlock(data string) {
	lastHash := make([]byte, 0)

	err := bc.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(blocksBucketKey)
		lastHash = bucket.Get(lastHashKey)

		return nil
	})

	if err != nil {
		panic(err)
	}

	newBlock := NewBlock(data, lastHash)

	err = bc.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(blocksBucketKey)
		err = bucket.Put(newBlock.Hash, newBlock.Serialize())
		err = bucket.Put(lastHashKey, newBlock.Hash)
		bc.tip = newBlock.Hash

		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		panic(err)
	}

}

func (bc *Blockchain) Iterator() *BlockchainIterator {
	return &BlockchainIterator{bc.tip, bc.db}
}

func NewBlockchain() *Blockchain {
	tip := make([]byte, 0)

	db, err := bolt.Open(dbFile, 0600, nil)

	err = db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(blocksBucket))

		if bucket == nil {
			genesis := NewGenesisBlock()
			bucket, err = tx.CreateBucket([]byte(blocksBucket))
			err = bucket.Put(genesis.Hash, genesis.Serialize())
			err = bucket.Put(lastHashKey, genesis.Hash)
			tip = genesis.Hash
		} else {
			tip = bucket.Get(lastHashKey)
		}

		if err != nil {
			return err
		}

		return nil
	})

	return &Blockchain{tip, db}
}

func (bci *BlockchainIterator) Next() *Block {
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

	bci.currentHash = block.PrevBlockHash

	return block
}
