package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
)

func HashString256(s string) string {
	h := sha256.Sum256([]byte(s))
	return hex.EncodeToString(h[:])
}

func (b Block) Hash() string {
	j, _ := json.Marshal(b)
	return HashString256(string(j))
}
