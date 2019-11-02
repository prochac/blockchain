package main

import (
	"fmt"
	// "github.com/google/uuid"
)

type Node struct {
	ID         string
	BlockChain BlockChain
}

func (n Node) ListenForInput() {
	for {
		fmt.Println("Please choose")
		fmt.Println("1: Add a new Transaction value")
		fmt.Println("2: Mine a new Block")
		fmt.Println("3: Output the clockchain blocks")
		fmt.Println("4: Output participants")
		fmt.Println("5: Check Transaction validity")
		fmt.Println("h: Manipulate the chain")
		fmt.Println("q: Quit")

		switch n.GetUserChoice() {
		case "1":
			txRecipient, txAmount := n.GetTransactionValue()
			if !n.BlockChain.AddTransaction(txRecipient, n.ID, txAmount) {
				fmt.Println("Transaction failed!")
				break
			}
			fmt.Println("Added Transaction!")
		case "2":
			n.BlockChain.MineBlock()
		case "3":
			n.PrintBlockChainElements()
		case "4":
			fmt.Println(participants)
		case "5":
			if !Verification.VerifyTransactions(n.BlockChain.OpenTransactions(), n.BlockChain.GetBalance) {
				fmt.Println("There are invalid Transactions")
				break
			}
			fmt.Println("All Transactions are valid")
		case "q":
			return
		default:
			fmt.Println("Input was invalid, please pick a value from the list!")
		}
		fmt.Printf("Balance of %s: %6.2f\n", n.ID, n.BlockChain.GetBalance())

		if !Verification.VerifyChain(n.BlockChain) {
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

func main() {
	// id := uuid.New().String()
	id := "Tom"
	n := Node{
		ID: id,
	}
	n.BlockChain = BlockChain{HostingNode: &n}
	n.BlockChain.LoadData()

	n.ListenForInput()
}
