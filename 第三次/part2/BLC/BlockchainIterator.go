package BLC

import (
	"github.com/boltdb/bolt"
	"log"
)

type BlockchainIterator struct {
	//当前hash
	CurrentHash []byte
	//数据库
	DB *bolt.DB
}

//往下迭代
func (blockchainIterator *BlockchainIterator) Next() *Block {

	var block *Block

	err := blockchainIterator.DB.View(func(tx *bolt.Tx) error {

		bucket := tx.Bucket([]byte(blockTableName))
		if bucket != nil {

			//获取当前迭代器 CurrentHash 对应的区块
			blockBytes := bucket.Get(blockchainIterator.CurrentHash)
			block = Deserialize(blockBytes)

			//更新迭代器CurrentHash
			blockchainIterator.CurrentHash = block.PrevBlockHash
		}

		return nil
	})

	if err != nil {
		log.Panic(err)
	}

	return block
}
