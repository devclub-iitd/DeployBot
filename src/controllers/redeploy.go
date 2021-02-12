package controllers

import (
	"github.com/devclub-iitd/DeployBot/src/history"
)

func redeploy(callbackID string, data map[string]interface{}) {
	// stop(callbackID, data)
	state, _ := history.GetState(data["git_repo"].(string))
	params := DeployAction{"deploy", callbackID, make(map[string]interface{})}

	params.data["subdomain"] = state.Subdomain
	params.data["access"] = state.Access
	params.data["server_name"] = state.Server

	if state.Status != "stopped" {
		params.command = "redeploy"
	}
	handleDeploy(&params)
}
