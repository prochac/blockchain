package main

import (
	"fmt"
	"os"
)

type blockchain []block

var bch blockchain

type block struct {
	previous *block
	Value    interface{}
}

func (b block) String() string {
	return fmt.Sprintf("{%s %v}", b.previous, b.Value)
}

func getLastBlock(bch blockchain) *block {
	if len(bch) == 0 {
		return nil
	}
	return &bch[len(bch)-1]
}

func addTransaction(val interface{}, prev *block) {
	bch = append(bch, block{
		previous: prev,
		Value:    val,
	})
}

func getUserChoice() string {
	var s string
	if _, err := fmt.Scan(&s); err != nil {
		panic(err)
	}
	return s
}

func getTransactionValue() float64 {
	fmt.Print("Input number: ")
	var f float64
	if _, err := fmt.Scanf("%f", &f); err != nil {
		panic(err)
	}
	return f
}

func verifyChain() bool {
	for i, b := range bch {
		if i == 0 {
			continue
		}

		if *b.previous != bch[i-1] {
			return false
		}
	}
	return true
}

func main() {
	for {
		fmt.Println("Please choose")
		fmt.Println("1: Add a new transaction value")
		fmt.Println("2: Output the clockchain blocks")
		fmt.Println("h: Manipulate the chain")
		fmt.Println("q: Quit")

		switch getUserChoice() {
		case "1":
			txAmount := getTransactionValue()
			addTransaction(txAmount, getLastBlock(bch))
		case "2":
			fmt.Println(bch)
		case "h":
			if len(bch) < 1 {
				continue
			}
			bch[0] = block{Value: 42}
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
