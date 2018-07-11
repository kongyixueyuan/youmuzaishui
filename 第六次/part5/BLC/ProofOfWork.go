package BLC

import (
	"math/big"
	"bytes"
	"crypto/sha256"
	"fmt"
)

type ProofOfWork struct {
	//当前要验证的区块
	ZQ_Block *Block
	//大数据存储 难度
	ZQ_target *big.Int
}

const targetBit = 16

//判断hash有效性
func (proofOfWork *ProofOfWork) ZQ_IsValid() bool {
	//hash 与 ZQ_target 进行比较
	var hashInt big.Int
	hashInt.SetBytes(proofOfWork.ZQ_Block.ZQ_Hash)
	if proofOfWork.ZQ_target.Cmp(&hashInt) == 1 {
		return true
	}

	return false;
}


//数据拼接成字节数组
func (proofOfWork *ProofOfWork) ZQ_prepareData(nonce int) []byte {
	data := bytes.Join(
		[][]byte{
			proofOfWork.ZQ_Block.ZQ_PrevBlockHash,
			[]byte{},
			//proofOfWork.ZQ_Block.ZQ_Txs,
			ZQ_IntToHex(proofOfWork.ZQ_Block.ZQ_Timestamp),
			ZQ_IntToHex(int64(targetBit)),
			ZQ_IntToHex(int64(nonce)),
			ZQ_IntToHex(int64(proofOfWork.ZQ_Block.ZQ_Height)),
		},
		[]byte{

		},
	)

	return data
}

//挖矿
func (proofOfWork *ProofOfWork) ZQ_Run() ([]byte, int64) {

	nonce := 0

	var hash [32]byte
	var hashInt big.Int

	for {
		//将block的属性拼接成字节数组
		dataBytes := proofOfWork.ZQ_prepareData(nonce)
		//生成hash
		hash = sha256.Sum256(dataBytes)
		fmt.Printf("\r%x", hash)
		//hash转换为int
		hashInt.SetBytes(hash[:])
		//判断hash有效性 满足条件， 跳出循环
		if proofOfWork.ZQ_target.Cmp(&hashInt) == 1 {
			break
		}

		nonce = nonce + 1
	}

	return hash[:], int64(nonce)
}

//创建新的工作量证明对象
func ZQ_NewProofOfWork(block *Block) *ProofOfWork {

	//创建一个初始值为1的target 左移256 - targetBit
	target := big.NewInt(1)
	target = target.Lsh(target, 256-targetBit)

	return &ProofOfWork{block, target}
}
