package BLC

import (
	"github.com/boltdb/bolt"
	"log"
	"math/big"
	"fmt"
	"time"
	"os"
	"strconv"
	"encoding/hex"
	"crypto/ecdsa"
	"bytes"
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
		fmt.Printf("Timestamp: %s\n", time.Unix(block.Timestamp, 0).Format("2006-01-02 03:04:05 PM"))
		fmt.Printf("Hash: %x\n", block.Hash)
		fmt.Printf("Nonce: %d\n", block.Nonce)
		fmt.Println("Txs:")
		for _, tx := range block.Txs {

			fmt.Printf("\tTxHash: %x\n", tx.TxHash)

			fmt.Println("\tVins:")
			for _, in := range tx.Vins {
				fmt.Printf("\t\tTxHash: %x\n", in.TxHash)
				fmt.Printf("\t\tVout: %d\n", in.Vout)
				fmt.Printf("\t\tPublicKey: %v\n", in.PublicKey)
				fmt.Printf("\t\tSignature: %v\n", in.Signature)
			}

			fmt.Println("\tVouts:")
			for _, out := range tx.Vouts {
				fmt.Printf("\t\tMoney: %d\n", out.Value)
				fmt.Printf("\t\tRipemd160Hash: %v\n", out.Ripemd160Hash)
			}

		}

		fmt.Println()

		hashInt.SetBytes(block.PrevBlockHash)

		//如果上一个hash是否是创世区块 break
		if big.NewInt(0).Cmp(&hashInt) == 0 {
			break
		}
	}
}

