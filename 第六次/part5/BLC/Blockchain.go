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
	ZQ_Tip []byte
	//数据库
	ZQ_DB *bolt.DB
}

//数据库名称 表名
const dbName = "blockchain.db"
const blockTableName = "blocks"

//创建迭代器
func (blockchain *Blockchain) Iterator() *BlockchainIterator {

	return &BlockchainIterator{blockchain.ZQ_Tip, blockchain.ZQ_DB}
}

//打印区块链
func (blockchain *Blockchain) ZQ_Printchain() {

	blockchainIterator := blockchain.Iterator()

	var hashInt big.Int

	for {
		block := blockchainIterator.ZQ_Next()

		fmt.Printf("Height: %d\n", block.ZQ_Height)
		fmt.Printf("PrevBlockHash: %x\n", block.ZQ_PrevBlockHash)
		fmt.Printf("Timestamp: %s\n", time.Unix(block.ZQ_Timestamp, 0).Format("2006-01-02 03:04:05 PM"))
		fmt.Printf("Hash: %x\n", block.ZQ_Hash)
		fmt.Printf("Nonce: %d\n", block.ZQ_Nonce)
		fmt.Println("Txs:")
		for _, tx := range block.ZQ_Txs {


			fmt.Printf("\tTxHash: %x\n", tx.ZQ_TxHash)
			fmt.Println("\tVins:")
			for _, in := range tx.ZQ_Vins {
				fmt.Printf("\t\tTxHash: %x\n", in.ZQ_TxHash)
				fmt.Printf("\t\tVout: %d\n", in.ZQ_Vout)
				fmt.Printf("\t\tPublicKey: %v\n", in.ZQ_PublicKey)
				fmt.Printf("\t\tSignature: %v\n", in.ZQ_Signature)
			}

			fmt.Println("\tVouts:")
			for _, out := range tx.ZQ_Vouts {
				fmt.Printf("\t\tMoney: %d\n", out.ZQ_Value)
				fmt.Printf("\t\tRipemd160Hash: %v\n", out.ZQ_Ripemd160Hash)
			}

			fmt.Println("--------------")
		}

		fmt.Println("#############")
		fmt.Println()

		hashInt.SetBytes(block.ZQ_PrevBlockHash)

		//如果上一个hash是否是创世区块 break
		if big.NewInt(0).Cmp(&hashInt) == 0 {
			break
		}
	}
}

//增加区块至区块链
func (blockchain *Blockchain) ZQ_AddBlockToBlockchain(txs []*Transaction) {

	//添加区块到数据库
	err := blockchain.ZQ_DB.Update(func(tx *bolt.Tx) error {

		bucket := tx.Bucket([]byte(blockTableName))
		if bucket != nil {

			//读取最新区块
			blockBytes := bucket.Get(blockchain.ZQ_Tip)
			block := Deserialize(blockBytes)

			//创建新区块
			newBlock := NewBlock(txs, block.ZQ_Height+1, block.ZQ_Hash)

			err := bucket.Put(newBlock.ZQ_Hash, newBlock.Serialize())
			if err != nil {
				log.Panic(err)
			}

			err = bucket.Put([]byte("lastHash"), newBlock.ZQ_Hash)
			if err != nil {
				log.Panic(err)
			}

			blockchain.ZQ_Tip = newBlock.ZQ_Hash
		}
		return nil
	})

	if err != nil {
		log.Panic(err)
	}
}

//数据库是否存在
func ZQ_DBExists() bool {
	if _, err := os.Stat(dbName); os.IsNotExist(err) {
		return false
	}

	return true
}

//创建带有创世区块的区块链
func ZQ_CreateBlockchainWithGenesisBlock(address string) *Blockchain {

	//数字库是否存在
	if ZQ_DBExists() {
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
			txCoinbase := ZQ_NewCoinbaseTransaction(address)

			//创建创世区块
			genesisBlock := CreateGenesisBlock(txCoinbase)

			//存入数据 hash => 序列化区块
			err = bucket.Put(genesisBlock.ZQ_Hash, genesisBlock.Serialize())
			if err != nil {
				log.Panic(err)
			}

			//存储最新的hash
			err = bucket.Put([]byte("lastHash"), genesisBlock.ZQ_Hash)
			if err != nil {
				log.Panic(err)
			}

			genesisHash = genesisBlock.ZQ_Hash
		}

		return nil
	})

	if err != nil {
		log.Panic(err)
	}

	return &Blockchain{genesisHash, db}
}

