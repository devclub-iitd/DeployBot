package controllers

import (
	"github.com/devclub-iitd/DeployBot/src/history"
)

func redeploy(callbackID string, data map[string]interface{}) {
	stop(callbackID, data)
	state := history.GetState(data["git_repo"].(string))
	if state.Status == "stopped" {
		if _, ok := data["subdomain"]; !ok && state.Subdomain != "" {
			data["subdomain"] = state.Subdomain
		}
		if _, ok := data["access"]; !ok && state.Access != "" {
			data["access"] = state.Access
		}
		if _, ok := data["server_name"]; !ok && state.Server != "" {
			data["server_name"] = state.Server
		}
		deploy(callbackID, data)
	}
}
