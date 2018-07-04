package BLC

//
type TXOutput struct {
	Value int64

	ScriptPubKey string
}

//判断当前input是否是某个用户
func (txOutput *TXOutput) UnLockScriptPubKeyWithAddress(address string) bool {

	return txOutput.ScriptPubKey == address
}