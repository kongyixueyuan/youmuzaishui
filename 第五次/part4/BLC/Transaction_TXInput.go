package BLC

import "bytes"

type TXInput struct {
	//交易的Hash
	TxHash []byte
	//未花费的
	Vout int
	//数字签名
	Signature []byte
	//公钥 钱包公钥
	PublicKey []byte
	//ScriptSig string
}

//判断当前input是否是某个用户
func (txInput *TXInput) UnLockWithRipemd160Hash(ripemd160Hash []byte) bool {

	publicKey := Ripemd160Hash(txInput.PublicKey)

	return bytes.Compare(publicKey, ripemd160Hash) == 0
}