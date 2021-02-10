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
func deploy(callbackID string, data map[string]interface{}) {
	channel := data["channel"].(string)
	actionLog := history.NewAction("deploy", data)
	if err := slack.PostChatMessage(channel, actionLog.String(), nil); err != nil {
		log.Errorf("cannot post begin deployment chat message - %v", err)
		return
	}
	go discord.PostActionMessage(callbackID, actionLog.EmbedFields())
	log.Infof("beginning %s with callback_id as %s", actionLog, callbackID)

	logPath := fmt.Sprintf("deploy/%s.txt", callbackID)

	redeploy := toRedeploy(data)
	output, err := internaldeploy(actionLog, redeploy)

	helper.WriteToFile(path.Join(logDir, logPath), string(output))
	actionLog.LogPath = logPath
	if err != nil {
		actionLog.Result = "failed"
		log.Errorf("%s - %v", actionLog, err)
		history.StoreAction(actionLog)
		slack.PostChatMessage(channel, fmt.Sprintf("%s\nError: %s\n", actionLog, err.Error()), nil)
	} else {
		actionLog.Result = "success"
		log.Info(actionLog.String())
		history.StoreAction(actionLog)
		slack.PostChatMessage(channel, fmt.Sprintf("%s\n", actionLog), nil)
	}
	go discord.PostActionMessage(callbackID, actionLog.EmbedFields())
}

// internaldeploy deploys the given app on the server specified.
func internaldeploy(a *history.ActionInstance, redeploy bool) ([]byte, error) {
	branch := defaultBranch

	// This is a value, and thus modifying it does not change the original state in the history map
	state, tag := history.GetState(a.RepoURL)

	var output []byte
	var err error
	switch state.Status {
	case "running":
		log.Infof("service(%s) is already running", a.RepoURL)
		output = []byte("Service is already running!")
		err = errors.New("already running")
	case "stopping":
		log.Infof("service(%s) is stopping. Can't deploy.", a.RepoURL)
		output = []byte("Service is stopping. Please wait for the process to be completed and try again.")
		err = errors.New("cannot deploy while service is stopping")
	case "deploying":
		log.Infof("service(%s) is being deployed", a.RepoURL)
		output = []byte("Service is being deployed. Cannot start another deploy instance.")
		err = errors.New("already deploying")
	// Assume that, either the service is stopped or does not exist, which means we can deploy.
	default:
		log.Infof("calling %s to deploy %s on %s", deployScriptName, a.RepoURL, a.Server)
		state.Subdomain = a.Subdomain
		state.Access = a.Access
		state.Server = a.Server
		state.Status = "deploying"
		tag, err1 := history.SetState(a.RepoURL, tag, state)
		if err1 != nil {
			log.Infof("setting state to deploying failed - %v", err1)
			output = []byte("InternalDeployError: cannot set state to deploying - " + err1.Error())
			return output, err1
		}

		kwargs := make(map[string]bool)
		kwargs["redeploy"] = redeploy
		args := GetDeployArgs(a.RepoURL, branch, a.Server, a.Subdomain, a.Access, kwargs)

		output, err = exec.Command(deployScriptName, args...).CombinedOutput()
		if err != nil {
			state.Status = "stopped"
		} else {
			state.Status = "running"
		}

		// There should be no error here, ever. Checking it to make sure
		// TODO: On error, set state to an "error" state which only stop should be able to modify
		tag, err1 = history.SetState(a.RepoURL, tag, state)
		for ; err1 != nil; tag, err1 = history.SetState(a.RepoURL, tag, state) {
			log.Errorf("setting state to %v failed - %v. Retrying...", state.Status, err1)
		}
		log.Infof("setting state to %v successful", state.Status)
	}
	return output, err
}

// GetDeployArgs - Get arguments to pass to deploy script as a string array
func GetDeployArgs(repoURL string, branch string, server string, subdomain string, access string, kwargs map[string]bool) []string {
	args := []string{"-n", "-u", repoURL, "-b", branch, "-m", server, "-s", subdomain, "-a", access}
	if v, ok := kwargs["redeploy"]; ok && v {
		args = append(args, "-x")
	}
	if v, ok := kwargs["restart"]; ok && v {
		args = append(args, "-r")
	}
	return args
}

func toRedeploy(data map[string]interface{}) bool {
	redeploy := false
	if v, ok := data["redeploy"]; ok {
		switch v.(type) {
		case bool:
			if v.(bool) {
				redeploy = true
			}
		}

	}
	return redeploy
}
