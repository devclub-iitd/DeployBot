package git

import (
	"log"

	"github.com/devclub-iitd/DeployBot/src/helper"
)

var (
	// githubSecret is the secret used to verify that requests come from Github
	githubSecret string

	// githubAccessToken is the OAuth token used to get details of GitHub
	// private repositories
	githubAccessToken string

	// githubUserName is the name of the user which is a member of the GitHub
	//organization. githubAccessToken is a personal access token of this user
	githubUserName string

	// githubActionToken is used to verify the request coming from GitHub Actions
	githubActionToken string

	// Logdir is the directory where all logs will be stored
	logDir string
)

const (
	// organizationName is the name of the organization from which apps will be
	// deployed. This is the github organization which will be looked for events
	organizationName = "devclub-iitd"

	// apiURL is the url at which all APIs of github are rooted
	apiURL = "https://api.github.com"

	// root path of pre-commit files
	BASE_FILES_PATH   string = "../../precommit/"
	// pre-commit config file
	PRE_COMMIT_CONFIG string = "../../precommit/CI/pre-commit-config.yaml"

	// CI script name
	ciScriptName = "ci.sh"
)

// Repository is the type that is used to store information about a repository
type Repository struct {
	Name     string   `json:"repo_name"`
	URL      string   `json:"url"`
	Branches []string `json:"branches"`
}

// CIAction : Used for unmarshalling the CI request
type CIAction struct {
	Repo    string `json:"repo"`
	Python  string `json:"python"`
	Golang  string `json:"golang"`
	Node    string `json:"node"`
	Ts      string `json:"ts"`
	Flutter string `json:"flutter"`
	Dart    string `json:"dart"`
	Docker  string `json:"docker"`
	Shell   string `json:"shell"`
}

func init() {
	githubSecret = helper.Env("GITHUB_SECRET", "None")
	if githubSecret == "None" {
		log.Fatal("GITHUB_SECRET is not present in environment, Exiting")
	}

	githubUserName = helper.Env("GITHUB_USERNAME", "None")
	if githubUserName == "None" {
		log.Fatal("GITHUB_USERNAME is not present in environment, Exiting")
	}

	githubAccessToken = helper.Env("GITHUB_ACCESS_TOKEN", "None")
	if githubAccessToken == "None" {
		log.Fatal("GITHUB_ACCESS_TOKEN is not present in environment, Exiting")
	}

	githubActionToken = helper.Env("GITHUB_ACTION_TOKEN", "None")
	if githubActionToken == "None" {
		log.Fatal("GITHUB_ACTION_TOKEN is not present in environment, Exiting")
	}

	logDir = helper.Env("LOG_DIR", "/var/logs/deploybot/")
}
