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

func getUI(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	fmt.Fprint(w, "This Works!")
}

func mine(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	block := blockchain.MineBlock()
	if block == nil {
		w.WriteHeader(500)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"message":       "Adding a block failed.",
			"wallet_set_up": wallet.PublicKey != "",
		})
		return
	}

	w.WriteHeader(201)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Block added successfully.",
		"block":   block,
	})
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
	wallet.CreateKeys()
	blockchain = BlockChain{HostingNode: wallet.PublicKey}
	blockchain.LoadData()

	http.HandleFunc("/", getUI)
	http.HandleFunc("/mine", mine)
	http.HandleFunc("/chain", getChain)

	fmt.Println("* Running on http://0.0.0.0:5000/ (Press CTRL+C to quit)")
	http.ListenAndServe("0.0.0.0:5000", nil)
}
