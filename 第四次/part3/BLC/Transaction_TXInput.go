package BLC

type TXInput struct {
	//交易的Hash
	TxHash []byte
	//未花费的
	Vout int
	//数字签名
	ScriptSig string
}

//判断当前input是否是某个用户
func (txInput *TXInput) UnLockWithAddress(address string) bool {

	return txInput.ScriptSig == address
}