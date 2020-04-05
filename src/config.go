package main

import (
	"os"
	"path"
	"time"

	log "github.com/sirupsen/logrus"
)

// DeployCount is the global number of deploy requests handled
var (
	DeployCount int32 = 0
	StopCount   int32 = 0
	LogsCount   int32 = 0
)

// HooksScriptName is the name of script used to setup hooks
const (
	HooksScriptName  = "hooks.sh"
	DeployScriptName = "deploy.sh"
	StopScriptName   = "stop.sh"
	LogScriptName    = "logs.sh"
	DefaultBranch    = "master"
)

// All the deploy and stop requests are logged in historyFile
// templateFile is html template for viewing running services
const (
	historyFile  = "/etc/nginx/history.json"
	templateFile = "status_template.html"
)

// Type for logging one deploy or stop request
type ActionInstance struct {
	Action    string    `json:"action"`
	Subdomain string    `json:"subdomain"`
	Server    string    `json:"server"`
	Access    string    `json:"access"`
	Result    string    `json:"result"`
	Timestamp time.Time `json:"timestamp"`
}

// Type for storing current state of service
type State struct {
	Status    string `json:"status"`
	Subdomain string `json:"subdomain"`
	Access    string `json:"access"`
	Server    string `json:"server"`
}

type Service struct {
	Actions []ActionInstance `json:"actions"`
	Current State            `json:"current"`
}

// LogDir is the place where all the logs are stored
var LogDir = getenv("LOG_DIR", "/var/logs/deploybot/")

// ServerURL is the URL at which the server is listening to requests
var ServerURL = getenv("SERVER_URL", "https://listen.devclub.iitd.ac.in")

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

// SlackAllHooksChanneID is the channel id where all the messages will be posted by default
var SlackAllHooksChannelID = getenv("SLACK_ALL_HOOKS_CHANNEL_ID", "CGN56SGDS")

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
	createDirIfNotExist(path.Join(LogDir, "service"))
	log.Infof("Log directory %s created", path.Join(LogDir, "service"))

	log.Info("Getting Servers Info")
	getServers()

}

type Dialog struct {
	Name    string
	Content []byte
}

// DeployDialog is the format of the menu that will displayed for deploying a service
var DeployDialog = Dialog{
	Name: "deploy",
	Content: []byte(`{
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
        "label": "Access",
        "name": "access",
        "type": "select",
        "options": [
              {
                  "label": "Internal",
                  "value": "internal"
              },
              {
                  "label": "External",
                  "value": "external"
              }
        ]
      },
      {
        "label": "APP Channel",
        "name": "channel",
        "type": "select",
        "data_source": "channels",
        "value": "CGN56SGDS"
      }
    ]
  }`),
}

// StopDialog is the format of the menu that will displayed for stopping a service
var StopDialog = Dialog{
	Name: "stop",
	Content: []byte(`{
    "callback_id": "stop-xxxx",
    "title": "Stop App",
    "submit_label": "Stop",
    "elements": [
      {
        "type": "select",
        "label": "Github Repository",
        "name": "git_repo",
        "data_source": "external"
      },
      {
        "label": "APP Channel",
        "name": "channel",
        "type": "select",
        "data_source": "channels",
        "value": "CGN56SGDS"
      }
    ]
  }`),
}

// LogsDialog is the format of the menu that will displayed for fetching logs
var LogsDialog = Dialog{
	Name: "logs",
	Content: []byte(`{
    "callback_id": "logs-xxxx",
    "title": "Fetch Logs",
    "submit_label": "Fetch",
    "elements": [
      {
        "type": "select",
        "label": "Github Repository",
        "name": "git_repo",
        "data_source": "external"
      },
      {
        "type": "text",
        "subtype": "number",
        "label": "Number of entries",
        "name": "tail_count",
        "placeholder": "\"all\" or number"
      },
      {
        "label": "APP Channel",
        "name": "channel",
        "type": "select",
        "data_source": "channels",
        "value": "CGN56SGDS"
      }
    ]
  }`),
}
