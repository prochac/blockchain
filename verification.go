package main

import (
	"encoding/json"
	"fmt"
	"strconv"
)

type Verification struct{}

func validProof(tx []Transaction, lastHash string, proof uint64) bool {
	b, _ := json.Marshal(tx)
	guess := string(b) + lastHash + strconv.FormatUint(proof, 10)
	h := HashString256(guess)
	fmt.Println(h)
	return h[:2] == "00"
}

func verifyChain(bch BlockChain) bool {
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

func verifyTransaction(tx Transaction, getBalance func(string) float64) bool {
	return getBalance(tx.Sender) >= tx.Amount
}

func verifyTransactions(openTransactions []Transaction, getBalance func(string) float64) bool {
	for _, tx := range openTransactions {
		if !verifyTransaction(tx, getBalance) {
			return false
		}
	}
	return true
}
