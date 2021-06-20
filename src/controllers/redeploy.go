package controllers

import (
	"github.com/devclub-iitd/DeployBot/src/history"
)

func redeploy(params *DeployAction) {
	state, _ := history.GetState(params.data["git_repo"].(string))

	params.data["subdomain"] = state.Subdomain
	params.data["access"] = state.Access
	params.data["server_name"] = state.Server

	if state.Status != "stopped" {
		params.command = "redeploy"
	}
	handleDeploy(params)
}
