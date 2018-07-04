package BLC

import (
	"crypto/sha256"
	"bytes"
	"encoding/gob"
	"log"
	"encoding/hex"
)

//
type Transaction struct {
	//交易Hash
	TxHash []byte

	//输入
	Vins []*TXInput

	//输出
	Vouts []*TXOutput
}

//将区块序列化为字节数组
func (transaction *Transaction) HashTransaction() {

	var result bytes.Buffer

	encoder := gob.NewEncoder(&result)

	err := encoder.Encode(transaction)
	if err != nil {
		log.Panic(err)
	}

	hash := sha256.Sum256(result.Bytes())
	transaction.TxHash = hash[:]
}

//判断区块是否是coinbase
func (transaction *Transaction) IsCoinbaseTransaction() bool {

	return len(transaction.Vins[0].TxHash) == 0 && transaction.Vins[0].Vout == -1
}

//创世区块的Transaction
func NewCoinbaseTransaction(address string) *Transaction {

	//消费
	txInput := &TXInput{[]byte{}, -1, "Genesis data..."}
	//未消费
	txOutput := &TXOutput{10, address}

	txCoinbase := &Transaction{[]byte{}, []*TXInput{txInput}, []*TXOutput{txOutput}}
	//设置Hash值
	txCoinbase.HashTransaction()

	return txCoinbase
}

//转账时产生的交易
func NewSimpleTransaction(from string, to string, amount int, blockchain *Blockchain,txs []*Transaction) *Transaction {

	//返回 转账用户 余额 未花费交易字典
	money, spendableUTXODic := blockchain.FindSpendableUTXOS(from,amount,txs)

	//from 未花费的 Transaction
	//unSpentTx := UnSpentTransactionsWithAddress(from)
	//fmt.Println(unSpentTx)
	//from 余额 以及有效的未花费 Transaction

	var txInputs []*TXInput
	var txOutputs []*TXOutput

	for txHash,indexArray := range spendableUTXODic  {

		txHashBytes,_ := hex.DecodeString(txHash)
		for _,index := range indexArray  {

			txInput := &TXInput{txHashBytes,index,from}
			txInputs = append(txInputs, txInput)
		}
	}

	//转账
	txOutput := &TXOutput{int64(amount), to}
	txOutputs = append(txOutputs, txOutput)

	//找零
	txOutput = &TXOutput{int64(money) - int64(amount), from}
	txOutputs = append(txOutputs, txOutput)

	tx := &Transaction{[]byte{}, txInputs, txOutputs}
	tx.HashTransaction()

	return tx
}
