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

// stop stops a running service based on the response from slack
func stop(params *DeployAction) {
	callbackID := params.callbackID
	data := params.data
	channel := data["channel"].(string)
	actionLog := history.NewAction("stop", data)

	if err := slack.PostChatMessage(channel, actionLog.String(), nil); err != nil {
		log.Warnf("error occured in posting chat message - %v", err)
		return
	}
	go discord.PostActionMessage(callbackID, actionLog.EmbedFields())
	log.Infof("beginning %s with callback_id as %s", actionLog, callbackID)

	logPath := fmt.Sprintf("stop/%s.txt", callbackID)

	output, err := internalStop(actionLog)
	helper.WriteToFile(path.Join(logDir, logPath), string(output))
	actionLog.LogPath = logPath
	if err != nil {
		actionLog.Result = "failed"
		history.StoreAction(actionLog)
		slack.PostChatMessage(channel, fmt.Sprintf("%s\nError: %s\n", actionLog, err.Error()), nil)
	} else {
		actionLog.Result = "success"
		history.StoreAction(actionLog)
		slack.PostChatMessage(channel, actionLog.String(), nil)
	}
	go discord.PostActionMessage(callbackID, actionLog.EmbedFields())
}

// internalStop actually runs the script to stop the given app.
func internalStop(a *history.ActionInstance) ([]byte, error) {
	state, tag := history.GetState(a.RepoURL)

	var output []byte
	var err error

	switch state.Status {
	case "deploying":
		log.Infof("service(%s) is being deployed", a.RepoURL)
		output = []byte("Service is being deployed. Please wait for the process to be completed and try again.")
		err = errors.New("cannot stop while deploying")
	case "stopping":
		log.Infof("service(%s) is being stopped. Can't start another stop instance", a.RepoURL)
		output = []byte("Service is already being stopped.")
		err = errors.New("already stopping")
	case "running":
		log.Infof("calling %s to stop service(%s)", stopScriptName, a.RepoURL)
		state.Status = "stopping"
		tag, err1 := history.SetState(a.RepoURL, tag, state)
		if err1 != nil {
			log.Infof("setting state to stopping failed - %v", err1)
			output = []byte("InternalStopError: cannot set state to stopping - " + err1.Error())
			return output, err1
		}

		if output, err = exec.Command(stopScriptName, state.Subdomain, a.RepoURL, state.Server).CombinedOutput(); err != nil {
			state.Status = "running"
		} else {
			state.Status = "stopped"
		}

		// There should be no error here, ever. Checking it to make sure
		// TODO: On error, set state to an "error" state which only stop should be able to modify
		tag, err1 = history.SetState(a.RepoURL, tag, state)
		for ; err1 != nil; tag, err1 = history.SetState(a.RepoURL, tag, state) {
			log.Errorf("setting state to %v failed - %v. Retrying...", state.Status, err1)
		}
		log.Infof("setting state to %v successful", state.Status)
	default:
		log.Infof("service(%s) is already stopped", a.RepoURL)
		output = []byte("Service is already stopped!")
		err = errors.New("already stopped")
	}
	return output, err
}
