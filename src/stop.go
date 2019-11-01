package main

import (
	"errors"
	"os/exec"

	log "github.com/sirupsen/logrus"
)

func stopGoRoutine(callbackID string,
	submissionDataMap map[string]interface{}) {

	if chatPostMessage(submissionDataMap["channel"].(string),
		"Stopping service...", nil) == false {
		log.Warn("Some error occured")
		return
	}

	log.Infof("Stopping service %s with callback_id as %s",
		submissionDataMap["git_repo"], callbackID)

	_, err := StopApp(submissionDataMap)

	if err != nil {
		CreateLogEntry(submissionDataMap, "down", "failed")

		_ = chatPostMessage(submissionDataMap["channel"].(string),
			submissionDataMap["git_repo"].(string)+" could not be stopped."+
				"\n ERROR: "+err.Error(), nil)
	} else {
		CreateLogEntry(submissionDataMap, "down", "successful")

		_ = chatPostMessage(submissionDataMap["channel"].(string),
			submissionDataMap["git_repo"].(string)+" stopped successfully.", nil)
	}
}

// StopApp stops the given app.
func StopApp(submissionData map[string]interface{}) ([]byte, error) {
	gitRepoURL := submissionData["git_repo"].(string)
	current := GetCurrent(gitRepoURL)
	subdomain := current.Subdomain
	serverName := current.Server
	status := current.Status

	var output []byte
	var err error

	if status == "running" {

		log.Infof("Calling %s to deploy", DeployScriptName)
		SetStatus(gitRepoURL, "stopping")

		output, err = exec.Command(StopScriptName, subdomain, gitRepoURL, serverName).CombinedOutput()

		if err != nil {
			SetStatus(gitRepoURL, "running")
		} else {
			SetCurrent(gitRepoURL, "stopped", "", "", "")
		}
	} else if status == "stopped" {

		log.Infof("Service is already stopped", DeployScriptName)

		output = []byte("Service is already stopped!")
		err = errors.New("already stopped")
	} else if status == "stopping" {

		log.Infof("Service is being stopped. Can't start another stop"+
			"instance.", DeployScriptName)

		output = []byte("Service is already being stopped.")
		err = errors.New("already stopping")
	} else {

		log.Infof("Service is being deployed.", DeployScriptName)

		output = []byte("Service is being deployed. Please wait for the " +
			"process to be completed and try again.")
		err = errors.New("cannot stop while deploying")
	}

	return output, err
}
