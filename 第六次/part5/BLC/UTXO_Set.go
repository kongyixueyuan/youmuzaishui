package BLC

import (
	"github.com/boltdb/bolt"
	"log"
	"encoding/hex"
	"fmt"
	"bytes"
	"publicChain/part4/BLC"
)

//存储所有未发费的交易输出
type UTXOSet struct {
	ZQ_Blockchain *Blockchain
}

const utxoTableName = "utxo"

//遍历整个数据库 读取所有的未花费的UTXO 存入数据库
//重置
func (utxoSet *UTXOSet) ZQ_ResetUTXOSet() {

	err := utxoSet.ZQ_Blockchain.ZQ_DB.Update(func(tx *bolt.Tx) error {

		bucket := tx.Bucket([]byte(utxoTableName))
		if bucket != nil {

			err := tx.DeleteBucket([]byte(utxoTableName))
			if err != nil {
				log.Panic(err)
			}
		}

		bucket, err := tx.CreateBucket([]byte(utxoTableName))
		if err != nil {
			log.Panic(err)
		}

		if bucket != nil {
			txOutputsMap := utxoSet.ZQ_Blockchain.ZQ_FindUTXOMap()
			for txHash, txOutputs := range txOutputsMap {

				fmt.Printf("写入交易数据：%v\n", txHash)
				txHashBytes, _ := hex.DecodeString(txHash)
				bucket.Put(txHashBytes, txOutputs.Serialize())
			}
		}

		return nil
	})

	if err != nil {
		log.Panic(err)
	}
}

//更新UTXO变化到数据库

//查找某地址的UTXO
func (utxoSet *UTXOSet) zQ_FindUTXOForAddress(address string) []*UTXO {

	var utxos []*UTXO

	err := utxoSet.ZQ_Blockchain.ZQ_DB.View(func(tx *bolt.Tx) error {

		bucket := tx.Bucket([]byte(utxoTableName))
		cursor := bucket.Cursor()

		for k, v := cursor.First(); k != nil; k, v = cursor.Next() {

			//fmt.Printf("key=%v, value=%v\n", k, v)

			txOutputs := DeserializeTXOutputs(v)
			fmt.Println(len(txOutputs.ZQ_UTXOS))

			for _, utxo := range txOutputs.ZQ_UTXOS {

				if utxo.ZQ_Output.ZQ_UnLockScriptPubKeyWithAddress(address) {

					//fmt.Println(utxo.ZQ_Output.ZQ_Value)
					utxos = append(utxos, utxo)
				}
			}
		}

		return nil
	})

	if err != nil {
		log.Panic(err)
	}

	return utxos
}

//查询余额
func (utxoSet *UTXOSet) ZQ_GetBalance(address string) int64 {

	UTXOS := utxoSet.zQ_FindUTXOForAddress(address)

	var amount int64
	for _, utxo := range UTXOS {

		amount += utxo.ZQ_Output.ZQ_Value
	}

	return amount
}

func (utxoSet *UTXOSet) ZQ_FindUnPackageSpendableUTXOS(from string, txs []*Transaction) []*UTXO {

	var unUTXOS []*UTXO


	spendableUTXO := make(map[string][]int)

	//处理未打包交易
	for _, tx := range txs {

		if tx.ZQ_IsCoinbaseTransaction() == false {

			for _, in := range tx.ZQ_Vins {

				publicKeyHash := ZQ_Base58Decode([]byte(from))
				ripemd160Hash := publicKeyHash[1 : len(publicKeyHash)-4]

				//是否能够解锁
				if in.ZQ_UnLockWithRipemd160Hash(ripemd160Hash) {

					key := hex.EncodeToString(in.ZQ_TxHash)

					spendableUTXO[key] = append(spendableUTXO[key], in.ZQ_Vout)
				}
			}
		}
	}

	for _, tx := range txs {

	Work1:
		for index, out := range tx.ZQ_Vouts {

			if out.ZQ_UnLockScriptPubKeyWithAddress(from) {
				fmt.Println(from)

				fmt.Println(spendableUTXO)

				if len(spendableUTXO) == 0 {
					utxo := &UTXO{tx.ZQ_TxHash, index, out}
					unUTXOS = append(unUTXOS, utxo)
				} else {
					for hash, indexArray := range spendableUTXO {

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
									unUTXOS = append(unUTXOS, utxo)
								}
							}
						} else {
							utxo := &UTXO{tx.ZQ_TxHash, index, out}
							unUTXOS = append(unUTXOS, utxo)
						}
					}
				}

			}

		}

	}

	return unUTXOS
}

