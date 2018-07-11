package BLC

import (
	"github.com/boltdb/bolt"
	"log"
)

type BlockchainIterator struct {
	//当前hash
	ZQ_CurrentHash []byte
	//数据库
	ZQ_DB *bolt.DB
}

//往下迭代
func (blockchainIterator *BlockchainIterator) ZQ_Next() *Block {

	var block *Block

	err := blockchainIterator.ZQ_DB.View(func(tx *bolt.Tx) error {

		bucket := tx.Bucket([]byte(blockTableName))
		if bucket != nil {

			//获取当前迭代器 ZQ_CurrentHash 对应的区块
			blockBytes := bucket.Get(blockchainIterator.ZQ_CurrentHash)
			block = Deserialize(blockBytes)

			//更新迭代器CurrentHash
			blockchainIterator.ZQ_CurrentHash = block.ZQ_PrevBlockHash
		}

		return nil
	})

	if err != nil {
		log.Panic(err)
	}

	return block
}
