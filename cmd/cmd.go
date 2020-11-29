package cmd

import (
	"blockchain-tutorial/blockchain"
	"flag"
	"fmt"
	"os"
	"runtime"
	"strconv"
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
	fmt.Println(" getbalance -address ADDRESS - get the balance for the address")
}

func (c *commandLine) validate() {
	if len(os.Args) < 2 {
		c.usage()
		runtime.Goexit()
	}
}

func (c *commandLine) init(address string) {
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
	chain := blockchain.ContinueBlockChain(address)
	defer chain.Close()

	balance := 0
	UTXOs := chain.FindUTXO(address)

	for _, out := range UTXOs {
		balance += out.Value
	}

	fmt.Printf("Balance of %s: %d\n", address, balance)
}

func (c *commandLine) send(sender, receiver string, amount int) {
	chain := blockchain.ContinueBlockChain(sender)
	defer chain.Close()

	tx := blockchain.NewTransaction(sender, receiver, amount, chain)
	chain.AddBlock([]*blockchain.Transaction{tx})
	fmt.Println("SUCCESS!")
}

func (c *commandLine) Run() {
	c.validate()

	initBlockChainCmd := flag.NewFlagSet("init", flag.ExitOnError)
	printChainCmd := flag.NewFlagSet("print", flag.ExitOnError)
	sendCmd := flag.NewFlagSet("send", flag.ExitOnError)
	getBalanceCmd := flag.NewFlagSet("getbalance", flag.ExitOnError)

	initBlockChainAddress := initBlockChainCmd.String("address", "", "The address in BlockChain")
	getBalanceAddress := getBalanceCmd.String("address", "", "The address in BlockChain")
	sendFrom := sendCmd.String("from", "", "Source wallet address")
	sendTo := sendCmd.String("to", "", "Destination wallet address")
	sendAmount := sendCmd.Int("amount", 0, "Amount to send")

	switch os.Args[1] {
	case "init":
		err := initBlockChainCmd.Parse(os.Args[2:])
		blockchain.HandleError(err)

	case "print":
		err := printChainCmd.Parse(os.Args[2:])
		blockchain.HandleError(err)

	case "send":
		err := sendCmd.Parse(os.Args[2:])
		blockchain.HandleError(err)

	case "getbalance":
		err := getBalanceCmd.Parse(os.Args[2:])
		blockchain.HandleError(err)

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
		c.send(*sendFrom, *sendTo, *sendAmount)
	}

	if getBalanceCmd.Parsed() {
		c.getBalance(*getBalanceAddress)
	}

}