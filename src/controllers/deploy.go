package controllers

import (
	"errors"
	"fmt"
	"os/exec"
	"path"

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
	log.Infof("beginning %s with callback_id as %s", actionLog, callbackID)

	logPath := fmt.Sprintf("deploy/%s.txt", callbackID)
	output, err := internaldeploy(actionLog)

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
}

// internaldeploy deploys the given app on the server specified.
func internaldeploy(a *history.ActionInstance) ([]byte, error) {
	branch := defaultBranch

	// This is a value, and thus modifying it does not change the original state in the history map
	state := history.GetState(a.RepoURL)

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
		if err1 := history.SetState(a.RepoURL, state); err1 != nil {
			log.Infof("setting state to deploying failed - %v", err1)
			output = []byte("InternalDeployError: cannot set state to deploying - " + err1.Error())
			return output, err1
		}
		output, err = exec.Command(deployScriptName, "-n", "-u", a.RepoURL, "-b", branch, "-m", a.Server, "-s", a.Subdomain, "-a", a.Access).CombinedOutput()
		if err != nil {
			state.Status = "stopped"
			history.SetState(a.RepoURL, state)
		} else {
			state.Status = "running"
			history.SetState(a.RepoURL, state)
		}

	}
	return output, err
}
