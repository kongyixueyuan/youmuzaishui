package BLC

import (
	"time"
	"strconv"
	"bytes"
	"crypto/sha256"
	"fmt"
	"encoding/gob"
	"log"
)

//最终以字节数组存放
type Block struct {
	//区块高度 编号
	ZQ_Height int64
	//上一个区块的Hash
	ZQ_PrevBlockHash []byte
	//交易数据
	ZQ_Txs []*Transaction
	//时间戳
	ZQ_Timestamp int64
	//ZQ_Hash
	ZQ_Hash []byte
	//ZQ_Nonce
	ZQ_Nonce int64
}

//将区块序列化为字节数组
func (block *Block) Serialize() []byte {

	var result bytes.Buffer

	encoder := gob.NewEncoder(&result)

	err := encoder.Encode(block)
	if err != nil {
		log.Panic(err)
	}

	return result.Bytes()
}

//反序列化字节数据为区块对象
func Deserialize(blockBytes []byte) *Block {

	var block Block

	decoder := gob.NewDecoder(bytes.NewReader(blockBytes))

	err := decoder.Decode(&block)
	if err != nil {
		log.Panic(err)
	}

	return &block
}

func (block *Block) SetHash() {
	//ZQ_Height -> 字节数组
	heightBytes := ZQ_IntToHex(block.ZQ_Height)

	//ZQ_Timestamp -> 字节数组
	timeString := strconv.FormatInt(block.ZQ_Timestamp, 2)
	timeBytes := []byte(timeString)

	//拼接
	blockBytes := bytes.Join([][]byte{heightBytes, block.ZQ_PrevBlockHash, []byte{}, timeBytes, block.ZQ_Hash}, []byte{})

	//生成Hash
	hash := sha256.Sum256(blockBytes)

	block.ZQ_Hash = hash[:]
}

//创建新的区块
func NewBlock(txs []*Transaction, height int64, prevBlockHash []byte) *Block {

	//创建区块
	block := &Block{height, prevBlockHash, txs, time.Now().Unix(), nil, 0}

	//生成Hash
	//block.SetHash()

	//调用工作量证明 且 返回有效的Hash ZQ_Nonce
	pow := ZQ_NewProofOfWork(block)
	hash, nonce := pow.ZQ_Run()

	block.ZQ_Hash = hash[:]
	block.ZQ_Nonce = nonce

	fmt.Println()

	return block
}

//生成创世区块
func CreateGenesisBlock(tx *Transaction) *Block {

	return NewBlock([]*Transaction{tx}, 1, []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0})
}