func (utxoSet *UTXOSet) ZQ_FindSpendableUTXOS(from string, amount int64, txs []*Transaction) (int64, map[string][]int) {

	unPackageUTXOS := utxoSet.ZQ_FindUnPackageSpendableUTXOS(from, txs)

	spentableUTXO := make(map[string][]int)
	var money int64 = 0

	for _, utxo := range unPackageUTXOS {

		money += utxo.ZQ_Output.ZQ_Value
		txHash := hex.EncodeToString(utxo.ZQ_TxHash)
		spentableUTXO[txHash] = append(spentableUTXO[txHash], utxo.ZQ_Index)

		if money >= amount {

			return money, spentableUTXO
		}
	}

	err := utxoSet.ZQ_Blockchain.ZQ_DB.View(func(tx *bolt.Tx) error {

		bucket := tx.Bucket([]byte(utxoTableName))

		if bucket != nil {

			cursor := bucket.Cursor()

			UTXOBREAK:
			for k, v := cursor.First(); k != nil; k, v = cursor.Next() {

				fmt.Println("---------------调试----------------")
				fmt.Printf("交易数据Hash %v\n", k)
				txOutputs := DeserializeTXOutputs(v)

				for _, utxo := range txOutputs.ZQ_UTXOS {

					//增加解锁 解决bug
					if utxo.ZQ_Output.ZQ_UnLockScriptPubKeyWithAddress(from){
						money += utxo.ZQ_Output.ZQ_Value

						txHash := hex.EncodeToString(utxo.ZQ_TxHash)
						spentableUTXO[txHash] = append(spentableUTXO[txHash], utxo.ZQ_Index)

						if money >= amount {
							break UTXOBREAK
						}
					}
				}
			}
		}

		return nil
	})

	if err != nil {
		log.Panic(err)
	}

	if money < amount {
		log.Panic("余额不足")
	}

	return money, spentableUTXO
}

//取出最新的block 循环交易数据 更新utxo
func (utxoSet *UTXOSet) Update()  {

	block := utxoSet.ZQ_Blockchain.Iterator().ZQ_Next()

	//所有已花费
	ins := []*TXInput{}

	outsMap := make(map[string]*TXOutputs)

	for _, tx := range block.ZQ_Txs {

		for _, in := range tx.ZQ_Vins {
			ins = append(ins, in)
		}
	}

	for _, tx := range block.ZQ_Txs {

		utxos := []*UTXO{}
		for index, out := range tx.ZQ_Vouts {

			isSpent := false
			for _, in := range ins {

				if in.ZQ_Vout == index  && bytes.Compare(tx.ZQ_TxHash, in.ZQ_TxHash) == 0 && bytes.Compare(out.ZQ_Ripemd160Hash, BLC.Ripemd160Hash(in.ZQ_PublicKey)) == 0{
					isSpent = true
					continue
				}
			}

			if isSpent == false {

				utxo := &UTXO{tx.ZQ_TxHash, index, out}
				utxos = append(utxos, utxo)
			}
		}

		if len(utxos) > 0 {
			txHash := hex.EncodeToString(tx.ZQ_TxHash)
			outsMap[txHash] = &TXOutputs{utxos}
		}
	}

	err := utxoSet.ZQ_Blockchain.ZQ_DB.Update(func(tx *bolt.Tx) error {

		bucket := tx.Bucket([]byte(utxoTableName))

		if bucket != nil {

			for _, in := range ins {

				txOutputsBytes := bucket.Get(in.ZQ_TxHash)
				if txOutputsBytes == nil {
					continue
				}

				txOutputs := DeserializeTXOutputs(txOutputsBytes)

				UTXOS := []*UTXO{}

				isNeedDelete := false

				for _, utxo := range txOutputs.ZQ_UTXOS {

					if in.ZQ_Vout == utxo.ZQ_Index && bytes.Compare(utxo.ZQ_Output.ZQ_Ripemd160Hash, BLC.Ripemd160Hash(in.ZQ_PublicKey)) == 0{

						isNeedDelete = true
					} else {
						UTXOS = append(UTXOS, utxo)
					}
				}

				if isNeedDelete == true {

					bucket.Delete(in.ZQ_TxHash)

					if len(UTXOS) > 0 {

						txHash := hex.EncodeToString(in.ZQ_TxHash)

						preTXoutputs := outsMap[txHash]
						if preTXoutputs == nil {

							outsMap[txHash] = &TXOutputs{UTXOS}
						} else {

							preTXoutputs.ZQ_UTXOS = append(preTXoutputs.ZQ_UTXOS, UTXOS...)
							outsMap[txHash] = preTXoutputs
						}
					}
				}

			}

			//遍历区块utxo 新增
			for txHash, outPuts := range outsMap {

				txHashBytes, _ := hex.DecodeString(txHash)
				bucket.Put(txHashBytes, outPuts.Serialize())
			}
		}

		return nil
	})

	if err != nil {
		log.Panic(err)
	}
}