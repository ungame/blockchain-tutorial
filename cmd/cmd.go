package cmd

import (
	"blockchain-tutorial/blockchain"
	"blockchain-tutorial/utils"
	"blockchain-tutorial/wallet"
	"errors"
	"flag"
	"fmt"
	"os"
	"runtime"
	"strconv"
	"strings"
)

type commandLine struct {
}

func NewCommandLine() *commandLine {
	return &commandLine{}
}

func (c *commandLine) usage() {
	fmt.Println("Usage:")
	fmt.Println(" init -address ADDRESS initialize a blockchain")
	fmt.Println(" print - Prints the blocks in the chain")
	fmt.Println(" send -from FROM -to TO -amount AMOUNT - Transfer coins")
	fmt.Println(" getbalance -address ADDRESS - Get the balance for an address")
	fmt.Println(" createwallet - Create a new Wallet")
	fmt.Println(" listaddresses - List the addresses in our wallet file")
}

func (c *commandLine) validate() {
	if len(os.Args) < 2 {
		c.usage()
		runtime.Goexit()
	}
}

func (c *commandLine) init(address string) {
	if !wallet.ValidateAddress(address) {
		utils.HandleError(wallet.ErrInvalidAddress)
	}

	chain := blockchain.InitBlockChain(address)
	chain.Close()
	fmt.Println("BlockChain initialized!")
}

func (c *commandLine) print() {

	chain := blockchain.ContinueBlockChain("")
	defer chain.Close()
	it := chain.Iterator()

	for {

		block := it.Next()

		pow := blockchain.NewProofOfWork(block)

		block.Info(strconv.FormatBool(pow.Validate()))

		if len(block.PrevHash) == 0 {
			break
		}
	}
}

func (c *commandLine) getBalance(address string) {
	if !wallet.ValidateAddress(address) {
		utils.HandleError(wallet.ErrInvalidAddress)
	}

	chain := blockchain.ContinueBlockChain(address)
	defer chain.Close()

	balance := 0
	pubKeyHash := wallet.Base58Decode([]byte(address))
	pubKeyHash = pubKeyHash[1: len(pubKeyHash) - 4]
	UTXOs := chain.FindUTXO(pubKeyHash)

	for _, out := range UTXOs {
		balance += out.Value
	}

	fmt.Printf("Balance of %s: %d\n", address, balance)
}

func (c *commandLine) send(sender, receiver string, amount int) {
	if !wallet.ValidateAddress(sender) {
		utils.HandleError(wallet.ErrInvalidAddress)
	}
	if !wallet.ValidateAddress(receiver) {
		utils.HandleError(wallet.ErrInvalidAddress)
	}

	chain := blockchain.ContinueBlockChain(sender)
	defer chain.Close()

	tx := blockchain.NewTransaction(sender, receiver, amount, chain)
	chain.AddBlock([]*blockchain.Transaction{tx})
	fmt.Println("SUCCESS!")
}

func (c *commandLine) listAddresses() {
	wallets, _ := wallet.LoadWallets()
	addresses := wallets.GetAddresses()
	for index := range addresses {
		fmt.Println(addresses[index])
	}
}

func (c *commandLine) createWallet() {
	wallets, _ := wallet.LoadWallets()
	address, pvtKey := wallets.AddWallet()
	wallets.SaveFile()
	fmt.Println("*********************************** WALLET ***********************************")
	fmt.Printf("New address: %s\n", address)
	fmt.Printf("Private Key: %s\n", strings.ToUpper(pvtKey))
}

func (c *commandLine) Run() {
	c.validate()

	initBlockChainCmd := flag.NewFlagSet("init", flag.ExitOnError)
	printChainCmd := flag.NewFlagSet("print", flag.ExitOnError)
	sendCmd := flag.NewFlagSet("send", flag.ExitOnError)
	getBalanceCmd := flag.NewFlagSet("getbalance", flag.ExitOnError)
	createWalletCmd := flag.NewFlagSet("createwallet", flag.ExitOnError)
	listAddressesCmd := flag.NewFlagSet("listaddresses", flag.ExitOnError)

	initBlockChainAddress := initBlockChainCmd.String("address", "", "The address in BlockChain")
	getBalanceAddress := getBalanceCmd.String("address", "", "The address in BlockChain")
	sendFrom := sendCmd.String("from", "", "Source wallet address")
	sendTo := sendCmd.String("to", "", "Destination wallet address")
	sendAmount := sendCmd.Int("amount", 0, "Amount to send")

	switch os.Args[1] {
	case "init":
		err := initBlockChainCmd.Parse(os.Args[2:])
		utils.HandleError(err)

	case "print":
		err := printChainCmd.Parse(os.Args[2:])
		utils.HandleError(err)

	case "send":
		err := sendCmd.Parse(os.Args[2:])
		utils.HandleError(err)

	case "getbalance":
		err := getBalanceCmd.Parse(os.Args[2:])
		utils.HandleError(err)

	case "createwallet":
		err := createWalletCmd.Parse(os.Args[2:])
		utils.HandleError(err)

	case "listaddresses":
		err := listAddressesCmd.Parse(os.Args[2:])
		utils.HandleError(err)

	default:
		c.usage()
		runtime.Goexit()
	}

	if initBlockChainCmd.Parsed() {
		c.init(*initBlockChainAddress)
	}

	if printChainCmd.Parsed() {
		c.print()
	}

	if sendCmd.Parsed() {
		err := validateSend(*sendFrom, *sendTo, *sendAmount)
		if  err != nil {
			fmt.Println("ERROR: ", err.Error())
			sendCmd.Usage()
			runtime.Goexit()
		}
		c.send(*sendFrom, *sendTo, *sendAmount)
	}

	if getBalanceCmd.Parsed() {
		c.getBalance(*getBalanceAddress)
	}

	if createWalletCmd.Parsed() {
		c.createWallet()
	}

	if listAddressesCmd.Parsed() {
		c.listAddresses()
	}
}

func validateSend(from, to string, amount int) error {
	if strings.TrimSpace(from) == "" {
		return errors.New("invalid -from address")
	}
	if strings.TrimSpace(to) == "" {
		return errors.New("invalid -to address")
	}
	if amount <= 0 {
		return fmt.Errorf("invalid -amount = %v", amount)
	}
	return nil
}
