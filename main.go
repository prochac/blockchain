package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
)

const MiningReward = 10

type blockchain []block
type block struct {
	previousHash string
	index        int64
	transactions []transaction
	proof        uint64
}

type transaction struct {
	Sender    string
	Recipient string
	Amount    float64
}

var (
	bch              = blockchain{{}}
	openTransactions = []transaction{}
	owner            = "Tomas"
	participants     = map[string]struct{}{
		"Tomas": struct{}{},
	}
)

func validProof(tx []transaction, lastHash string, proof uint64) bool {
	b, _ := json.Marshal(tx)
	guess := string(b) + lastHash + strconv.FormatUint(proof, 10)
	h := HashString256(guess)
	fmt.Println(h)
	return h[:2] == "00"
}

func proofOfWork() uint64 {
	lastHash := getLastBlock(bch).Hash()
	var proof uint64
	for !validProof(openTransactions, lastHash, proof) {
		proof++
	}
	return proof
}

func getBalance(participant string) float64 {
	var (
		txSender     float64
		openTxSender float64
		txRecipient  float64
	)
	for _, block := range bch {
		for _, tx := range block.transactions {
			if tx.Sender == participant {
				txSender += tx.Amount
			}
			if tx.Recipient == participant {
				txRecipient += tx.Amount
			}
		}
	}
	for _, tx := range openTransactions {
		if tx.Sender == participant {
			openTxSender += tx.Amount
		}
	}

	return txRecipient - (txSender + openTxSender)
}

func getLastBlock(bch blockchain) *block {
	return &bch[len(bch)-1]
}

func verifyTransaction(tx transaction) bool {
	return getBalance(tx.Sender) >= tx.Amount
}

func addTransaction(recipient string, amount float64) bool {
	return addTransactionWithSender(owner, recipient, amount)
}

func addTransactionWithSender(sender, recipient string, amount float64) bool {
	tx := transaction{
		Sender:    sender,
		Recipient: recipient,
		Amount:    amount,
	}

	if !verifyTransaction(tx) {
		return false
	}

	openTransactions = append(openTransactions, tx)
	participants[sender] = struct{}{}
	participants[recipient] = struct{}{}
	return true
}

func mineBlock() {
	hashedBlock := getLastBlock(bch).Hash()
	proof := proofOfWork()

	rewardTx := transaction{
		Sender:    "MINING",
		Recipient: owner,
		Amount:    MiningReward,
	}

	copiedTransactions := make([]transaction, len(openTransactions))
	copy(copiedTransactions, openTransactions)

	copiedTransactions = append(copiedTransactions, rewardTx)

	b := block{
		previousHash: hashedBlock,
		index:        int64(len(bch)),
		transactions: copiedTransactions,
		proof:        proof,
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
		if !validProof(b.transactions[:len(b.transactions)-1], b.previousHash, b.proof) {
			fmt.Println("Proof of work is invalid")
			return false
		}
	}
	return true
}

func verifyTransactions() bool {
	for _, tx := range openTransactions {
		if !verifyTransaction(tx) {
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
		fmt.Println("4: Check transaction validity")
		fmt.Println("h: Manipulate the chain")
		fmt.Println("q: Quit")

		switch getUserChoice() {
		case "1":
			txSender, txAmount := getTransactionValue()
			if !addTransaction(txSender, txAmount) {
				fmt.Println("Transaction failed!")
				break
			}
			fmt.Println("Added transaction!")
		case "2":
			mineBlock()
			openTransactions = nil
		case "3":
			fmt.Println(bch)
		case "4":
			fmt.Println(participants)
		case "5":
			if !verifyTransactions() {
				fmt.Println("There are invalid transactions")
				break
			}
			fmt.Println("All transactions are valid")
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
		fmt.Printf("Balance of %s: %6.2f\n", owner, getBalance(owner))

		if !verifyChain() {
			panic("chain is broken")
		}
	}
}
