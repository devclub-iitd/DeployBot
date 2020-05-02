package controllers

import (
	"github.com/devclub-iitd/DeployBot/src/history"
)

func redeploy(callbackID string, data map[string]interface{}) {
	stop(callbackID, data)
	state := history.GetState(data["git_repo"].(string))
	if state.Status == "stopped" {
		if !data["subdomain"] {
			data["subdomain"] = state.Subdomain
		}
		if !data["access"] {
			data["access"] = state.Access
		}
		if !data["server_name"] {
			data["server_name"] = state.Server
		}
		deploy(callbackID, data)
	}
}
