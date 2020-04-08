package git

import (
	"log"

	"github.com/devclub-iitd/DeployBot/src/helper"
)

var (
	// githubSecret is the secret used to verify that requests come from Github
	githubSecret string
)

const (
	// organizationName is the name of the organization from which apps will be
	// deployed. This is the github organization which will be looked for events
	organizationName = "devclub-iitd"

	// apiURL is the url at which all APIs of github are rooted
	apiURL = "https://api.github.com"
)

// Repository is the type that is used to store information about a repository
type Repository struct {
	Name string `json:"repo_name"`
	URL  string `json:"url"`
}

func init() {
	githubSecret = helper.Env("GITHUB_SECRET", "None")
	if githubSecret == "None" {
		log.Fatal("GITHUB_SECRET is not present in environment, Exiting")
	}
}
