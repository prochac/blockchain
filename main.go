package main

import (
	"fmt"
	"os"
)

type blockchain []block
type block struct {
	previousHash string
	index        int64
	transactions []transaction
}

func (b block) Hash() string {
	return fmt.Sprintf("%s-%d-%v",
		b.previousHash,
		b.index,
		b.transactions,
	)
}

type transaction struct {
	Sender    string
	Recipient string
	Amount    float64
}

var (
	bch              = blockchain{{}}
	openTransactions []transaction
	owner            = "Tomas"
	participants     = map[string]struct{}{
		"Tomas": struct{}{},
	}
)

func getLastBlock(bch blockchain) *block {
	return &bch[len(bch)-1]
}

func addTransaction(recipient string, amount float64) {
	addTransactionWithSender(owner, recipient, amount)
}

func addTransactionWithSender(sender, recipient string, amount float64) {
	tx := transaction{
		Sender:    sender,
		Recipient: recipient,
		Amount:    amount,
	}
	openTransactions = append(openTransactions, tx)
	participants[sender] = struct{}{}
	participants[recipient] = struct{}{}
}

func mineBlock() {
	b := block{
		previousHash: getLastBlock(bch).Hash(),
		index:        int64(len(bch)),
		transactions: openTransactions,
	}
	bch = append(bch, b)
}

func getTransactionValue() (string, float64) {
	fmt.Print("Enter the recipient of the transaction: ")
	var s string
	if _, err := fmt.Scanf("%s", &s); err != nil {
		panic(err)
	}

	fmt.Print("Your transaction amount please: ")
	var f float64
	if _, err := fmt.Scanf("%f", &f); err != nil {
		panic(err)
	}
	return s, f
}

func verifyChain() bool {
	for i, b := range bch {
		if i == 0 {
			continue
		}

		if b.previousHash != bch[i-1].Hash() {
			return false
		}
	}
	return true
}

func getUserChoice() string {
	var s string
	if _, err := fmt.Scan(&s); err != nil {
		panic(err)
	}
	return s
}

func main() {
	for {
		fmt.Println("Please choose")
		fmt.Println("1: Add a new transaction value")
		fmt.Println("2: Mine a new block")
		fmt.Println("3: Output the clockchain blocks")
		fmt.Println("4: Output participants")
		fmt.Println("h: Manipulate the chain")
		fmt.Println("q: Quit")

		switch getUserChoice() {
		case "1":
			txSender, txAmount := getTransactionValue()
			addTransaction(txSender, txAmount)
		case "2":
			mineBlock()
		case "3":
			fmt.Println(bch)
		case "4":
			fmt.Println(participants)
		case "h":
			if len(bch) < 1 {
				continue
			}
			bch[0] = block{
				transactions: []transaction{
					{
						Sender:    "Chris",
						Recipient: "max",
						Amount:    100,
					},
				},
			}
		case "q":
			os.Exit(0)
		default:
			fmt.Println("Input was invalid, please pick a value from the list!")
		}

		if !verifyChain() {
			panic("chain is broken")
		}
	}
}
