package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"
)

const MiningReward = 10
const Owner = "Tomas"

type BlockChain []Block
type Block struct {
	PreviousHash string
	Index        int64
	Transactions []Transaction
	Proof        uint64
}

type Transaction struct {
	Sender    string
	Recipient string
	Amount    float64
}

var (
	bch              BlockChain
	openTransactions []Transaction
	participants     = map[string]struct{}{
		Owner: struct{}{},
	}
)

func loadData() {
	f, err := os.Open("blockchain.txt")
	if err != nil {
		var pErr *os.PathError
		if errors.As(err, &pErr) {
			bch = BlockChain{{}}
			openTransactions = make([]Transaction, 0)
			return
		}
		panic(err)
	}
	defer f.Close()
	r := bufio.NewReader(f)

	chainLine, err := r.ReadSlice('\n')
	if err != nil {
		panic(err)
	}
	if err := json.Unmarshal(chainLine, &bch); err != nil {
		panic(err)
	}

	txLine, err := r.ReadSlice('\n')
	if err != nil {
		panic(err)
	}
	if err := json.Unmarshal(txLine, &openTransactions); err != nil {
		panic(err)
	}
}

func saveData() {
	f, err := os.OpenFile("blockchain.txt", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		panic(err)
	}
	defer func() {
		if err := f.Close(); err != nil {
			panic(err)
		}
	}()

	e := json.NewEncoder(f)
	if err := e.Encode(bch); err != nil {
		panic(err)
	}
	if err := e.Encode(openTransactions); err != nil {
		panic(err)
	}
}

func validProof(tx []Transaction, lastHash string, proof uint64) bool {
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
		for _, tx := range block.Transactions {
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

func getLastBlock(bch BlockChain) *Block {
	return &bch[len(bch)-1]
}

func verifyTransaction(tx Transaction) bool {
	return getBalance(tx.Sender) >= tx.Amount
}

func addTransaction(recipient string, amount float64) bool {
	return addTransactionWithSender(Owner, recipient, amount)
}

func addTransactionWithSender(sender, recipient string, amount float64) bool {
	tx := Transaction{
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

	rewardTx := Transaction{
		Sender:    "MINING",
		Recipient: Owner,
		Amount:    MiningReward,
	}

	copiedTransactions := make([]Transaction, len(openTransactions))
	copy(copiedTransactions, openTransactions)

	copiedTransactions = append(copiedTransactions, rewardTx)

	b := Block{
		PreviousHash: hashedBlock,
		Index:        int64(len(bch)),
		Transactions: copiedTransactions,
		Proof:        proof,
	}
	bch = append(bch, b)
}

func getTransactionValue() (string, float64) {
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

func verifyChain() bool {
	for i, b := range bch {
		if i == 0 {
			continue
		}

		if b.PreviousHash != bch[i-1].Hash() {
			return false
		}
		if !validProof(b.Transactions[:len(b.Transactions)-1], b.PreviousHash, b.Proof) {
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
	loadData()

	for {
		fmt.Println("Please choose")
		fmt.Println("1: Add a new Transaction value")
		fmt.Println("2: Mine a new Block")
		fmt.Println("3: Output the clockchain blocks")
		fmt.Println("4: Output participants")
		fmt.Println("5: Check Transaction validity")
		fmt.Println("h: Manipulate the chain")
		fmt.Println("q: Quit")

		switch getUserChoice() {
		case "1":
			txSender, txAmount := getTransactionValue()
			if !addTransaction(txSender, txAmount) {
				fmt.Println("Transaction failed!")
				break
			}
			saveData()
			fmt.Println("Added Transaction!")
		case "2":
			mineBlock()
			openTransactions = make([]Transaction, 0)
			saveData()
		case "3":
			fmt.Println(bch)
		case "4":
			fmt.Println(participants)
		case "5":
			if !verifyTransactions() {
				fmt.Println("There are invalid Transactions")
				break
			}
			fmt.Println("All Transactions are valid")
		case "h":
			if len(bch) < 1 {
				continue
			}
			bch[0] = Block{
				Transactions: []Transaction{
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
		fmt.Printf("Balance of %s: %6.2f\n", Owner, getBalance(Owner))

		if !verifyChain() {
			panic("chain is broken")
		}
	}
}
