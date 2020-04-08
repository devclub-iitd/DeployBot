package controllers

import (
	"errors"
	"fmt"
	"os/exec"

	"github.com/devclub-iitd/DeployBot/src/history"
	"github.com/devclub-iitd/DeployBot/src/slack"
	log "github.com/sirupsen/logrus"
)

// stop stops a running service based on the response from slack
func stop(callbackID string, data map[string]interface{}) {
	repoURL := data["git_repo"].(string)
	channelID := data["channel"].(string)
	if err := slack.PostChatMessage(channelID, fmt.Sprintf("Stopping service (%s) ...", repoURL), nil); err != nil {
		log.Warnf("error occured in posting chat message - %v", err)
		return
	}
	log.Infof("stopping service %s with callback_id as %s", repoURL, callbackID)

	if _, err := internalStop(data); err != nil {
		history.CreateLogEntry(data, "down", "failed")
		_ = slack.PostChatMessage(channelID, fmt.Sprintf("%s cannot be stopped.\n ERROR: %s", repoURL, err.Error()), nil)
	} else {
		history.CreateLogEntry(data, "down", "successful")
		_ = slack.PostChatMessage(channelID, fmt.Sprintf("%s stopped successfully.\n", repoURL), nil)
	}
}

// internalStop actually runs the script to stop the given app.
func internalStop(data map[string]interface{}) ([]byte, error) {
	gitRepoURL := data["git_repo"].(string)
	current := history.GetCurrent(gitRepoURL)
	subdomain := current.Subdomain
	serverName := current.Server
	status := current.Status

	var output []byte
	var err error

	if status == "running" {
		log.Infof("calling %s to stop service(%s)", stopScriptName, gitRepoURL)
		history.SetStatus(gitRepoURL, "stopping")
		if output, err = exec.Command(stopScriptName, subdomain, gitRepoURL, serverName).CombinedOutput(); err != nil {
			history.SetStatus(gitRepoURL, "running")
		} else {
			history.SetCurrent(gitRepoURL, "stopped", "", "", "")
		}
	} else if status == "stopped" {
		log.Infof("service(%s) is already stopped", gitRepoURL)
		output = []byte("Service is already stopped!")
		err = errors.New("already stopped")
	} else if status == "stopping" {
		log.Infof("service(%s) is being stopped. Can't start another stop instance", gitRepoURL)
		output = []byte("Service is already being stopped.")
		err = errors.New("already stopping")
	} else if status == "deploying" {
		log.Infof("service(%s) is being deployed", gitRepoURL)
		output = []byte("Service is being deployed. Please wait for the process to be completed and try again.")
		err = errors.New("cannot stop while deploying")
	}
	return output, err
}
