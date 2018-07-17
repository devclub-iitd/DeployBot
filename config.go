package main

import "encoding/json"

// OrganizationName is the name of the organization from which apps will be
// deployed. This is the github organization which will be looked for events
var OrganizationName = getenv("ORG_NAME", "devclub-iitd")

// SlackAccessToken is the access token that is used to authenticate the app to
// communicate with slack
var SlackAccessToken = getenv("SLACK_ACCESS_TOKEN", "None")

// GithubAPIURL is the url at which all APIs of github are rooted
var GithubAPIURL = getenv("GITHUB_API_URL", "https://api.github.com")

// SlackDialogURL is the url of the dialog API of slack
var SlackDialogURL = getenv("SLACK_DIALOG_URL", "https://api.slack.com/api/dialog.open")

// Repo is the type that is used to store information about a repository
type Repo struct {
	Name string `json:"repo_name"`
	URL  string `json:"url"`
	// Branches are the branches of the git repo
	Branches []string `json:"branches"`
	// Maintainers array is the list of slack usernames of the people
	// responsible for that project
	Maintainers []string `json:"maintainers"`
}

// Repositories is the list of repositories that we have on our git repo
var Repositories []Repo

// Server is the type storing the information about our server names and their
// IP
type Server struct {
	Name string `json:"label"`
	IP   string `json:"value"`
}

// ServersList is the list of all the servers that we have
var ServersList = []Server{
	Server{"Baadal", "10.17.51.99"},
	Server{"AWS", "13.127.68.152"},
}

// ServerOptionsByte is the byte array of the options of server
var ServerOptionsByte []byte

// ReporOption is the type for the options of repo
type RepoOption struct {
	Name string `json:"label"`
	URL  string `json:"value"`
}

// RepoOptionsByte is the byte array of the options of repositories
var RepoOptionsByte []byte

func initialize() {
	if SlackAccessToken == "None" {
		panic("Slack Access Token is not present\nExiting\n")
	}

	getGitRepos()

	serverOptions := make(map[string]interface{})
	serverOptions["options"] = ServersList
	var err error
	ServerOptionsByte, err = json.Marshal(serverOptions)
	if err != nil {
		panic(err)
	}

}
