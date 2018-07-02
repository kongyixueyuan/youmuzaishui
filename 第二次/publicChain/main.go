package main

import (
	"publicChain/BLC"
	"fmt"
)

func main() {

	//创建区块链
	blockchain := BLC.CreateBlockchainWithGenesisBlock()

	//新区块
	blockchain.AddBlockToBlockchain("hi hi hi ", blockchain.Blocks[len(blockchain.Blocks) - 1].Height + 1, blockchain.Blocks[len(blockchain.Blocks) - 1].Hash)

	blockchain.AddBlockToBlockchain("hey hey hey ", blockchain.Blocks[len(blockchain.Blocks) - 1].Height + 1, blockchain.Blocks[len(blockchain.Blocks) - 1].Hash)

	blockchain.AddBlockToBlockchain("ha ha ha ", blockchain.Blocks[len(blockchain.Blocks) - 1].Height + 1, blockchain.Blocks[len(blockchain.Blocks) - 1].Hash)

	fmt.Println(blockchain.Blocks)
}