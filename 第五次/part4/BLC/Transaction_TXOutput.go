package BLC

import (
	"bytes"
)

//
type TXOutput struct {
	Value int64

	Ripemd160Hash []byte
	//ScriptPubKey string
}

func (txOutput *TXOutput) Lock(address string)  {

	publicKeyHash := Base58Decode([]byte(address))
	txOutput.Ripemd160Hash = publicKeyHash[1:len(publicKeyHash) - 4]
}

//判断当前input是否是某个用户
func (txOutput *TXOutput) UnLockScriptPubKeyWithAddress(Ripemd160Hash []byte) bool {

	publicKeyHash := Base58Decode(Ripemd160Hash)
	ripemd160Hash := publicKeyHash[1:len(publicKeyHash) - 4]

	return bytes.Compare(txOutput.Ripemd160Hash, ripemd160Hash) == 0
}

func NewTXOutput(value int64, address string) *TXOutput {

	txOutput := &TXOutput{value, nil}

	//设置Ripemd160Hash
	txOutput.Lock(address)

	return txOutput
}