//获取区块链对象
func ZQ_GetBlockchainObject() *Blockchain {

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
func (blockchain *Blockchain) ZQ_UnSpentTransactionsWithAddress(address string) []*TXOutput {

	var unUTXOs []*TXOutput

	//address下所对应的vins
	spentTXOutputs := make(map[string][]int)

	//遍历数据库
	iterator := blockchain.Iterator()
	var hasInt big.Int

	for {
		block := iterator.ZQ_Next()
		fmt.Println(block)
		fmt.Println()

		hasInt.SetBytes(block.ZQ_PrevBlockHash)
		if hasInt.Cmp(big.NewInt(0)) == 0 {
			break
		}

		for _, tx := range block.ZQ_Txs {

			if tx.ZQ_IsCoinbaseTransaction() == false {

				for _, in := range tx.ZQ_Vins {

					publicKeyHash := ZQ_Base58Decode([]byte(address))
					ripemd160Hash := publicKeyHash[1 : len(publicKeyHash)-4]

					//是否能够解锁
					if in.ZQ_UnLockWithRipemd160Hash(ripemd160Hash) {

						key := hex.EncodeToString(in.ZQ_TxHash)
						spentTXOutputs[key] = append(spentTXOutputs[key], in.ZQ_Vout)
					}
				}
			}

			for index, out := range tx.ZQ_Vouts {

				//是否能够解锁
				if out.ZQ_UnLockScriptPubKeyWithAddress(address) {

					if spentTXOutputs != nil {

						for txHash, indexArray := range spentTXOutputs {

							if txHash == hex.EncodeToString(tx.ZQ_TxHash) {

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
func (blockchain *Blockchain) ZQ_UnUTXOs(address string, txs []*Transaction) []*UTXO {

	var unUTXOs []*UTXO

	spentTXOutputs := make(map[string][]int)

	for _, tx := range txs {

		//是否创世区块交易
		if tx.ZQ_IsCoinbaseTransaction() == false {
			for _, in := range tx.ZQ_Vins {

				publicKeyHash := ZQ_Base58Decode([]byte(address))
				ripemd160Hash := publicKeyHash[1 : len(publicKeyHash)-4]

				//是否能够解锁
				if in.ZQ_UnLockWithRipemd160Hash(ripemd160Hash) {

					key := hex.EncodeToString(in.ZQ_TxHash)

					spentTXOutputs[key] = append(spentTXOutputs[key], in.ZQ_Vout)
				}
			}
		}
	}

	for _, tx := range txs {

	Work1:
		for index, out := range tx.ZQ_Vouts {

			if out.ZQ_UnLockScriptPubKeyWithAddress(address) {
				fmt.Println(address)

				fmt.Println(spentTXOutputs)

				if len(spentTXOutputs) == 0 {
					utxo := &UTXO{tx.ZQ_TxHash, index, out}
					unUTXOs = append(unUTXOs, utxo)
				} else {
					for hash, indexArray := range spentTXOutputs {

						txHashStr := hex.EncodeToString(tx.ZQ_TxHash)

						if hash == txHashStr {

							var isUnSpentUTXO bool

							for _, outIndex := range indexArray {

								if index == outIndex {
									isUnSpentUTXO = true
									continue Work1
								}

								if isUnSpentUTXO == false {
									utxo := &UTXO{tx.ZQ_TxHash, index, out}
									unUTXOs = append(unUTXOs, utxo)
								}
							}
						} else {
							utxo := &UTXO{tx.ZQ_TxHash, index, out}
							unUTXOs = append(unUTXOs, utxo)
						}
					}
				}

			}

		}

	}

	blockIterator := blockchain.Iterator()
	for {

		block := blockIterator.ZQ_Next()

		fmt.Println(block)
		fmt.Println()

		for i := len(block.ZQ_Txs) - 1; i >= 0; i-- {

			tx := block.ZQ_Txs[i]

			if tx.ZQ_IsCoinbaseTransaction() == false {
				for _, in := range tx.ZQ_Vins {

					publicKeyHash := ZQ_Base58Decode([]byte(address))
					ripemd160Hash := publicKeyHash[1 : len(publicKeyHash)-4]
					//是否能够解锁
					if in.ZQ_UnLockWithRipemd160Hash(ripemd160Hash) {

						key := hex.EncodeToString(in.ZQ_TxHash)

						spentTXOutputs[key] = append(spentTXOutputs[key], in.ZQ_Vout)
					}

				}
			}

		work:
			for index, out := range tx.ZQ_Vouts {

				if out.ZQ_UnLockScriptPubKeyWithAddress(address) {

					fmt.Println(out)
					fmt.Println(spentTXOutputs)

					if spentTXOutputs != nil {

						if len(spentTXOutputs) != 0 {

							var isSpentUTXO bool

							for txHash, indexArray := range spentTXOutputs {

								for _, i := range indexArray {
									if index == i && txHash == hex.EncodeToString(tx.ZQ_TxHash) {
										isSpentUTXO = true
										continue work
									}
								}
							}

							if isSpentUTXO == false {

								utxo := &UTXO{tx.ZQ_TxHash, index, out}
								unUTXOs = append(unUTXOs, utxo)

							}
						} else {
							utxo := &UTXO{tx.ZQ_TxHash, index, out}
							unUTXOs = append(unUTXOs, utxo)
						}
					}
				}
			}
		}

		fmt.Println(spentTXOutputs)

		var hashInt big.Int
		hashInt.SetBytes(block.ZQ_PrevBlockHash)

		if hashInt.Cmp(big.NewInt(0)) == 0 {
			break
		}
	}

	return unUTXOs
}

//转账时查找可用的UTXO
func (blockchain *Blockchain) ZQ_FindSpendableUTXOS(from string, amount int, txs []*Transaction) (int64, map[string][]int) {

	//获取所有的UTXO
	utxos := blockchain.ZQ_UnUTXOs(from, txs)

	spendableUTXO := make(map[string][]int)

	//遍历utxos
	var value int64
	for _, utxo := range utxos {

		value = value + utxo.ZQ_Output.ZQ_Value

		hash := hex.EncodeToString(utxo.ZQ_TxHash)
		spendableUTXO[hash] = append(spendableUTXO[hash], utxo.ZQ_Index)

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
func (blockchain *Blockchain) ZQ_MineNewBlock(from []string, to []string, amount []string) {

	utxoSet := &UTXOSet{blockchain}

	var txs []*Transaction

	for index, address := range from {

		value, _ := strconv.Atoi(amount[index])
		//生成单条交易数据
		tx := ZQ_NewSimpleTransaction(address, to[index], int64(value), utxoSet, txs)

		txs = append(txs, tx)
	}

	//区块奖励
	tx := ZQ_NewCoinbaseTransaction(from[0])
	txs = append(txs, tx)

	//通过相关算法 建立Transaction数组
	var block *Block

	err := blockchain.ZQ_DB.View(func(tx *bolt.Tx) error {

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

	_txs := []*Transaction{}

	for _, tx := range txs {

		if blockchain.ZQ_VerifyTransaction(tx, _txs) == false {
			log.Panic("签名验证失败")
		}

		_txs = append(_txs, tx)
	}


	//建立新的区块
	block = NewBlock(txs, block.ZQ_Height+1, block.ZQ_Hash)

	//将新区块存储到数据库
	err = blockchain.ZQ_DB.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(blockTableName))
		if bucket != nil {
			bucket.Put(block.ZQ_Hash, block.Serialize())
			bucket.Put([]byte("lastHash"), block.ZQ_Hash)

			blockchain.ZQ_Tip = block.ZQ_Hash
		}

		return nil
	})

	if err != nil {
		log.Panic(err)
	}
}

//地址余额
func (blockchain *Blockchain) ZQ_GetBalance(address string) int64 {

	utxos := blockchain.ZQ_UnUTXOs(address, []*Transaction{})

	var amount int64

	for _, utxo := range utxos {
		amount = amount + utxo.ZQ_Output.ZQ_Value
	}

	return amount
}

func (blockchain *Blockchain) ZQ_SignTransaction(transaction *Transaction, privateKey ecdsa.PrivateKey, txs []*Transaction) {

	if transaction.ZQ_IsCoinbaseTransaction() {
		return
	}

	prevTXs := make(map[string]Transaction)

	for _, vin := range transaction.ZQ_Vins {
		prevTX, err := blockchain.ZQ_FindTransaction(vin.ZQ_TxHash, txs)

		if err != nil {
			log.Panic(err)
		}

		prevTXs[hex.EncodeToString(prevTX.ZQ_TxHash)] = prevTX
	}

	transaction.Sign(privateKey, prevTXs)
}

func (blockchain *Blockchain) ZQ_FindTransaction(ID []byte, txs []*Transaction) (Transaction, error) {

	for _, tx := range txs {
		if bytes.Compare(tx.ZQ_TxHash, ID) == 0 {
			return *tx, nil
		}
	}

	bci := blockchain.Iterator()

	for {
		block := bci.ZQ_Next()

		for _, tx := range block.ZQ_Txs {

			if bytes.Compare(tx.ZQ_TxHash, ID) == 0 {
				return *tx, nil
			}
		}

		var hashInt big.Int
		hashInt.SetBytes(block.ZQ_PrevBlockHash)
		if big.NewInt(0).Cmp(&hashInt) == 0 {
			break
		}
	}

	return Transaction{}, nil
}

func (blockchain *Blockchain) ZQ_VerifyTransaction(tx *Transaction, txs []*Transaction) bool {

	prevTXs := make(map[string]Transaction)

	for _, vin := range tx.ZQ_Vins {
		prevTX, err := blockchain.ZQ_FindTransaction(vin.ZQ_TxHash, txs)
		if err != nil {
			log.Panic(err)
		}

		prevTXs[hex.EncodeToString(prevTX.ZQ_TxHash)] = prevTX
	}

	return tx.Verify(prevTXs)
}


//返回map[string]*TXOutputs
func (blockchain *Blockchain) ZQ_FindUTXOMap() map[string]*TXOutputs {

	//存储已花费UTXO信息
	var spentableUTXOSMAP = make(map[string][]*TXInput)
	//存储有效的未花费
	var utxoMaps = make(map[string]*TXOutputs)

	iterator := blockchain.Iterator()

	for {
		block := iterator.ZQ_Next()

		for i := len(block.ZQ_Txs) - 1; i >= 0; i-- {

			txOutputs := &TXOutputs{[]*UTXO{}}

			tx := block.ZQ_Txs[i]
			txHash := hex.EncodeToString(tx.ZQ_TxHash)

			if tx.ZQ_IsCoinbaseTransaction() == false {

				for _, inInput := range tx.ZQ_Vins{
					txInputHash := hex.EncodeToString(inInput.ZQ_TxHash)
					spentableUTXOSMAP[txInputHash] = append(spentableUTXOSMAP[txInputHash], inInput)
				}
			}

			WorkOutLoop:
			for index,out := range tx.ZQ_Vouts {

				isSpent := false
				txInputs := spentableUTXOSMAP[txHash]

				if len(txInputs) > 0 {

					for _, in := range txInputs {

						//if index == in.ZQ_Vout {

							outPublicKey := out.ZQ_Ripemd160Hash
							inPublicKey := in.ZQ_PublicKey

							if bytes.Compare(outPublicKey, ZQ_Ripemd160Hash(inPublicKey)) == 0{

								if index == in.ZQ_Vout {
									isSpent = true
									continue WorkOutLoop
								}
							}
						//}
					}
				}

				if isSpent == false {
					fmt.Printf("set: %d\n", out.ZQ_Value)
					utxo := &UTXO{tx.ZQ_TxHash, index, out}
					txOutputs.ZQ_UTXOS = append(txOutputs.ZQ_UTXOS, utxo)
				}
			}

			utxoMaps[txHash] = txOutputs
		}

		var hashInt big.Int
		hashInt.SetBytes(block.ZQ_PrevBlockHash)

		if hashInt.Cmp(big.NewInt(0)) == 0{
			break
		}
	}

	return utxoMaps
}
