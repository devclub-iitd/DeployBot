package main

import (
	"crypto/hmac"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/hex"
	"hash"
	"os"
)

func getenv(key, fallback string) string {
	value := os.Getenv(key)
	if len(value) == 0 {
		return fallback
	}
	return value
}

func getHash(payload string, secret string, hashType string) string {

	var hashFn func() hash.Hash
	if hashType == "sha256" {
		hashFn = sha256.New
	} else if hashType == "sha1" {
		hashFn = sha1.New
	} else {
		panic("Invalid hash type passed in argument")
	}

	h := hmac.New(hashFn, []byte(secret))

	h.Write([]byte(payload))
	hash := hex.EncodeToString(h.Sum(nil))
	return hash
}
