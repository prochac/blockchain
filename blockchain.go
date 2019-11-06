package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
)

const MiningReward = 10

var participants = map[string]struct{}{}

type BlockChain struct {
	HostingNode      string
	chain            []Block
	openTransactions []Transaction
	peerNodes        []string // TODO transform to set (map[string]struct{})
}

func (b *BlockChain) Chain() []Block {
	cp := make([]Block, len(b.chain))
	copy(cp, b.chain)
	return cp
}

func (b *BlockChain) OpenTransactions() []Transaction {
	cp := make([]Transaction, len(b.openTransactions))
	copy(cp, b.openTransactions)
	return cp
}

func (b *BlockChain) LoadData() {
	f, err := os.Open("blockchain.txt")
	if err != nil {
		var pErr *os.PathError
		if errors.As(err, &pErr) {
			b.chain = []Block{{}}
			b.openTransactions = make([]Transaction, 0)
			return
		}
		panic(err)
	}
	defer f.Close()

	r := bufio.NewReader(f)

	var buf bytes.Buffer
	chainLine, err := r.ReadSlice('\n')
	for err == bufio.ErrBufferFull {
		buf.Write(chainLine)
		chainLine, err = r.ReadSlice('\n')
	}
	if err != nil {
		panic(err)
	}
	buf.Write(chainLine)
	if err := json.NewDecoder(&buf).Decode(&b.chain); err != nil {
		panic(err)
	}

	buf.Reset()
	txLine, err := r.ReadSlice('\n')
	for err == bufio.ErrBufferFull {
		buf.Write(txLine)
		txLine, err = r.ReadSlice('\n')
	}
	if err != nil {
		panic(err)
	}
	buf.Write(chainLine)
	if err := json.NewEncoder(&buf).Encode(&b.openTransactions); err != nil {
		panic(err)
	}

	buf.Reset()
	nodesLine, err := r.ReadSlice('\n')
	for err == bufio.ErrBufferFull {
		buf.Write(nodesLine)
		nodesLine, err = r.ReadSlice('\n')
	}
	if err != nil {
		panic(err)
	}
	buf.Write(nodesLine)

	fmt.Println(buf.String())

	var s []string
	if err := json.Unmarshal(buf.Bytes(), &s); err != nil {
		panic(err)
	}
	b.peerNodes = s
}

func (b BlockChain) SaveData() {
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
	if err := e.Encode(b.chain); err != nil {
		panic(err)
	}
	if err := e.Encode(b.openTransactions); err != nil {
		panic(err)
	}
	if err := e.Encode(b.peerNodes); err != nil {
		panic(err)
	}
}

func (b BlockChain) ProofOfWork() uint64 {
	lastHash := b.GetLastBlock().Hash()
	var proof uint64
	for !Verification.ValidProof(b.openTransactions, lastHash, proof) {
		proof++
	}
	return proof
}

func (b BlockChain) GetBalance() float64 {
	if b.HostingNode == "" {
		return -1
	}

	participant := b.HostingNode

	var (
		txSender     float64
		openTxSender float64
		txRecipient  float64
	)
	for _, block := range b.chain {
		for _, tx := range block.Transactions {
			if tx.Sender == participant {
				txSender += tx.Amount
			}
			if tx.Recipient == participant {
				txRecipient += tx.Amount
			}
		}
	}
	for _, tx := range b.openTransactions {
		if tx.Sender == participant {
			openTxSender += tx.Amount
		}
	}

	return txRecipient - (txSender + openTxSender)
}

func (b BlockChain) GetLastBlock() *Block {
	return &b.chain[len(b.chain)-1]
}

func (b *BlockChain) AddTransaction(recipient, sender string, signature string, amount float64) bool {
	if b.HostingNode == "" {
		return false
	}

	tx := Transaction{
		Sender:    sender,
		Recipient: recipient,
		Amount:    amount,
		Signature: signature,
	}

	if !Verification.VerifyTransaction(tx, b.GetBalance) {
		return false
	}

	b.openTransactions = append(b.openTransactions, tx)
	b.SaveData()
	return true
}

func (b *BlockChain) MineBlock() *Block {
	if b.HostingNode == "" {
		return nil
	}

	hashedBlock := b.GetLastBlock().Hash()
	proof := b.ProofOfWork()

	rewardTx := Transaction{
		Sender:    "MINING",
		Recipient: b.HostingNode,
		Amount:    MiningReward,
	}

	copiedTransactions := make([]Transaction, len(b.openTransactions))
	copy(copiedTransactions, b.openTransactions)

	for _, tx := range copiedTransactions {
		if !(Wallet{}).VerifyTransaction(tx) {
			return nil
		}
	}

	copiedTransactions = append(copiedTransactions, rewardTx)

	block := Block{
		PreviousHash: hashedBlock,
		Index:        int64(len(b.chain)),
		Transactions: copiedTransactions,
		Proof:        proof,
	}
	b.chain = append(b.chain, block)
	b.openTransactions = make([]Transaction, 0)
	b.SaveData()
	return &block
}

func (b *BlockChain) PeerNodes() []string {
	cp := make([]string, len(b.peerNodes))
	copy(cp, b.peerNodes)
	return cp
}

func (b *BlockChain) AddPeerNode(node string) {
	b.peerNodes = append(b.peerNodes, node)
	b.SaveData()
}

func (b *BlockChain) RemovePeerNode(node string) {
	filtered := b.peerNodes[:0]
	for _, x := range b.peerNodes {
		if x == node {
			filtered = append(filtered, x)
		}
	}

	b.peerNodes = filtered
	b.SaveData()
}
