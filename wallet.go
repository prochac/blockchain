package main

import (
	"bufio"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/asn1"
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"strings"
)

type Wallet struct {
	NodeID     int
	PrivateKey string `json:"private_key"`
	PublicKey  string `json:"public_key"`
}

func (w *Wallet) CreateKeys() {
	reader := rand.Reader
	privateKey, err := rsa.GenerateKey(reader, 1024)
	if err != nil {
		panic(err)
	}

	publicKey, err := asn1.Marshal(privateKey.PublicKey)
	if err != nil {
		panic(err)
	}

	w.PrivateKey = base64.StdEncoding.EncodeToString(x509.MarshalPKCS1PrivateKey(privateKey))
	w.PublicKey = base64.StdEncoding.EncodeToString(publicKey)
}

func (w *Wallet) SaveKeys() bool {
	if w.PublicKey == "" || w.PrivateKey == "" {
		return false
	}

	f, err := os.OpenFile(fmt.Sprintf("wallet-%d.txt", w.NodeID), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		panic(err)
	}
	defer func() {
		if err := f.Close(); err != nil {
			panic(err)
		}
	}()
	if _, err := f.WriteString(w.PublicKey); err != nil {
		panic(err)
	}
	if _, err := f.WriteString("\n"); err != nil {
		panic(err)
	}
	if _, err := f.WriteString(w.PrivateKey); err != nil {
		panic(err)
	}

	return true
}

func (w *Wallet) LoadKeys() bool {
	f, err := os.Open(fmt.Sprintf("wallet-%d.txt", w.NodeID))
	if err != nil {
		panic(err)
	}
	defer f.Close()

	r := bufio.NewReader(f)

	publicKey, err := r.ReadString('\n')
	if err != nil {
		panic(err)
	}
	publicKey = strings.TrimSuffix(publicKey, "\n")

	privateKey, err := r.ReadString('\n')
	if err != nil && err != io.EOF {
		panic(err)
	}
	privateKey = strings.TrimSuffix(privateKey, "\n")

	w.PrivateKey = privateKey
	w.PublicKey = publicKey

	return true
}

func (w *Wallet) SignTransaction(sender, recipient string, amount float64) string {
	privateKey, err := base64.StdEncoding.DecodeString(w.PrivateKey)
	if err != nil {
		panic(err)
	}

	signer, err := x509.ParsePKCS1PrivateKey(privateKey)
	if err != nil {
		panic(err)
	}

	hash := sha256.Sum256([]byte(fmt.Sprintf("%s%s%f", sender, recipient, amount)))
	signature, err := rsa.SignPKCS1v15(rand.Reader, signer, crypto.SHA256, hash[:])
	if err != nil {
		panic(err)
	}

	return base64.StdEncoding.EncodeToString(signature)
}

func (w Wallet) VerifyTransaction(transaction Transaction) bool {
	publicKey, err := base64.StdEncoding.DecodeString(transaction.Sender)
	if err != nil {
		panic(err)
	}

	verifier, err := x509.ParsePKCS1PublicKey(publicKey)
	if err != nil {
		panic(err)
	}

	signature, err := base64.StdEncoding.DecodeString(transaction.Signature)
	if err != nil {
		panic(err)
	}

	hash := sha256.Sum256([]byte(fmt.Sprintf("%s%s%f", transaction.Sender, transaction.Recipient, transaction.Amount)))
	return rsa.VerifyPKCS1v15(verifier, crypto.SHA256, hash[:], signature) == nil
}
