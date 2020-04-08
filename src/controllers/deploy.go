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
	if err := slack.PostChatMessage(data["channel"].(string), "Deployment in Progress", nil); err != nil {
		log.Errorf("cannot post begin deployment chat message - %v", err)
		return
	}
	log.Infof("beginning deployment of %s on server %s with subdomain as %s with callback_id as %s", data["git_repo"], data["server_name"], data["subdomain"], callbackID)

	output, err := internaldeploy(data)
	helper.WriteToFile(path.Join(logDir, "deploy", callbackID+".txt"), string(output))

	if err != nil {
		log.Errorf("Deployment Failed - %v", err)
		history.CreateLogEntry(data, "up", "failed")
		_ = slack.PostChatMessage(data["channel"].(string),
			"Deployment of "+data["git_repo"].(string)+" on "+
				data["server_name"].(string)+" with subdomain "+
				data["subdomain"].(string)+" FAILED\n\n  "+
				"See logs at: "+serverURL+"/logs/deploy/"+callbackID+".txt\n"+
				"ERROR: "+err.Error(), nil)
	} else {
		log.Info("Deployment Successful")
		history.CreateLogEntry(data, "up", "successful")
		_ = slack.PostChatMessage(data["channel"].(string),
			"Deployment of "+data["git_repo"].(string)+" on "+
				data["server_name"].(string)+" with subdomain "+
				data["subdomain"].(string)+" Successful\n\n  "+
				"See logs at: "+serverURL+"/logs/deploy/"+callbackID+".txt", nil)
	}
}

// internaldeploy deploys the given app on the server specified.
func internaldeploy(data map[string]interface{}) ([]byte, error) {
	gitRepoURL := data["git_repo"].(string)
	serverName := data["server_name"].(string)
	subdomain := data["subdomain"].(string)
	access := data["access"].(string)
	branch := defaultBranch

	status, err := history.GetStatus(gitRepoURL)
	if err != nil {
		return nil, fmt.Errorf("cannot get current status of service(%s) - %v", gitRepoURL, err)
	}

	var output []byte
	if status == "stopped" {
		log.Infof("calling %s to deploy %s on %s", deployScriptName, gitRepoURL, serverName)
		history.SetStatus(gitRepoURL, "deploying")
		output, err = exec.Command(deployScriptName, "-n", "-u",
			gitRepoURL, "-b", branch, "-m", serverName, "-s", subdomain, "-a",
			access).CombinedOutput()
		if err != nil {
			history.SetStatus(gitRepoURL, "stopped")
		} else {
			history.SetCurrent(gitRepoURL, "running", subdomain, access, serverName)
		}
	} else if status == "running" {
		log.Infof("service(%s) is already running", gitRepoURL)
		output = []byte("Service is already running!")
		err = errors.New("already running")
	} else if status == "stopping" {
		log.Infof("service(%s) is stopping. Can't deploy.", gitRepoURL)
		output = []byte("Service is stopping. Please wait for the process to be completed and try again.")
		err = errors.New("cannot deploy while service is stopping")
	} else if status == "deploying" {
		log.Infof("service(%s) is being deployed", gitRepoURL)
		output = []byte("Service is being deployed. Cannot start another deploy instance.")
		err = errors.New("already deploying")
	}
	return output, err
}
