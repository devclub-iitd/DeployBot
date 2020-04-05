package main

import (
	"crypto/hmac"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
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

func getHash(payload, secret, hashType string) (string, error) {
	var hashFn func() hash.Hash
	switch hashType {
	case "sha256":
		hashFn = sha256.New
	case "sha1":
		hashFn = sha1.New
	default:
		return "", fmt.Errorf("invalid hashType specified")
	}
	h := hmac.New(hashFn, []byte(secret))
	h.Write([]byte(payload))
	hash := hex.EncodeToString(h.Sum(nil))
	return hash, nil
}

func writeToFile(filePath, text string) error {
	f, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("cannot create file - %v", err)
	}
	defer f.Close()
	_, err = f.WriteString(text + "\n")
	if err != nil {
		return fmt.Errorf("cannot write to file - %v", err)
	}
	f.Sync()
	return nil
}

func createDirIfNotExist(dir string) error {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err1 := os.MkdirAll(dir, 0755); err1 != nil {
			return fmt.Errorf("cannot create directory - %v", err1)
		}
	}
	return nil
}
