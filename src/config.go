package main

import (
	"os"
	"path"

	log "github.com/sirupsen/logrus"
)

// DeployCount is the global number of deploy requests handled
var DeployCount = 0

// HooksScriptName is the name of script used to setup hooks
const (
	HooksScriptName  = "hooks.sh"
	DeployScriptName = "deploy.sh"
	DefaultBranch    = "master"
)

// LogDir is the place where all the logs are stored
var LogDir = getenv("LOG_DIR", "/var/logs/deploybot/")

// ServerURL is the URL at which the server is listening to requests
var ServerURL = getenv("SERVER_URL", "https://listen.devclub.in")

// Port is the HTTP Port on which the go code will listen
var Port = getenv("PORT", "7777")

// SlackAccessToken is the access token that is used to authenticate the app to
// communicate with slack
var SlackAccessToken = getenv("SLACK_ACCESS_TOKEN", "None")

// SlackSigningSecret is the secret key with which slack signs its request
var SlackSigningSecret = getenv("SLACK_SIGNING_SECRET", "None")

// GithubSecret is the secret used to verify that requests come from Github
var GithubSecret = getenv("GITHUB_SECRET", "None")

const (
	// OrganizationName is the name of the organization from which apps will be
	// deployed. This is the github organization which will be looked for events
	OrganizationName = "devclub-iitd"

	// GithubAPIURL is the url at which all APIs of github are rooted
	GithubAPIURL = "https://api.github.com"

	// SlackDialogURL is the url of the dialog API of slack
	SlackDialogURL = "https://api.slack.com/api/dialog.open"

	// SlackPostMessageURL is the url of the chat post message API of slack
	SlackPostMessageURL = "https://api.slack.com/api/chat.postMessage"
)

// SlackGeneralChannelID is the url of the chat post message API of slack
var SlackGeneralChannelID = getenv("SLACK_GENERAL_CHANNEL_ID", "C4Q43PCN5")

// Repo is the type that is used to store information about a repository
type Repo struct {
	Name string `json:"repo_name"`
	URL  string `json:"url"`
}

// Repositories is the list of repositories that we have on our git repo
var Repositories []Repo

// Server is the type storing the information about our server names and their
// IP
type Server struct {
	Name       string `json:"label"`
	DeployName string `json:"value"`
}

// ServersList is the list of all the servers that we have
var ServersList []Server

// ServerOptionsByte is the byte array of the options of server
var ServerOptionsByte []byte

// RepoOption is the type for the options of repo
type RepoOption struct {
	Name string `json:"label"`
	URL  string `json:"value"`
}

// RepoOptionsByte is the byte array of the options of repositories
var RepoOptionsByte []byte

func initialize() {

	log.SetFormatter(&log.TextFormatter{
		DisableTimestamp: false,
		FullTimestamp:    true,
		TimestampFormat:  "Mon Jan _2 15:04:05 2006",
	})

	// Output to stdout instead of the default stderr
	log.SetOutput(os.Stdout)

	if SlackAccessToken == "None" {
		log.Fatal("Slack Access Token is not present in environment, Exiting")
	}

	if SlackSigningSecret == "None" {
		log.Fatal("Slack Signing Secret is not present in environment, Exiting")
	}

	if GithubSecret == "None" {
		log.Fatal("Github Secret is not present in environment, Exiting")
	}

	log.Info("Setting Up Log directory")
	createDirIfNotExist(path.Join(LogDir, "deploy"))
	log.Infof("Log directory %s created", path.Join(LogDir, "deploy"))
	createDirIfNotExist(path.Join(LogDir, "git"))
	log.Infof("Log directory %s created", path.Join(LogDir, "git"))

	log.Info("Getting Servers Info")
	getServers()

}

// DialogMenu is the format of the menu that will displayed for deploy dialog
var DialogMenu = []byte(`{
	"callback_id": "deploy-xxxx",
	"title": "Deploy App",
	"submit_label": "Deploy",
	"elements": [
		{
			"type": "select",
			"label": "Github Repository",
			"name": "git_repo",
			"data_source": "external"
		},
		{
			"type": "select",
			"label": "Server Name",
			"name": "server_name",
			"data_source": "external"
		},
		{
			"label": "Subdomain",
			"name": "subdomain",
			"type": "text"
		},
		{
			"label": "APP Channel",
			"name": "channel",
			"type": "select",
			"data_source": "channels"
		}
	]
}`)
