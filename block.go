package main

type Block struct {
	PreviousHash string        `json:"previous_hash"`
	Index        int64         `json:"index"`
	Transactions []Transaction `json:"transactions"`
	Proof        uint64        `json:"proof"`
}
