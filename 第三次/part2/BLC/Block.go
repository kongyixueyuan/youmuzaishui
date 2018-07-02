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
	Height int64
	//上一个区块的Hash
	PrevBlockHash []byte
	//交易数据
	Data []byte
	//时间戳
	Timestamp int64
	//Hash
	Hash []byte
	//Nonce
	Nonce int64
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
	//Height -> 字节数组
	heightBytes := IntToHex(block.Height)

	//Timestamp -> 字节数组
	timeString := strconv.FormatInt(block.Timestamp, 2)
	timeBytes := []byte(timeString)

	//拼接
	blockBytes := bytes.Join([][]byte{heightBytes, block.PrevBlockHash, block.Data, timeBytes, block.Hash}, []byte{})

	//生成Hash
	hash := sha256.Sum256(blockBytes)

	block.Hash = hash[:]
}

//创建新的区块
func NewBlock(data string, height int64, prevBlockHash []byte) *Block {

	//创建区块
	block := &Block{height, prevBlockHash, []byte(data), time.Now().Unix(), nil, 0}

	//生成Hash
	//block.SetHash()

	//调用工作量证明 且 返回有效的Hash Nonce
	pow := NewProofOfWork(block)
	hash, nonce := pow.Run()

	block.Hash = hash[:]
	block.Nonce = nonce

	fmt.Println()

	return block
}

//生成创世区块
func CreateGenesisBlock(data string) *Block {

	return NewBlock(data, 1, []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0})
}
