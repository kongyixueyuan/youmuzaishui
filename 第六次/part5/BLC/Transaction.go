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
	"time"
)

//
type Transaction struct {
	//交易Hash
	ZQ_TxHash []byte

	//输入
	ZQ_Vins []*TXInput

	//输出
	ZQ_Vouts []*TXOutput
}

//将区块序列化为字节数组
func (transaction *Transaction) ZQ_HashTransaction() {

	var result bytes.Buffer

	encoder := gob.NewEncoder(&result)

	err := encoder.Encode(transaction)
	if err != nil {
		log.Panic(err)
	}

	resultBytes := bytes.Join([][]byte{ZQ_IntToHex(time.Now().Unix()), result.Bytes()}, []byte{})
	hash := sha256.Sum256(resultBytes)
	transaction.ZQ_TxHash = hash[:]
}

//判断区块是否是coinbase
func (transaction *Transaction) ZQ_IsCoinbaseTransaction() bool {

	return len(transaction.ZQ_Vins[0].ZQ_TxHash) == 0 && transaction.ZQ_Vins[0].ZQ_Vout == -1
}

//创世区块的Transaction
func ZQ_NewCoinbaseTransaction(address string) *Transaction {

	//消费
	txInput := &TXInput{[]byte{}, -1, nil, []byte{}}
	//未消费
	txOutput := ZQ_NewTXOutput(10, address)

	txCoinbase := &Transaction{[]byte{}, []*TXInput{txInput}, []*TXOutput{txOutput}}
	//设置Hash值
	txCoinbase.ZQ_HashTransaction()

	return txCoinbase
}

//转账时产生的交易
func ZQ_NewSimpleTransaction(from string, to string, amount int64, utxoSet *UTXOSet,txs []*Transaction) *Transaction {

	//返回 转账用户 余额 未花费交易字典
	money, spendableUTXODic := utxoSet.ZQ_FindSpendableUTXOS(from,amount,txs)

	//from 未花费的 Transaction
	//unSpentTx := ZQ_UnSpentTransactionsWithAddress(from)
	//fmt.Println(unSpentTx)
	//from 余额 以及有效的未花费 Transaction
	wallets, _ := ZQ_NewWallets()
	wallet := wallets.ZQ_WalletsMap[from]

	var txInputs []*TXInput
	var txOutputs []*TXOutput

	for txHash,indexArray := range spendableUTXODic  {

		txHashBytes,_ := hex.DecodeString(txHash)
		for _,index := range indexArray  {

			txInput := &TXInput{txHashBytes,index,nil,wallet.ZQ_PublicKey}
			txInputs = append(txInputs, txInput)
		}
	}

	//转账
	txOutput := ZQ_NewTXOutput(int64(amount), to)
	txOutputs = append(txOutputs, txOutput)

	//找零
	txOutput = ZQ_NewTXOutput(int64(money) - int64(amount), from)
	txOutputs = append(txOutputs, txOutput)

	tx := &Transaction{[]byte{}, txInputs, txOutputs}
	tx.ZQ_HashTransaction()

	//数字签名
	utxoSet.ZQ_Blockchain.ZQ_SignTransaction(tx, wallet.ZQ_PrivateKey, txs)

	return tx
}

func (transaction *Transaction) ZQ_TrimmedCopy() Transaction {

	var inputs []*TXInput
	var outputs []*TXOutput

	for _, vin := range transaction.ZQ_Vins {
		inputs = append(inputs, &TXInput{vin.ZQ_TxHash, vin.ZQ_Vout, nil, nil})
	}

	for _, vout := range transaction.ZQ_Vouts {
		outputs = append(outputs, &TXOutput{vout.ZQ_Value, vout.ZQ_Ripemd160Hash})
	}

	txCopy := Transaction{transaction.ZQ_TxHash, inputs, outputs}

	return txCopy

}

//生成交易hash
func (transaction *Transaction) ZQ_Hash() []byte {

	txCopy := transaction
	txCopy.ZQ_TxHash = []byte{}

	hash := sha256.Sum256(txCopy.ZQ_Serialize())

	return hash[:]
}

//序列化
func (transaction *Transaction) ZQ_Serialize() []byte {

	var encoded bytes.Buffer

	enc := gob.NewEncoder(&encoded)
	err := enc.Encode(transaction)
	if err != nil {
		log.Panic(err)
	}

	return encoded.Bytes()
}

func (transaction *Transaction) Sign(privateKey ecdsa.PrivateKey, prevTXs map[string]Transaction)  {

	if transaction.ZQ_IsCoinbaseTransaction() {
		return
	}

	for _, vin := range transaction.ZQ_Vins {

		if prevTXs[hex.EncodeToString(vin.ZQ_TxHash)].ZQ_TxHash == nil {
			log.Panic("Error: prevTXs transaction is not correct")
		}
	}

	txCopy := transaction.ZQ_TrimmedCopy()

	for inID, vin := range txCopy.ZQ_Vins {
		prevTx := prevTXs[hex.EncodeToString(vin.ZQ_TxHash)]

		txCopy.ZQ_Vins[inID].ZQ_Signature = nil
		txCopy.ZQ_Vins[inID].ZQ_PublicKey = prevTx.ZQ_Vouts[vin.ZQ_Vout].ZQ_Ripemd160Hash
		txCopy.ZQ_TxHash = txCopy.ZQ_Hash()
		txCopy.ZQ_Vins[inID].ZQ_PublicKey = nil

		//签名代码
		r, s, err := ecdsa.Sign(rand.Reader, &privateKey, txCopy.ZQ_TxHash)
		if err != nil {
			log.Panic(err)
		}

		signature := append(r.Bytes(), s.Bytes()...)

		transaction.ZQ_Vins[inID].ZQ_Signature = signature
	}
}

//验证数字签名
func (transaction *Transaction) Verify(txMap map[string]Transaction) bool {

	if transaction.ZQ_IsCoinbaseTransaction() {
		return true
	}

	for _, vin := range transaction.ZQ_Vins {
		if(txMap[hex.EncodeToString(vin.ZQ_TxHash)].ZQ_TxHash == nil){
			log.Panic("Error: Previous transaction is not correct")
		}
	}

	txCopy := transaction.ZQ_TrimmedCopy()

	curve := elliptic.P256()

	for inID, vin := range transaction.ZQ_Vins {

		prevTx := txMap[hex.EncodeToString(vin.ZQ_TxHash)]

		txCopy.ZQ_Vins[inID].ZQ_Signature = nil
		txCopy.ZQ_Vins[inID].ZQ_PublicKey = prevTx.ZQ_Vouts[vin.ZQ_Vout].ZQ_Ripemd160Hash
		txCopy.ZQ_TxHash = txCopy.ZQ_Hash()
		txCopy.ZQ_Vins[inID].ZQ_PublicKey = nil

		r := big.Int{}
		s := big.Int{}
		sigLen := len(vin.ZQ_Signature)
		r.SetBytes(vin.ZQ_Signature[:(sigLen / 2)])
		s.SetBytes(vin.ZQ_Signature[(sigLen / 2):])

		x := big.Int{}
		y := big.Int{}
		keyLen := len(vin.ZQ_PublicKey)
		x.SetBytes(vin.ZQ_PublicKey[:(keyLen / 2)])
		y.SetBytes(vin.ZQ_PublicKey[(keyLen / 2):])

		rawPubKey := ecdsa.PublicKey{curve, &x, &y}
		if ecdsa.Verify(&rawPubKey, txCopy.ZQ_TxHash, &r, &s) == false {
			return false
		}
	}

	return true
}