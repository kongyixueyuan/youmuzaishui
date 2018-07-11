package BLC

import (
	"bytes"
)

//
type TXOutput struct {
	ZQ_Value int64

	ZQ_Ripemd160Hash []byte
	//ScriptPubKey string
}

func (txOutput *TXOutput) ZQ_Lock(address string)  {

	publicKeyHash := ZQ_Base58Decode([]byte(address))
	txOutput.ZQ_Ripemd160Hash = publicKeyHash[1:len(publicKeyHash) - 4]
}

//判断当前input是否是某个用户
func (txOutput *TXOutput) ZQ_UnLockScriptPubKeyWithAddress(address string) bool {

	publicKeyHash := ZQ_Base58Decode([]byte(address))
	ripemd160Hash := publicKeyHash[1:len(publicKeyHash) - 4]

	return bytes.Compare(txOutput.ZQ_Ripemd160Hash, ripemd160Hash) == 0
}

func ZQ_NewTXOutput(value int64, address string) *TXOutput {

	txOutput := &TXOutput{value, nil}

	//设置Ripemd160Hash
	txOutput.ZQ_Lock(address)

	return txOutput
}