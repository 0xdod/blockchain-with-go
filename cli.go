package main

import (
	"flag"
	"fmt"
	"os"
)

type CLI struct{}

func (cli *CLI) printUsage() {
	fmt.Println("Usage: ")
	fmt.Println("\tcreateblockchain -addr ADDRESS - create a blockchain and send genesis reward to ADDRESS")
	fmt.Println("\tbalance -addr ADDRESS - gets the a balance for ADDRESS")
	fmt.Println("\tprintchain - prints all the blocks in the blockchain")
	fmt.Println("\tsend - send -from FROM -to TO -amount AMOUNT  - send AMOUNT worth of coins from FROM address to TO")

}

func (cli *CLI) validateArgs() {
	if len(os.Args) < 2 {
		cli.printUsage()
		os.Exit(1)
	}
}

func (cli *CLI) Run() {
	cli.validateArgs()

	createBlockchainCmd := flag.NewFlagSet("createblockchain", flag.ExitOnError)
	printChainCmd := flag.NewFlagSet("printchain", flag.ExitOnError)
	getBalanceCmd := flag.NewFlagSet("balance", flag.ExitOnError)
	sendCmd := flag.NewFlagSet("send", flag.ExitOnError)

	createBlockchainAddr := createBlockchainCmd.String("addr", "", "Address to send genesis block reward to")
	balanceAddr := getBalanceCmd.String("addr", "", "address to find balance for")
	sendFromAddr := sendCmd.String("from", "", "address to send from")
	sendToAddr := sendCmd.String("to", "", "address to send to")
	sendAmount := sendCmd.Int("amount", 0, "amount to send")

	var err error
	switch os.Args[1] {
	case "createblockchain":
		err = createBlockchainCmd.Parse(os.Args[2:])
	case "printchain":
		err = printChainCmd.Parse(os.Args[2:])
	case "balance":
		err = getBalanceCmd.Parse(os.Args[2:])
	case "send":
		err = sendCmd.Parse(os.Args[2:])

	default:
		cli.printUsage()
		os.Exit(1)
	}

	if err != nil {
		panic(err)
	}

	if createBlockchainCmd.Parsed() {
		if *createBlockchainAddr == "" {
			createBlockchainCmd.Usage()
			os.Exit(1)
		}

		cli.createBlockchain(*createBlockchainAddr)
	}

	if sendCmd.Parsed() {
		if *sendFromAddr == "" || *sendToAddr == "" || *sendAmount <= 0 {
			sendCmd.Usage()
			os.Exit(1)
		}

		cli.send(*sendFromAddr, *sendToAddr, *sendAmount)
	}

	if getBalanceCmd.Parsed() {
		if *balanceAddr == "" {
			getBalanceCmd.Usage()
			os.Exit(1)
		}

		cli.getBalance(*balanceAddr)
	}

	if printChainCmd.Parsed() {
		cli.printChain()
	}
}

func (cli *CLI) createBlockchain(address string) {
	bc := CreateBlockchain(address)
	bc.db.Close()
	fmt.Println("Done!")
}

func (cli *CLI) printChain() {
	bc := NewBlockchain()
	defer bc.db.Close()

	bci := bc.Iterator()

	for bci.Scan() {
		block := bci.Next()
		fmt.Print(block)
	}
}

func (cli *CLI) getBalance(address string) {
	bc := NewBlockchain()
	defer bc.db.Close()

	bal := 0
	UTXOs := bc.FindUTXO(address)

	for _, out := range UTXOs {
		bal += out.Value
	}

	fmt.Printf("Balance of %q: %d\n", address, bal)
}

func (cli *CLI) send(from, to string, amount int) {
	bc := NewBlockchain()
	defer bc.db.Close()

	tx := NewUTXOTransaction(from, to, amount, bc)

	bc.MineBlock([]*Transaction{tx})

	fmt.Println("Success!")
}
