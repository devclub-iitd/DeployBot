package controllers

import (
	"errors"
	"fmt"

	"github.com/devclub-iitd/DeployBot/src/history"
	"github.com/devclub-iitd/DeployBot/src/slack"
	log "github.com/sirupsen/logrus"
)

func redeploy(callbackID string, data map[string]interface{}) {
	log.Infof("service(%s) is being stopped", data["git_repo"].(string))
	stop(callbackID, data)
	channel := data["channel"].(string)
	state := history.GetState(data["git_repo"].(string))
	if state.Status == "stopped" {
		if _, ok := data["subdomain"]; !ok {
			if state.Subdomain == "" {
				log.Infof("Subdomain Has Not Been Set")
				var err = errors.New("Subdomain Has Not Been Set")
				slack.PostChatMessage(channel, fmt.Sprintf("Error: %s\n", err.Error()), nil)
				return
			}
			data["subdomain"] = state.Subdomain
		}
		if _, ok := data["access"]; !ok {
			if state.Access == "" {
				log.Infof("Access Has Not Been Set")
				var err = errors.New("Access Has Not Been Set")
				slack.PostChatMessage(channel, fmt.Sprintf("Error: %s\n", err.Error()), nil)
				return
			}
			data["access"] = state.Access
		}
		if _, ok := data["server_name"]; !ok {
			if state.Server == "" {
				log.Infof("Server Name Has Not Been Set")
				var err = errors.New("Server Name Has Not Been Set")
				slack.PostChatMessage(channel, fmt.Sprintf("Error: %s\n", err.Error()), nil)
				return
			}
			data["server_name"] = state.Server
		}
		deploy(callbackID, data)
	}
}
