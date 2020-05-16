package controllers

import (
	"github.com/devclub-iitd/DeployBot/src/history"
)

func redeploy(callbackID string, data map[string]interface{}) {
	stop(callbackID, data)
	state, _ := history.GetState(data["git_repo"].(string))
	if state.Status == "stopped" {
		data["subdomain"] = state.Subdomain
		data["access"] = state.Access
		data["server_name"] = state.Server
		deploy(callbackID, data)
	}
}
