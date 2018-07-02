package BLC

import (
	"github.com/boltdb/bolt"
	"log"
	"math/big"
	"fmt"
	"time"
	"os"
)

type Blockchain struct {
	//最新区块hash
	Tip []byte
	//数据库
	DB *bolt.DB
}

//数据库名称 表名
const dbName = "blockchain.db"
const blockTableName = "blocks"

//创建迭代器
func (blockchain *Blockchain) Iterator() *BlockchainIterator {

	return &BlockchainIterator{blockchain.Tip, blockchain.DB}
}

//打印区块链
func (blockchain *Blockchain) Printchain() {

	blockchainIterator := blockchain.Iterator()

	var hashInt big.Int

	for {
		block := blockchainIterator.Next()

		fmt.Printf("Height: %d\n", block.Height)
		fmt.Printf("PrevBlockHash: %x\n", block.PrevBlockHash)
		fmt.Printf("Height: %s\n", block.Data)
		fmt.Printf("Timestamp: %s\n", time.Unix(block.Timestamp, 0).Format("2006-01-02 03:04:05 PM"))
		fmt.Printf("Hash: %x\n", block.Hash)
		fmt.Printf("Nonce: %d\n", block.Nonce)
		fmt.Println()

		hashInt.SetBytes(block.PrevBlockHash)

		//如果上一个hash是否是创世区块 break
		if big.NewInt(0).Cmp(&hashInt) == 0 {
			break
		}
	}
}

//增加区块至区块链
func (blockchain *Blockchain) AddBlockToBlockchain(data string) {

	//添加区块到数据库
	err := blockchain.DB.Update(func(tx *bolt.Tx) error {

		bucket := tx.Bucket([]byte(blockTableName))
		if bucket != nil {

			//读取最新区块
			blockBytes := bucket.Get(blockchain.Tip)
			block := Deserialize(blockBytes)

			//创建新区块
			newBlock := NewBlock(data, block.Height+1, block.Hash)

			err := bucket.Put(newBlock.Hash, newBlock.Serialize())
			if err != nil {
				log.Panic(err)
			}

			err = bucket.Put([]byte("lastHash"), newBlock.Hash)
			if err != nil {
				log.Panic(err)
			}

			blockchain.Tip = newBlock.Hash
		}
		return nil
	})

	if err != nil {
		log.Panic(err)
	}
}

//数据库是否存在
func DBExists() bool {
	if _, err := os.Stat(dbName); os.IsNotExist(err) {
		return false
	}

	return true
}

//创建带有创世区块的区块链
func CreateBlockchainWithGenesisBlock(data string) {

	//数字库是否存在
	if DBExists() {
		fmt.Println("创世区块已经存在...")
		os.Exit(1)
	}

	fmt.Println("正在创建创世区块...")

	//尝试打开/创建 数据库
	db, err := bolt.Open(dbName, 0600, nil)
	if err != nil {
		log.Fatal(err)
	}

	err = db.Update(func(tx *bolt.Tx) error {

		bucket, err := tx.CreateBucket([]byte(blockTableName))

		if err != nil {
			log.Panic(err)
		}

		if bucket != nil {
			//创建创世区块
			genesisBlock := CreateGenesisBlock(data)

			//存入数据 hash => 序列化区块
			err = bucket.Put(genesisBlock.Hash, genesisBlock.Serialize())
			if err != nil {
				log.Panic(err)
			}

			//存储最新的hash
			err = bucket.Put([]byte("lastHash"), genesisBlock.Hash)
			if err != nil {
				log.Panic(err)
			}

		}

		return nil
	})

	if err != nil {
		log.Panic(err)
	}
}

//获取区块链对象
func GetBlockchainObject() *Blockchain {

	//尝试打开/创建 数据库
	db, err := bolt.Open(dbName, 0600, nil)
	if err != nil {
		log.Fatal(err)
	}

	//blockchain中最新hash
	var lastHash []byte

	err = db.View(func(tx *bolt.Tx) error {

		bucket := tx.Bucket([]byte(blockTableName))

		if bucket != nil {

			lastHash = bucket.Get([]byte("lastHash"))
		}

		return nil
	})

	if err != nil {
		log.Panic(err)
	}

	return &Blockchain{lastHash, db}
}
