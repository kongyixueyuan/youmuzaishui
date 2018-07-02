package BLC

import (
	"math/big"
	"bytes"
	"crypto/sha256"
	"fmt"
)

type ProofOfWork struct {
	//当前要验证的区块
	Block *Block
	//大数据存储 难度
	target *big.Int
}

const targetBit = 16

//判断hash有效性
func (proofOfWork *ProofOfWork) IsValid() bool {
	//hash 与 target 进行比较
	var hashInt big.Int
	hashInt.SetBytes(proofOfWork.Block.Hash)
	if proofOfWork.target.Cmp(&hashInt) == 1 {
		return true
	}

	return false;
}

//数据拼接成字节数组
func (proofOfWork *ProofOfWork) prepareData(nonce int) []byte {
	data := bytes.Join(
		[][]byte{
			proofOfWork.Block.PrevBlockHash,
			proofOfWork.Block.Data,
			IntToHex(proofOfWork.Block.Timestamp),
			IntToHex(int64(targetBit)),
			IntToHex(int64(nonce)),
			IntToHex(int64(proofOfWork.Block.Height)),
		},
		[]byte{

		},
	)

	return data
}

//挖矿
func (proofOfWork *ProofOfWork) Run() ([]byte, int64) {

	nonce := 0

	var hash [32]byte
	var hashInt big.Int

	for {
		//将block的属性拼接成字节数组
		dataBytes := proofOfWork.prepareData(nonce)
		//生成hash
		hash = sha256.Sum256(dataBytes)
		fmt.Printf("\r%x", hash)
		//hash转换为int
		hashInt.SetBytes(hash[:])
		//判断hash有效性 满足条件， 跳出循环
		if proofOfWork.target.Cmp(&hashInt) == 1 {
			break
		}

		nonce = nonce + 1
	}

	return hash[:], int64(nonce)
}

//创建新的工作量证明对象
func NewProofOfWork(block *Block) *ProofOfWork {

	//创建一个初始值为1的target 左移256 - targetBit
	target := big.NewInt(1)
	target = target.Lsh(target, 256-targetBit)

	return &ProofOfWork{block, target}
}
