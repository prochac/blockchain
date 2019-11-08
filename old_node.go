package main

import (
	"fmt"
)

type Node struct {
	Wallet     Wallet
	BlockChain BlockChain
}

func (n *Node) ListenForInput() {
	for {
		fmt.Println("Please choose")
		fmt.Println("1: Add a new transaction value")
		fmt.Println("2: Mine a new block")
		fmt.Println("3: Output the blockchain blocks")
		fmt.Println("4: Check transaction validity")
		fmt.Println("5: Create wallet")
		fmt.Println("6: Load wallet")
		fmt.Println("7: Save keys")
		fmt.Println("q: Quit")

		switch n.GetUserChoice() {
		case "1":
			txRecipient, txAmount := n.GetTransactionValue()
			signature := n.Wallet.SignTransaction(n.Wallet.PublicKey, txRecipient, txAmount)
			if !n.BlockChain.AddTransaction(txRecipient, n.Wallet.PublicKey, signature, txAmount) {
				fmt.Println("Transaction failed!")
				break
			}
			fmt.Println("Added Transaction!")
		case "2":
			if n.BlockChain.MineBlock() == nil {
				fmt.Println("Mining failed. Got no wallet?")
			}
		case "3":
			n.PrintBlockChainElements()
		case "4":
			if !Verification.VerifyTransactions(n.BlockChain.OpenTransactions()) {
				fmt.Println("There are invalid Transactions")
				break
			}
			fmt.Println("All Transactions are valid")
		case "5":
			n.Wallet.CreateKeys()
			n.BlockChain.PublicKey = n.Wallet.PublicKey
			n.BlockChain.LoadData()
		case "6":
			n.Wallet.LoadKeys()
			n.BlockChain.PublicKey = n.Wallet.PublicKey
			n.BlockChain.LoadData()
		case "7":
			n.Wallet.SaveKeys()
		case "q":
			return
		default:
			fmt.Println("Input was invalid, please pick a value from the list!")
		}
		fmt.Printf("Balance of %s: %6.2f\n", n.Wallet, n.BlockChain.GetBalance())

		if !Verification.VerifyChain(n.BlockChain.Chain()) {
			panic("chain is broken")
		}
	}
}

func (n Node) GetTransactionValue() (string, float64) {
	fmt.Print("Enter the recipient of the Transaction: ")
	var s string
	if _, err := fmt.Scanf("%s", &s); err != nil {
		panic(err)
	}

	fmt.Print("Your Transaction amount please: ")
	var f float64
	if _, err := fmt.Scanf("%f", &f); err != nil {
		panic(err)
	}
	return s, f
}

func (n Node) GetUserChoice() string {
	var s string
	if _, err := fmt.Scan(&s); err != nil {
		panic(err)
	}
	return s
}

func (n Node) PrintBlockChainElements() {
	for _, block := range n.BlockChain.Chain() {
		fmt.Println(block)
	}
}

func _main() {
	var n Node
	n.Wallet.CreateKeys()
	n.BlockChain = BlockChain{PublicKey: n.Wallet.PublicKey}
	n.BlockChain.LoadData()
	n.ListenForInput()
}
