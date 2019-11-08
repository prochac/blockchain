package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"
)

const MiningReward = 10

var participants = map[string]struct{}{}

type BlockChain struct {
	NodeID           int
	PublicKey        string
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
	f, err := os.Open(fmt.Sprintf("blockchain-%d.txt", b.NodeID))
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

	var chainBuf bytes.Buffer
	chainLine, err := r.ReadSlice('\n')
	for err == bufio.ErrBufferFull {
		chainBuf.Write(chainLine)
		chainLine, err = r.ReadSlice('\n')
	}
	if err != nil {
		panic(err)
	}
	chainBuf.Write(chainLine)
	if err := json.NewDecoder(&chainBuf).Decode(&b.chain); err != nil {
		panic(err)
	}

	var txBuf bytes.Buffer
	txLine, err := r.ReadSlice('\n')
	for err == bufio.ErrBufferFull {
		txBuf.Write(txLine)
		txLine, err = r.ReadSlice('\n')
	}
	if err != nil {
		panic(err)
	}
	txBuf.Write(txLine)
	if err := json.NewDecoder(&txBuf).Decode(&b.openTransactions); err != nil {
		panic(err)
	}

	var nodesBuf bytes.Buffer
	nodesLine, err := r.ReadSlice('\n')
	for err == bufio.ErrBufferFull {
		nodesBuf.Write(nodesLine)
		nodesLine, err = r.ReadSlice('\n')
	}
	if err != nil {
		panic(err)
	}
	nodesBuf.Write(nodesLine)
	if err := json.NewDecoder(&nodesBuf).Decode(&b.peerNodes); err != nil {
		panic(err)
	}
}

func (b BlockChain) SaveData() {
	f, err := os.OpenFile(fmt.Sprintf("blockchain-%d.txt", b.NodeID), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
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
	if b.PublicKey == "" {
		return -1
	}
	return b.GetBalanceWithSender(b.PublicKey)
}

func (b BlockChain) GetBalanceWithSender(sender string) float64 {
	var (
		txSender     float64
		openTxSender float64
		txRecipient  float64
	)
	for _, block := range b.chain {
		for _, tx := range block.Transactions {
			if tx.Sender == sender {
				txSender += tx.Amount
			}
			if tx.Recipient == sender {
				txRecipient += tx.Amount
			}
		}
	}
	for _, tx := range b.openTransactions {
		if tx.Sender == sender {
			openTxSender += tx.Amount
		}
	}

	return txRecipient - (txSender + openTxSender)
}

func (b BlockChain) GetLastBlock() *Block {
	return &b.chain[len(b.chain)-1]
}

func (b *BlockChain) AddTransactionReceiving(recipient, sender string, signature string, amount float64) bool {
	if b.PublicKey == "" {
		return false
	}

	tx := Transaction{
		Sender:    sender,
		Recipient: recipient,
		Amount:    amount,
		Signature: signature,
	}

	if !Verification.VerifyTransaction(tx, b.GetBalanceWithSender) {
		return false
	}

	b.openTransactions = append(b.openTransactions, tx)
	b.SaveData()

	return true
}

func (b *BlockChain) RemoveTransaction(tx Transaction) {
	for i := range b.openTransactions {
		if b.openTransactions[i] == tx {
			// https://github.com/golang/go/wiki/SliceTricks#delete
			b.openTransactions = b.openTransactions[:i+copy(b.openTransactions[i:], b.openTransactions[i+1:])]
			return
		}
	}
}

func (b *BlockChain) AddTransaction(recipient, sender string, signature string, amount float64) bool {
	if !b.AddTransactionReceiving(recipient, sender, signature, amount) {
		return false
	}

	for _, node := range b.peerNodes {
		var buf bytes.Buffer
		_ = json.NewEncoder(&buf).Encode(map[string]interface{}{
			"sender":    sender,
			"recipient": recipient,
			"amount":    amount,
			"signature": signature,
		})
		resp, err := (&http.Client{Timeout: time.Second}).Post("http://"+node+"/broadcast-transaction", "application/json", &buf)
		if resp != nil {
			defer resp.Body.Close()
		}
		if err != nil {
			// TODO only connection error
			continue
		}
		if resp.StatusCode == 400 || resp.StatusCode == 500 {
			fmt.Println("Transaction declined, needs resolving")
			return false
		}
	}

	return true
}

func (b *BlockChain) MineBlock() *Block {
	if b.PublicKey == "" {
		return nil
	}

	hashedBlock := b.GetLastBlock().Hash()
	proof := b.ProofOfWork()

	rewardTx := Transaction{
		Sender:    "MINING",
		Recipient: b.PublicKey,
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

	for _, node := range b.peerNodes {
		var buf bytes.Buffer
		_ = json.NewEncoder(&buf).Encode(map[string]interface{}{
			"block": block,
		})
		resp, err := (&http.Client{Timeout: time.Second}).Post("http://"+node+"/broadcast-block", "application/json", &buf)
		if resp != nil {
			defer resp.Body.Close()
		}
		if err != nil {
			// TODO only connection error
			continue
		}
		if resp.StatusCode == 400 || resp.StatusCode == 500 {
			fmt.Println("Block declined, needs resolving")
			return nil
		}
	}

	return &block
}

func (b *BlockChain) AddBlock(block Block) bool {
	if !Verification.ValidProof(block.Transactions[:len(block.Transactions)-1], block.PreviousHash, block.Proof) {
		return false
	}

	if b.GetLastBlock().Hash() != block.PreviousHash {
		return false
	}
	b.chain = append(b.chain, block)

	storedTransactions := b.OpenTransactions()
	for _, tx := range block.Transactions {
		for _, openTx := range storedTransactions {
			if openTx == tx {
				blockchain.RemoveTransaction(openTx)
			}
		}
	}

	b.SaveData()
	return true
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
