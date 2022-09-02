package controllers

import (
	"errors"
	"fmt"
	"os/exec"
	"path"

	"github.com/devclub-iitd/DeployBot/src/discord"
	"github.com/devclub-iitd/DeployBot/src/helper"
	"github.com/devclub-iitd/DeployBot/src/history"
	"github.com/devclub-iitd/DeployBot/src/slack"
	log "github.com/sirupsen/logrus"
)

// deploy deploys a given slack request using the deploy.sh
func deploy(params *deployAction) {
	channel := params.data["channel"].(string)
	log.Infof("Deploy params data: %v\n", params.data)
	actionLog := history.NewAction(params.command, params.data)
	if err := slack.PostChatMessage(channel, actionLog.String(), nil); err != nil {
		log.Errorf("cannot post begin deployment chat message - %v", err)
		return
	}
	go discord.PostActionMessage(params.callbackID, actionLog.EmbedFields())
	log.Infof("beginning %s with callback_id as %s", actionLog, params.callbackID)

	logPath := fmt.Sprintf("%s/%s.txt", params.command, params.callbackID)

	output, err := internalDeploy(actionLog)

	writeErr := helper.WriteToFile(path.Join(logDir, logPath), string(output))
	if writeErr != nil {
		log.Errorf("An error occured while writing to %s: %s", path.Join(logDir, logPath), writeErr.Error())
		slack.PostChatMessage(channel, fmt.Sprintf("%s\nCould not write to %s\nerror: %s\n", actionLog, path.Join(logDir, logPath), writeErr.Error()), nil)
	}
	actionLog.LogPath = logPath
	if err != nil {
		actionLog.Result = "failed"
		log.Errorf("ActionLog: %v\nOutput: %s\nERROR: %s", actionLog, string(output), err.Error())
		history.StoreAction(actionLog)
		slack.PostChatMessage(channel, fmt.Sprintf("%s\nError: %s\n", actionLog, err.Error()), nil)
	} else {
		actionLog.Result = "success"
		log.Info(actionLog.String())
		history.StoreAction(actionLog)
		slack.PostChatMessage(channel, fmt.Sprintf("%s\n", actionLog), nil)
	}
	go discord.PostActionMessage(params.callbackID, actionLog.EmbedFields())
}

// internalDeploy deploys the given app on the server specified.
func internalDeploy(a *history.ActionInstance) ([]byte, error) {
	branch := defaultBranch
	if a.Branch != "" {
		branch = a.Branch
	}

	url := a.CompleteURL

	// This is a value, and thus modifying it does not change the original state in the history map
	state, tag := history.GetState(url)

	var output []byte
	var err error
	switch state.Status {
	case "running":
		log.Infof("service(%s) is already running", url)
		output = []byte("Service is already running!")
		err = errors.New("already running")
	case "stopping":
		log.Infof("service(%s) is stopping. Can't deploy.", url)
		output = []byte("Service is stopping. Please wait for the process to be completed and try again.")
		err = errors.New("cannot deploy while service is stopping")
	case "deploying":
	case "redeploying":
		log.Infof("service(%s) is being deployed", url)
		output = []byte("Service is being deployed. Cannot start another deploy instance.")
		err = errors.New("already deploying")
	// Assume that, either the service is stopped or does not exist, which means we can deploy.
	default:
		log.Infof("calling %s to deploy %s on %s", deployScriptName, url, a.Server)
		state.Subdomain = a.Subdomain
		state.Access = a.Access
		state.Server = a.Server
		state.Status = "deploying"
		tag, err1 := history.SetState(url, tag, state)
		if err1 != nil {
			log.Infof("setting state to deploying failed - %v", err1)
			output = []byte("InternalDeployError: cannot set state to deploying - " + err1.Error())
			return output, err1
		}

		args := getDeployArgs(a.RepoURL, branch, a.Server, a.Subdomain, a.Access, make(map[string]bool))
		log.Infof("args: %v", args)
		output, err = exec.Command(deployScriptName, args...).CombinedOutput()
		if err != nil {
			state.Status = "stopped"
		} else {
			state.Status = "running"
		}

		// There should be no error here, ever. Checking it to make sure
		// TODO: On error, set state to an "error" state which only stop should be able to modify
		tag, err1 = history.SetState(url, tag, state)
		for ; err1 != nil; tag, err1 = history.SetState(url, tag, state) {
			log.Errorf("setting state to %v failed - %v. Retrying...", state.Status, err1)
		}
		log.Infof("setting state to %v successful", state.Status)
	}
	return output, err
}

// getDeployArgs - Get arguments to pass to deploy script as a string array
func getDeployArgs(repoURL string, branch string, server string, subdomain string, access string, kwargs map[string]bool) []string {
	args := []string{"-n", "-u", repoURL, "-b", branch, "-m", server, "-s", subdomain, "-a", access}
	if v, ok := kwargs["redeploy"]; ok && v {
		args = append(args, "-x")
	}
	if v, ok := kwargs["restart"]; ok && v {
		args = append(args, "-r")
	}
	return args
}
