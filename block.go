package main

type Block struct {
	PreviousHash string
	Index        int64
	Transactions []Transaction
	Proof        uint64
}
