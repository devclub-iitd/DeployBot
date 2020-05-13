package discord

import (
	"fmt"

	"github.com/devclub-iitd/DeployBot/src/helper"
	log "github.com/sirupsen/logrus"
)

var (
	// postMessageHookURL is the webhook url for posting messages on discord
	postMessageHookURL string
	serverURL string
)

func init() {
	postMessageHookURL = helper.Env("DISCORD_MESSAGE_WEBHOOK", "None")
	if postMessageHookURL == "None" {
		log.Fatal("DISCORD_MESSAGE_WEBHOOK not present in environment, Exiting")
	}
	serverURL = helper.Env("SERVER_URL", "https://listen.devclub.iitd.ac.in")
}

type Embed struct {
	Title  string        `json:"title"`
	Color  int           `json:"color"`
	Fields []interface{} `json:"fields"`
}

type Message struct {
	Content string   `json:"content"`
	Embeds  []*Embed `json:"embeds"`
}

func newMessage(callbackID, action, result string, fields []interface{}) Message {
	var callbackField interface{}
	callbackField = map[string]interface{}{
		"name": "Action ID",
		"value": callbackID,
		"inline": true,
	}
	fields = append([]interface{}{callbackField}, fields...)
	embed := &Embed{Fields: fields}
	switch result {
	case "success":
		embed.Title = fmt.Sprintf("%s action completed", action)
		embed.Color = 6749952
	case "failed":
		embed.Title = fmt.Sprintf("%s action failed", action)
		embed.Color = 16724736
	default:
		embed.Title = fmt.Sprintf("%s action in progress", action)
		embed.Color = 3381759
	}

	return Message{
		Embeds: []*Embed{embed},
	}
}
