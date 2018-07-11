package BLC

import "bytes"

type TXInput struct {
	//交易的Hash
	ZQ_TxHash []byte
	//未花费的
	ZQ_Vout int
	//数字签名
	ZQ_Signature []byte
	//公钥 钱包公钥
	ZQ_PublicKey []byte
	//ScriptSig string
}

//判断当前input是否是某个用户
func (txInput *TXInput) ZQ_UnLockWithRipemd160Hash(ripemd160Hash []byte) bool {

	publicKey := ZQ_Ripemd160Hash(txInput.ZQ_PublicKey)

	return bytes.Compare(publicKey, ripemd160Hash) == 0
}