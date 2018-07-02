package BLC

type Blockchain struct {
	Blocks []*Block
}

//增加区块至区块链
func (blockchain *Blockchain) AddBlockToBlockchain(data string, height int64, prevHash []byte)  {
	//创建新区块
	newBlock := NewBlock(data, height, prevHash)
	//添加区块到链
	blockchain.Blocks = append(blockchain.Blocks, newBlock)
}

//创建带有创世区块的区块链
func CreateBlockchainWithGenesisBlock() *Blockchain{
	//创建创世区块
	genesisBlock := CreateGenesisBlock("Genesis Data...")
	//返回区块链对象
	return &Blockchain{[]*Block{genesisBlock}}
}
