package main

import (
	"encoding/json"
	"fmt"
	"net/http"
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
		json.NewEncoder(w).Encode(map[string]interface{}{
			"message": "Saving the keys failed.",
		})
		return
	}

	blockchain.HostingNode = wallet.PublicKey

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
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
		json.NewEncoder(w).Encode(map[string]interface{}{
			"message": "Loading the keys failed.",
		})
		return
	}

	blockchain.HostingNode = wallet.PublicKey

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
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
		json.NewEncoder(w).Encode(map[string]interface{}{
			"message":       "Loading balance failed.",
			"wallet_set_up": wallet.PublicKey != "",
		})
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Fetched balance successfully.",
		"funds":   balance,
	})
}

func getUI(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	http.ServeFile(w, r, "ui/node.html")
}

func addTransaction(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}
	if wallet.PublicKey == "" {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"message": "No wallet set up.",
		})
		return
	}

	var data struct {
		Recipient string
		Amount    float64
	}
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil || data.Recipient == "" || data.Amount == 0 {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"message": "Required data is missing.",
		})
		return
	}
	signature := wallet.SignTransaction(wallet.PublicKey, data.Recipient, data.Amount)

	if !blockchain.AddTransaction(data.Recipient, wallet.PublicKey, signature, data.Amount) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"message": "Creating a transaction failed.",
		})
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
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
		json.NewEncoder(w).Encode(map[string]interface{}{
			"message":       "Adding a block failed.",
			"wallet_set_up": wallet.PublicKey != "",
		})
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
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
	json.NewEncoder(w).Encode(transactions)
}

func getChain(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	chainSnapshot := blockchain.Chain()
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(chainSnapshot)
}

func main() {
	blockchain = BlockChain{HostingNode: wallet.PublicKey}
	blockchain.LoadData()

	http.HandleFunc("/", getUI)
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

	fmt.Println("* Running on http://0.0.0.0:5000/ (Press CTRL+C to quit)")
	http.ListenAndServe("0.0.0.0:5000", nil)
}
