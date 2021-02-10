package controllers

import (
	"github.com/devclub-iitd/DeployBot/src/history"
)

func redeploy(callbackID string, data map[string]interface{}) {
	// stop(callbackID, data)
	state, _ := history.GetState(data["git_repo"].(string))

	data["subdomain"] = state.Subdomain
	data["access"] = state.Access
	data["server_name"] = state.Server
	data["redeploy"] = false

	if state.Status != "stopped" {
		data["redeploy"] = true
	}

	deploy(callbackID, data)
}
