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
	fmt.Println("\tsend -from 源地址 -to 目的地址 -amount 金额")
	fmt.Println("\tprintchain --输出区块信息")
	fmt.Println("\tgetbalance -address --输入余额")
}

func (cli *CLI) addBlock(data string) {

	if DBExists() == false {
		fmt.Println("数据库不存在...")
		os.Exit(1)
	}

	blockchain := GetBlockchainObject()

	defer blockchain.DB.Close()

	blockchain.AddBlockToBlockchain([]*Transaction{})
}

//打印区块链
func (cli *CLI) printchain() {

	if DBExists() == false {
		fmt.Println("数据库不存在...")
		os.Exit(1)
	}

	blockchain := GetBlockchainObject()

	defer blockchain.DB.Close()

	blockchain.Printchain()
}

//创建创世区块
func (cli *CLI) createGenesisBlockchani(address string) {

	blockchain := CreateBlockchainWithGenesisBlock(address)
	defer blockchain.DB.Close()
}

//进行交易
func (cli *CLI) send(from []string, to []string, amount []string)  {

	if DBExists() == false {
		fmt.Println("数据库不存在...")
		os.Exit(1)
	}

	blockchain := GetBlockchainObject()
	defer blockchain.DB.Close()

	blockchain.MineNewBlock(from, to, amount)
}

//获取用户余额
func (cli *CLI) getBalance(address string) {

	fmt.Println("address:", address)

	if DBExists() == false {
		fmt.Println("数据库不存在...")
		os.Exit(1)
	}

	blockchain := GetBlockchainObject()
	defer blockchain.DB.Close()

	amount := blockchain.GetBalance(address)

	fmt.Printf("%s一共有%d个Token\n",address,amount)
}



func (cli *CLI) Run() {

	isValidArgs()

	sendBlockCmd := flag.NewFlagSet("send", flag.ExitOnError)
	printChainCmd := flag.NewFlagSet("printchain", flag.ExitOnError)
	createBlockchainCmd := flag.NewFlagSet("createblockchain", flag.ExitOnError)
	getBalanceCmd := flag.NewFlagSet("getbalance", flag.ExitOnError)

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

	case "getbalance":
		err := getBalanceCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}

	default:
		printUsage()
		os.Exit(1)
	}

	if sendBlockCmd.Parsed() {
		if *fromData == "" || *toData == "" || *amountData == ""{
			printUsage()
			os.Exit(1)
		}

		from := JSONToArray(*fromData)
		to := JSONToArray(*toData)
		amount := JSONToArray(*amountData)

		cli.send(from, to, amount)

	}

	if printChainCmd.Parsed() {
		cli.printchain()
	}

	if createBlockchainCmd.Parsed() {
		if *createBlockchainData == "" {
			fmt.Println("地址不能为空...")
			printUsage()
			os.Exit(1)
		}

		cli.createGenesisBlockchani(*createBlockchainData)
	}

	if getBalanceCmd.Parsed() {
		if *getBalanceData == "" {
			fmt.Println("地址不能为空...")
			printUsage()
			os.Exit(1)
		}

		cli.getBalance(*getBalanceData)
	}
}

//json -> 数组
func JSONToArray(jsonString string)  []string{

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
