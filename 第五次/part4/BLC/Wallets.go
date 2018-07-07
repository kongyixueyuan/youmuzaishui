package BLC

import (
	"bytes"
	"encoding/gob"
	"crypto/elliptic"
	"log"
	"io/ioutil"
	"os"
)

const walletFile = "Wallets.dat"

type Wallets struct {
	WalletsMap map[string]*Wallet
}

//创建钱包集合
func NewWallets() (*Wallets, error) {

	if _, err := os.Stat(walletFile); os.IsNotExist(err) {
		wallets := &Wallets{}
		wallets.WalletsMap = make(map[string]*Wallet)
		return wallets, err
	}

	fileContent, err := ioutil.ReadFile(walletFile)
	if err != nil {
		log.Panic(err)
	}

	//取出已有数据
	var wallets Wallets
	gob.Register(elliptic.P256())
	decoder := gob.NewDecoder(bytes.NewReader(fileContent))
	err = decoder.Decode(&wallets)
	if err != nil {
		log.Panic(err)
	}

	return &wallets, nil
}

//创建钱包
func (wallets *Wallets) CreateNewWallet() {

	wallet := NewWallet()

	wallets.WalletsMap[string(wallet.GetAddress())] = wallet

	wallets.SaveWallets()
}

//钱包信息写入文件
func (wallets *Wallets) SaveWallets() {

	var content bytes.Buffer

	//注册是为了可序列化任何类型
	gob.Register(elliptic.P256())

	encoder := gob.NewEncoder(&content)
	err := encoder.Encode(&wallets)
	if err != nil {
		log.Panic(err)
	}

	//覆盖写入文件
	err = ioutil.WriteFile(walletFile, content.Bytes(), 0644)
	if err != nil {
		log.Panic(err)
	}
}

//加载包钱文件
func (wallets *Wallets) LoadFileWallets() error {


	return nil
}
