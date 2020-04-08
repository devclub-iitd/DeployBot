package slack

import (
	"log"

	"github.com/devclub-iitd/DeployBot/src/helper"
)

var (
	// accessToken is the access token that is used to authenticate the app to communicate with slack
	accessToken string
	// signingSecret is the secret key with which slack signs its request
	signingSecret string
	// generalChannelID is the channel id of the general channel on slack
	generalChannelID string
	// AllHooksChannelID is the channel id where all the messages will be posted by default
	AllHooksChannelID string
)

const (
	// dialogURL is the url of the dialog API of slack
	dialogURL = "https://api.slack.com/api/dialog.open"
	// postMessageURL is the url of the chat post message API of slack
	chatPostMessageURL = "https://api.slack.com/api/chat.postMessage"
)

func init() {
	accessToken = helper.Env("SLACK_ACCESS_TOKEN", "None")
	signingSecret = helper.Env("SLACK_SIGNING_SECRET", "None")
	if accessToken == "None" {
		log.Fatal("SLACK_ACCESS_TOKEN is not present in environment, Exiting")
	}
	if signingSecret == "None" {
		log.Fatal("SLACK_SIGNING_SECRET is not present in environment, Exiting")
	}
	generalChannelID = helper.Env("SLACK_GENERAL_CHANNEL_ID", "C4Q43PCN5")
	AllHooksChannelID = helper.Env("SLACK_ALL_HOOKS_CHANNEL_ID", "CGN56SGDS")
}

type dialog struct {
	name    string
	content []byte
}

// deployDialog is the format of the menu that will displayed for deploying a service
var deployDialog = dialog{
	name: "deploy",
	content: []byte(`{
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

// stopDialog is the format of the menu that will displayed for stopping a service
var stopDialog = dialog{
	name: "stop",
	content: []byte(`{
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

// logsDialog is the format of the menu that will displayed for fetching logs
var logsDialog = dialog{
	name: "logs",
	content: []byte(`{
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
