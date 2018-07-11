package BLC

import (
	"fmt"
	"os"
	"flag"
	"log"
	"encoding/json"
)

type CLI struct{}

func printUsage() {

	fmt.Println("Usage:")
	fmt.Println("\tcreateblockchain -address 交易数据")
	fmt.Println("\tcreatewallet --创建钱包")
	fmt.Println("\taddresslist --输出所有钱包地址")
	fmt.Println("\tsend -from 源地址 -to 目的地址 -amount 金额")
	fmt.Println("\tprintchain --输出区块信息")
	fmt.Println("\tgetbalance -address --输入余额")
	fmt.Println("\ttest --测试")
}


//测试方法
func (cli *CLI) TestMethod() {

	blockchain := ZQ_GetBlockchainObject()

	defer blockchain.ZQ_DB.Close()

	utxoSet := &UTXOSet{blockchain}
	utxoSet.ZQ_ResetUTXOSet()
}

//打印所有钱包地址
func (cli *CLI) addressList() {

	fmt.Println("打印所有钱包地址")

	wallets, _ := ZQ_NewWallets()

	for address, _ := range wallets.ZQ_WalletsMap {

		fmt.Println(address)
	}
}

//创建钱包
func (cli *CLI) createWallet() {

	wallets, _ := ZQ_NewWallets()
	wallets.ZQ_CreateNewWallet()

	fmt.Println(wallets.ZQ_WalletsMap)
}

//添加区块
func (cli *CLI) addBlock(data string) {

	if ZQ_DBExists() == false {
		fmt.Println("数据库不存在...")
		os.Exit(1)
	}

	blockchain := ZQ_GetBlockchainObject()

	defer blockchain.ZQ_DB.Close()

	blockchain.ZQ_AddBlockToBlockchain([]*Transaction{})
}

//打印区块链
func (cli *CLI) printchain() {

	if ZQ_DBExists() == false {
		fmt.Println("数据库不存在...")
		os.Exit(1)
	}

	blockchain := ZQ_GetBlockchainObject()

	defer blockchain.ZQ_DB.Close()

	blockchain.ZQ_Printchain()
}

//创建创世区块 重置未花费交易
func (cli *CLI) createGenesisBlockchani(address string) {

	blockchain := ZQ_CreateBlockchainWithGenesisBlock(address)

	defer blockchain.ZQ_DB.Close()

	utxoSet := &UTXOSet{blockchain}

	utxoSet.ZQ_ResetUTXOSet()
}

//进行交易
func (cli *CLI) send(from []string, to []string, amount []string) {

	if ZQ_DBExists() == false {
		fmt.Println("数据库不存在...")
		os.Exit(1)
	}

	blockchain := ZQ_GetBlockchainObject()
	defer blockchain.ZQ_DB.Close()

	//验证
	blockchain.ZQ_MineNewBlock(from, to, amount)

	//转账成功 更新utxo
	utxoSet := &UTXOSet{blockchain}
	utxoSet.Update()
}

//获取用户余额
func (cli *CLI) getBalance(address string) {

	fmt.Println("address:", address)

	if ZQ_DBExists() == false {
		fmt.Println("数据库不存在...")
		os.Exit(1)
	}

	blockchain := ZQ_GetBlockchainObject()
	defer blockchain.ZQ_DB.Close()

	utxoSet := &UTXOSet{blockchain}

	amount := utxoSet.ZQ_GetBalance(address)

	fmt.Printf("%s一共有%d个Token\n", address, amount)
}

func (cli *CLI) Run() {

	isValidArgs()

	testMethodCmd := flag.NewFlagSet("test", flag.ExitOnError)
	sendBlockCmd := flag.NewFlagSet("send", flag.ExitOnError)
	printChainCmd := flag.NewFlagSet("printchain", flag.ExitOnError)
	createBlockchainCmd := flag.NewFlagSet("createblockchain", flag.ExitOnError)
	getBalanceCmd := flag.NewFlagSet("getbalance", flag.ExitOnError)
	createWalletCmd := flag.NewFlagSet("createwallet", flag.ExitOnError)
	addressListCmd := flag.NewFlagSet("addresslist", flag.ExitOnError)

	fromData := sendBlockCmd.String("from", "", "转账源地址...")
	toData := sendBlockCmd.String("to", "", "转账目的地址...")
	amountData := sendBlockCmd.String("amount", "", "转账金额...")
	createBlockchainData := createBlockchainCmd.String("address", "create Genesisblock", "创世区块的地址")
	getBalanceData := getBalanceCmd.String("address", "", "查询余额")

	switch os.Args[1] {

	case "send":
		err := sendBlockCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "test":
		err := testMethodCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}

	case "printchain":
		err := printChainCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}

	case "createblockchain":
		err := createBlockchainCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}

	case "createwallet":
		err := createWalletCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}

	case "addresslist":
		err := addressListCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}

	case "getbalance":
		err := getBalanceCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}

	default:
		printUsage()
		os.Exit(1)
	}

	if testMethodCmd.Parsed() {

		cli.TestMethod()
	}

	if sendBlockCmd.Parsed() {
		if *fromData == "" || *toData == "" || *amountData == "" {
			printUsage()
			os.Exit(1)
		}

		from := JSONToArray(*fromData)
		to := JSONToArray(*toData)
		amount := JSONToArray(*amountData)

		for index, fromAddress := range from {
			if ZQ_IsValidForAddress([]byte(fromAddress)) == false || ZQ_IsValidForAddress([]byte(to[index])) == false {
				fmt.Println("地址无效或不合法")
				os.Exit(1)
			}
		}

		cli.send(from, to, amount)

	}

	if printChainCmd.Parsed() {
		cli.printchain()
	}

	if addressListCmd.Parsed() {
		cli.addressList()
	}

	if createWalletCmd.Parsed() {
		cli.createWallet()
	}

	if createBlockchainCmd.Parsed() {

		if ZQ_IsValidForAddress([]byte(*createBlockchainData)) == false {
			fmt.Println("地址无效或不合法")
			printUsage()
			os.Exit(1)
		}

		cli.createGenesisBlockchani(*createBlockchainData)
	}

	if getBalanceCmd.Parsed() {
		if ZQ_IsValidForAddress([]byte(*getBalanceData)) == false {
			fmt.Println("地址无效或不合法")
			printUsage()
			os.Exit(1)
		}

		cli.getBalance(*getBalanceData)
	}
}

//json -> 数组
func JSONToArray(jsonString string) []string {

	var sArr []string

	if err := json.Unmarshal([]byte(jsonString), &sArr); err != nil {
		log.Panic(err)
	}

	return sArr
}

//参数有效性判断
func isValidArgs() {

	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}
}
