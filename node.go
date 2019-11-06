package main

import (
	"encoding/json"
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

	blockchain.HostingNode = wallet.PublicKey

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

	blockchain.HostingNode = wallet.PublicKey

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
	blockchain = BlockChain{HostingNode: wallet.PublicKey}
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

	fmt.Println("* Running on http://0.0.0.0:5000/ (Press CTRL+C to quit)")
	log.Fatal(http.ListenAndServe("0.0.0.0:5000", nil))
}
