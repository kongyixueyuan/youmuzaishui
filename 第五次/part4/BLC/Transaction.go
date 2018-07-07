package BLC

import (
	"crypto/sha256"
	"bytes"
	"encoding/gob"
	"log"
	"encoding/hex"
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/elliptic"
	"math/big"
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
	txInput := &TXInput{[]byte{}, -1, nil, []byte{}}
	//未消费
	txOutput := NewTXOutput(10, address)

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
	wallets, _ := NewWallets();
	wallet := wallets.WalletsMap[from]

	var txInputs []*TXInput
	var txOutputs []*TXOutput

	for txHash,indexArray := range spendableUTXODic  {

		txHashBytes,_ := hex.DecodeString(txHash)
		for _,index := range indexArray  {

			txInput := &TXInput{txHashBytes,index,nil,wallet.PublicKey}
			txInputs = append(txInputs, txInput)
		}
	}

	//转账
	txOutput := NewTXOutput(int64(amount), to)
	txOutputs = append(txOutputs, txOutput)

	//找零
	txOutput = NewTXOutput(int64(money) - int64(amount), from)
	txOutputs = append(txOutputs, txOutput)

	tx := &Transaction{[]byte{}, txInputs, txOutputs}
	tx.HashTransaction()

	//数字签名
	blockchain.SignTransaction(tx, wallet.PrivateKey)

	return tx
}

func (transaction *Transaction) TrimmedCopy() Transaction {

	var inputs []*TXInput
	var outputs []*TXOutput

	for _, vin := range transaction.Vins {
		inputs = append(inputs, &TXInput{vin.TxHash, vin.Vout, nil, nil})
	}

	for _, vout := range transaction.Vouts {
		outputs = append(outputs, &TXOutput{vout.Value, vout.Ripemd160Hash})
	}

	txCopy := Transaction{transaction.TxHash, inputs, outputs}

	return txCopy

}

//生成交易hash
func (transaction *Transaction)  Hash() []byte {

	txCopy := transaction
	txCopy.TxHash = []byte{}

	hash := sha256.Sum256(txCopy.Serialize())

	return hash[:]
}

//序列化
func (transaction *Transaction) Serialize() []byte {

	var encoded bytes.Buffer

	enc := gob.NewEncoder(&encoded)
	err := enc.Encode(transaction)
	if err != nil {
		log.Panic(err)
	}

	return encoded.Bytes()
}

func (transaction *Transaction) Sign(privateKey ecdsa.PrivateKey, prevTXs map[string]Transaction)  {

	if transaction.IsCoinbaseTransaction() {
		return
	}

	for _, vin := range transaction.Vins {

		if prevTXs[hex.EncodeToString(vin.TxHash)].TxHash == nil {
			log.Panic("Error: prevTXs transaction is not correct")
		}
	}

	txCopy := transaction.TrimmedCopy()

	for inID, vin := range txCopy.Vins {
		prevTx := prevTXs[hex.EncodeToString(vin.TxHash)]

		txCopy.Vins[inID].Signature = nil
		txCopy.Vins[inID].PublicKey = prevTx.Vouts[vin.Vout].Ripemd160Hash
		txCopy.TxHash = txCopy.Hash()
		txCopy.Vins[inID].PublicKey = nil

		//签名代码
		r, s, err := ecdsa.Sign(rand.Reader, &privateKey, txCopy.TxHash)
		if err != nil {
			log.Panic(err)
		}

		signature := append(r.Bytes(), s.Bytes()...)

		transaction.Vins[inID].Signature = signature
	}
}

//验证数字签名
func (transaction *Transaction) Verify(txMap map[string]Transaction) bool {

	if transaction.IsCoinbaseTransaction() {
		return true
	}

	for _, vin := range transaction.Vins {
		if(txMap[hex.EncodeToString(vin.TxHash)].TxHash == nil){
			log.Panic("Error: Previous transaction is not correct")
		}
	}

	txCopy := transaction.TrimmedCopy()

	curve := elliptic.P256()

	for inID, vin := range transaction.Vins {

		prevTx := txMap[hex.EncodeToString(vin.TxHash)]

		txCopy.Vins[inID].Signature = nil
		txCopy.Vins[inID].PublicKey = prevTx.Vouts[vin.Vout].Ripemd160Hash
		txCopy.TxHash = txCopy.Hash()
		txCopy.Vins[inID].PublicKey = nil

		r := big.Int{}
		s := big.Int{}
		sigLen := len(vin.Signature)
		r.SetBytes(vin.Signature[:(sigLen / 2)])
		s.SetBytes(vin.Signature[(sigLen / 2):])

		x := big.Int{}
		y := big.Int{}
		keyLen := len(vin.PublicKey)
		x.SetBytes(vin.PublicKey[:(keyLen / 2)])
		y.SetBytes(vin.PublicKey[(keyLen / 2):])

		rawPubKey := ecdsa.PublicKey{curve, &x, &y}
		if ecdsa.Verify(&rawPubKey, txCopy.TxHash, &r, &s) == false {
			return false
		}
	}

	return true
}