//增加区块至区块链
func (blockchain *Blockchain) AddBlockToBlockchain(txs []*Transaction) {

	//添加区块到数据库
	err := blockchain.DB.Update(func(tx *bolt.Tx) error {

		bucket := tx.Bucket([]byte(blockTableName))
		if bucket != nil {

			//读取最新区块
			blockBytes := bucket.Get(blockchain.Tip)
			block := Deserialize(blockBytes)

			//创建新区块
			newBlock := NewBlock(txs, block.Height+1, block.Hash)

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
func CreateBlockchainWithGenesisBlock(address string) *Blockchain {

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

	var genesisHash []byte

	err = db.Update(func(tx *bolt.Tx) error {

		bucket, err := tx.CreateBucket([]byte(blockTableName))

		if err != nil {
			log.Panic(err)
		}

		if bucket != nil {

			//创建coinbase的 Transaction
			txCoinbase := NewCoinbaseTransaction(address)

			//创建创世区块
			genesisBlock := CreateGenesisBlock(txCoinbase)

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

			genesisHash = genesisBlock.Hash
		}

		return nil
	})

	if err != nil {
		log.Panic(err)
	}

	return &Blockchain{genesisHash, db}
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

//找出用户TXOutput列表
func (blockchain *Blockchain) UnSpentTransactionsWithAddress(address string) []*TXOutput {

	var unUTXOs []*TXOutput

	//address下所对应的vins
	spentTXOutputs := make(map[string][]int)

	//遍历数据库
	iterator := blockchain.Iterator()
	var hasInt big.Int

	for {
		block := iterator.Next()
		fmt.Println(block)
		fmt.Println()

		hasInt.SetBytes(block.PrevBlockHash)
		if hasInt.Cmp(big.NewInt(0)) == 0 {
			break
		}

		for _, tx := range block.Txs {

			if tx.IsCoinbaseTransaction() == false {

				for _, in := range tx.Vins {

					publicKeyHash := Base58Decode([]byte(address))
					ripemd160Hash := publicKeyHash[1 : len(publicKeyHash)-4]

					//是否能够解锁
					if in.UnLockWithRipemd160Hash(ripemd160Hash) {

						key := hex.EncodeToString(in.TxHash)
						spentTXOutputs[key] = append(spentTXOutputs[key], in.Vout)
					}
				}
			}

			for index, out := range tx.Vouts {

				//是否能够解锁
				if out.UnLockScriptPubKeyWithAddress([]byte(address)) {

					if spentTXOutputs != nil {

						for txHash, indexArray := range spentTXOutputs {

							if txHash == hex.EncodeToString(tx.TxHash) {

								for _, i := range indexArray {

									if index == i {
										continue
									} else {
										unUTXOs = append(unUTXOs, out)
									}
								}
							}
						}
					}
				}
			}

			//fmt.Prin
		}
	}

	return nil
}

//地址对应的所有未花费
func (blockchain *Blockchain) UnUTXOs(address string, txs []*Transaction) []*UTXO {

	var unUTXOs []*UTXO

	spentTXOutputs := make(map[string][]int)

	for _, tx := range txs {

		//是否创世区块交易
		if tx.IsCoinbaseTransaction() == false {
			for _, in := range tx.Vins {

				publicKeyHash := Base58Decode([]byte(address))
				ripemd160Hash := publicKeyHash[1 : len(publicKeyHash)-4]

				//是否能够解锁
				if in.UnLockWithRipemd160Hash(ripemd160Hash) {

					key := hex.EncodeToString(in.TxHash)

					spentTXOutputs[key] = append(spentTXOutputs[key], in.Vout)
				}
			}
		}
	}

	for _, tx := range txs {

	Work1:
		for index, out := range tx.Vouts {

			if out.UnLockScriptPubKeyWithAddress([]byte(address)) {
				fmt.Println("看看是否是俊诚...")
				fmt.Println(address)

				fmt.Println(spentTXOutputs)

				if len(spentTXOutputs) == 0 {
					utxo := &UTXO{tx.TxHash, index, out}
					unUTXOs = append(unUTXOs, utxo)
				} else {
					for hash, indexArray := range spentTXOutputs {

						txHashStr := hex.EncodeToString(tx.TxHash)

						if hash == txHashStr {

							var isUnSpentUTXO bool

							for _, outIndex := range indexArray {

								if index == outIndex {
									isUnSpentUTXO = true
									continue Work1
								}

								if isUnSpentUTXO == false {
									utxo := &UTXO{tx.TxHash, index, out}
									unUTXOs = append(unUTXOs, utxo)
								}
							}
						} else {
							utxo := &UTXO{tx.TxHash, index, out}
							unUTXOs = append(unUTXOs, utxo)
						}
					}
				}

			}

		}

	}

	blockIterator := blockchain.Iterator()
	for {

		block := blockIterator.Next()

		fmt.Println(block)
		fmt.Println()

		for i := len(block.Txs) - 1; i >= 0; i-- {

			tx := block.Txs[i]

			if tx.IsCoinbaseTransaction() == false {
				for _, in := range tx.Vins {

					publicKeyHash := Base58Decode([]byte(address))
					ripemd160Hash := publicKeyHash[1 : len(publicKeyHash)-4]
					//是否能够解锁
					if in.UnLockWithRipemd160Hash(ripemd160Hash) {

						key := hex.EncodeToString(in.TxHash)

						spentTXOutputs[key] = append(spentTXOutputs[key], in.Vout)
					}

				}
			}

		work:
			for index, out := range tx.Vouts {

				if out.UnLockScriptPubKeyWithAddress([]byte(address)) {

					fmt.Println(out)
					fmt.Println(spentTXOutputs)

					if spentTXOutputs != nil {

						if len(spentTXOutputs) != 0 {

							var isSpentUTXO bool

							for txHash, indexArray := range spentTXOutputs {

								for _, i := range indexArray {
									if index == i && txHash == hex.EncodeToString(tx.TxHash) {
										isSpentUTXO = true
										continue work
									}
								}
							}

							if isSpentUTXO == false {

								utxo := &UTXO{tx.TxHash, index, out}
								unUTXOs = append(unUTXOs, utxo)

							}
						} else {
							utxo := &UTXO{tx.TxHash, index, out}
							unUTXOs = append(unUTXOs, utxo)
						}
					}
				}
			}
		}

		fmt.Println(spentTXOutputs)

		var hashInt big.Int
		hashInt.SetBytes(block.PrevBlockHash)

		if hashInt.Cmp(big.NewInt(0)) == 0 {
			break
		}
	}

	return unUTXOs
}

//转账时查找可用的UTXO
func (blockchain *Blockchain) FindSpendableUTXOS(from string, amount int, txs []*Transaction) (int64, map[string][]int) {

	//获取所有的UTXO
	utxos := blockchain.UnUTXOs(from, txs)

	spendableUTXO := make(map[string][]int)

	//遍历utxos
	var value int64
	for _, utxo := range utxos {

		value = value + utxo.Output.Value

		hash := hex.EncodeToString(utxo.TxHash)
		spendableUTXO[hash] = append(spendableUTXO[hash], utxo.Index)

		if value >= int64(amount) {
			break
		}
	}

	if value < int64(amount) {

		fmt.Printf("%s's balance 不足\n", from)
		os.Exit(1)
	}

	return value, spendableUTXO
}

//挖新的区块
func (blockchain *Blockchain) MineNewBlock(from []string, to []string, amount []string) {

	//验证

	var txs []*Transaction

	for index, address := range from {

		value, _ := strconv.Atoi(amount[index])
		//生成单条交易数据
		tx := NewSimpleTransaction(address, to[index], value, blockchain, txs)

		txs = append(txs, tx)
	}

	//通过相关算法 建立Transaction数组
	var block *Block

	err := blockchain.DB.View(func(tx *bolt.Tx) error {

		bucket := tx.Bucket([]byte(blockTableName))
		if bucket != nil {

			lastHash := bucket.Get([]byte("lastHash"))
			lastBlockBytes := bucket.Get(lastHash)
			block = Deserialize(lastBlockBytes)
		}

		return nil
	})

	if err != nil {
		log.Panic(err)
	}

	//进行签名验证
	for _, tx := range txs {

		if blockchain.VerifyTransaction(tx) == false {
			log.Panic("签名验证失败")
		}
	}


	//建立新的区块
	block = NewBlock(txs, block.Height+1, block.Hash)

	//将新区块存储到数据库
	err = blockchain.DB.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(blockTableName))
		if bucket != nil {
			bucket.Put(block.Hash, block.Serialize())
			bucket.Put([]byte("lastHash"), block.Hash)

			blockchain.Tip = block.Hash
		}

		return nil
	})

	if err != nil {
		log.Panic(err)
	}
}

//地址余额
func (blockchain *Blockchain) GetBalance(address string) int64 {

	utxos := blockchain.UnUTXOs(address, []*Transaction{})

	var amount int64

	for _, utxo := range utxos {
		amount = amount + utxo.Output.Value
	}

	return amount
}

func (blockchain *Blockchain) SignTransaction(transaction *Transaction, privateKey ecdsa.PrivateKey) {

	if transaction.IsCoinbaseTransaction() {
		return
	}

	prevTXs := make(map[string]Transaction)

	for _, vin := range transaction.Vins {
		prevTX, err := blockchain.FindTransaction(vin.TxHash)

		if err != nil {
			log.Panic(err)
		}

		prevTXs[hex.EncodeToString(prevTX.TxHash)] = prevTX
	}

	transaction.Sign(privateKey, prevTXs)
}

func (blockchain *Blockchain) FindTransaction(ID []byte) (Transaction, error) {

	bci := blockchain.Iterator()

	for {
		block := bci.Next()

		for _, tx := range block.Txs {

			if bytes.Compare(tx.TxHash, ID) == 0 {
				return *tx, nil
			}
		}

		var hashInt big.Int
		hashInt.SetBytes(block.PrevBlockHash)
		if big.NewInt(0).Cmp(&hashInt) == 0 {
			break
		}
	}

	return Transaction{}, nil
}

func (blockchain *Blockchain) VerifyTransaction(tx *Transaction) bool {

	prevTXs := make(map[string]Transaction)

	for _, vin := range tx.Vins {
		prevTX, err := blockchain.FindTransaction(vin.TxHash)
		if err != nil {
			log.Panic(err)
		}

		prevTXs[hex.EncodeToString(prevTX.TxHash)] = prevTX
	}

	return tx.Verify(prevTXs)
}
