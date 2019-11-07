package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"strings"
)

var (
	wallet     Wallet
	blockchain BlockChain
)

func createKeys(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	wallet.CreateKeys()
	if !wallet.SaveKeys() {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"message": "Saving the keys failed.",
		})
		return
	}

	blockchain.PublicKey = wallet.PublicKey

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"public_key":  wallet.PublicKey,
		"private_key": wallet.PrivateKey,
		"funds":       blockchain.GetBalance(),
	})
}

func loadKeys(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	if !wallet.LoadKeys() {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"message": "Loading the keys failed.",
		})
		return
	}

	blockchain.PublicKey = wallet.PublicKey

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"public_key":  wallet.PublicKey,
		"private_key": wallet.PrivateKey,
		"funds":       blockchain.GetBalance(),
	})
}

func getBalance(w http.ResponseWriter, r *http.Request) {
	balance := blockchain.GetBalance()
	if balance < 0 {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"message":       "Loading balance failed.",
			"wallet_set_up": wallet.PublicKey != "",
		})
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Fetched balance successfully.",
		"funds":   balance,
	})
}

func getNodeUI(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	http.ServeFile(w, r, "ui/node.html")
}

func getNetworkUI(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	http.ServeFile(w, r, "ui/network.html")
}

func broadcastTransaction(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	var tx Transaction
	if json.NewDecoder(r.Body).Decode(&tx) != nil || tx.Sender == "" || tx.Recipient == "" || tx.Amount == 0 || tx.Signature == "" {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"message": "Some tx is missing.",
		})
		return
	}

	if !blockchain.AddTransactionReceiving(tx.Recipient, tx.Sender, tx.Signature, tx.Amount) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"message": "Creating a transaction failed.",
		})
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"message":     "Successfully added transaction.",
		"transaction": tx,
	})
}

func broadcastBlock(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	var data struct {
		Block *Block `json:"block"`
	}
	if json.NewDecoder(r.Body).Decode(&data) != nil || data.Block == nil {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"message": "Some tx is missing.",
		})
		return
	}
	block := *data.Block
	if block.Index < blockchain.GetLastBlock().Index+1 {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusConflict)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"message": "Blockchain seems to be shorter, block not added.",
		})
		return
	}

	if block.Index > blockchain.GetLastBlock().Index+1 {
		// too much new
	}
	//if block.Index == blockchain.GetLastBlock().Index+1
	if !blockchain.AddBlock(block) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"message": "Block seems invalid.",
		})
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Block added.",
	})
}

func addTransaction(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	if wallet.PublicKey == "" {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"message": "No wallet set up.",
		})
		return
	}

	var data struct {
		Recipient string  `json:"recipient"`
		Amount    float64 `json:"amount"`
	}
	if json.NewDecoder(r.Body).Decode(&data) != nil || data.Recipient == "" || data.Amount == 0 {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"message": "Required data is missing.",
		})
		return
	}
	signature := wallet.SignTransaction(wallet.PublicKey, data.Recipient, data.Amount)

	if !blockchain.AddTransaction(data.Recipient, wallet.PublicKey, signature, data.Amount) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"message": "Creating a transaction failed.",
		})
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Successfully added transaction.",
		"transaction": Transaction{
			Sender:    wallet.PublicKey,
			Recipient: data.Recipient,
			Amount:    data.Amount,
			Signature: signature,
		},
		"funds": blockchain.GetBalance(),
	})
}

func mine(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	block := blockchain.MineBlock()
	if block == nil {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"message":       "Adding a block failed.",
			"wallet_set_up": wallet.PublicKey != "",
		})
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Block added successfully.",
		"block":   block,
		"funds":   blockchain.GetBalance(),
	})
}

func getTransactions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	transactions := blockchain.OpenTransactions()

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(transactions)
}

func getChain(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	chainSnapshot := blockchain.Chain()
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(chainSnapshot)
}

func addNode(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	var data struct {
		Node string `json:"node"`
	}
	if json.NewDecoder(r.Body).Decode(&data) != nil || data.Node == "" {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"message": "No data attached.",
		})
		return
	}

	blockchain.AddPeerNode(data.Node)

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"message":   "Node added successfully.",
		"all_nodes": blockchain.PeerNodes(),
	})
}

func removeNode(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	nodeURL := strings.SplitN(strings.TrimPrefix(r.URL.Path, "/nodeURL/"), "/", 1)[0]
	if nodeURL == "" {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"message": "No nodeURL found.",
		})
		return
	}

	blockchain.RemovePeerNode(nodeURL)

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"message":   "Node added successfully.",
		"all_nodes": blockchain.PeerNodes(),
	})
}

func getNode(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"all_nodes": blockchain.PeerNodes(),
	})
}

func main() {
	var port int
	flag.IntVar(&port, "port", 5000, "")
	flag.IntVar(&port, "p", 5000, "")
	flag.Parse()

	wallet.NodeID = port
	blockchain = BlockChain{PublicKey: wallet.PublicKey, NodeID: port}
	blockchain.LoadData()

	http.HandleFunc("/", getNodeUI)
	http.HandleFunc("/network", getNetworkUI)
	http.HandleFunc("/mine", mine)
	http.HandleFunc("/chain", getChain)
	http.HandleFunc("/wallet", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			loadKeys(w, r)
		case http.MethodPost:
			createKeys(w, r)
		default:
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
			return
		}
	})
	http.HandleFunc("/transaction", addTransaction)
	http.HandleFunc("/transactions", getTransactions)
	http.HandleFunc("/balance", getBalance)
	http.HandleFunc("/nodes", getNode)
	http.HandleFunc("/node", addNode)
	http.HandleFunc("/node/", removeNode)
	http.HandleFunc("/broadcast-transaction", broadcastTransaction)
	http.HandleFunc("/broadcast-block", broadcastBlock)

	addr := fmt.Sprintf("0.0.0.0:%d", port)
	fmt.Printf("* Running on http://%s/ (Press CTRL+C to quit)\n", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}
