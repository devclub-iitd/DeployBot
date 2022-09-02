// Package helper contains various helper functions.
package helper

import (
	"crypto/hmac"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"hash"
	"os"
	fp "path/filepath"
	"strings"
)

const (
	urlPrefix = "git@github.com:devclub-iitd/"
	delim     = ":"
	suffix    = ".git"
)

// DeployCount is the global number of deploy requests handled
var (
	DeployCount   int32 = 0
	StopCount     int32 = 0
	LogsCount     int32 = 0
	RedeployCount int32 = 0
)

// Env returns the value of the environment variable if present, or returns the fallback value
func Env(key, fallback string) string {
	value := os.Getenv(key)
	if len(value) == 0 {
		return fallback
	}
	return value
}

// Hash hashes a payload, with a secret with a given hash type (sha1 or sha256)
func Hash(payload, secret, hashType string) (string, error) {
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

// WriteToFile writes a given string to a file. It creates the file if it is not present
func WriteToFile(filePath, text string) error {
	CreateDirIfNotExist(fp.Dir(filePath))
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

// CreateDirIfNotExist creates a given directory if it does not exist
func CreateDirIfNotExist(dir string) error {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err1 := os.MkdirAll(dir, 0755); err1 != nil {
			return fmt.Errorf("cannot create directory - %v", err1)
		}
	}
	return nil
}

// FileExists returns true if the file exists
func FileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func SerializeRepo(repoName string, branch string) string {
	return fmt.Sprintf("%s%s%s", repoName, delim, branch)
}

func DeserializeRepo(gitRepo string) (repoURL string, branch string, completeUrl string) {
	gitRepoSplit := strings.Split(gitRepo, delim)
	repoURL = urlPrefix + strings.Join(gitRepoSplit[:len(gitRepoSplit)-1], delim) + suffix
	branch = gitRepoSplit[len(gitRepoSplit)-1]
	completeUrl = fmt.Sprintf("%s%s%s", repoURL, delim, branch)
	return repoURL, branch, completeUrl
}
