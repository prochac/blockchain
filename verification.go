package main

import (
	"encoding/json"
	"fmt"
	"strconv"
)

var Verification struct {
	ValidProof         func(tx []Transaction, lastHash string, proof uint64) bool
	VerifyChain        func(bch BlockChain) bool
	VerifyTransaction  func(tx Transaction, getBalance func() float64) bool
	VerifyTransactions func(openTransactions []Transaction, getBalance func() float64) bool
}

func init() {
	Verification.ValidProof = func(tx []Transaction, lastHash string, proof uint64) bool {
		b, _ := json.Marshal(tx)
		guess := string(b) + lastHash + strconv.FormatUint(proof, 10)
		h := HashString256(guess)
		fmt.Println(h)
		return h[:2] == "00"
	}
	Verification.VerifyChain = func(bch BlockChain) bool {
		for i, b := range bch.Chain() {
			if i == 0 {
				continue
			}

			if b.PreviousHash != bch.Chain()[i-1].Hash() {
				return false
			}
			if !Verification.ValidProof(b.Transactions[:len(b.Transactions)-1], b.PreviousHash, b.Proof) {
				fmt.Println("Proof of work is invalid")
				return false
			}
		}
		return true
	}
	Verification.VerifyTransaction = func(tx Transaction, getBalance func() float64) bool {
		return getBalance() >= tx.Amount
	}
	Verification.VerifyTransactions = func(openTransactions []Transaction, getBalance func() float64) bool {
		for _, tx := range openTransactions {
			if !Verification.VerifyTransaction(tx, getBalance) {
				return false
			}
		}
		return true
	}
}
