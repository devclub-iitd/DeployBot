package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"path"
	"strings"

	log "github.com/sirupsen/logrus"
)

func deployGoRoutine(callbackID string, submissionDataMap map[string]interface{}) {
	if err := chatPostMessage(submissionDataMap["channel"].(string), "Deployment in Progress", nil); err != nil {
		log.Errorf("cannot post begin deployment chat message - %v", err)
		return
	}
	log.Infof("beginning deployment of %s on server %s with subdomain as %s with callback_id as %s", submissionDataMap["git_repo"], submissionDataMap["server_name"], submissionDataMap["subdomain"], callbackID)

	output, err := internaldeploy(submissionDataMap)
	writeToFile(path.Join(LogDir, "deploy", callbackID+".txt"), string(output))

	if err != nil {
		log.Errorf("Deployment Failed - %v", err)
		CreateLogEntry(submissionDataMap, "up", "failed")
		_ = chatPostMessage(submissionDataMap["channel"].(string),
			"Deployment of "+submissionDataMap["git_repo"].(string)+" on "+
				submissionDataMap["server_name"].(string)+" with subdomain "+
				submissionDataMap["subdomain"].(string)+" FAILED\n\n  "+
				"See logs at: "+ServerURL+"/logs/deploy/"+callbackID+".txt\n"+
				"ERROR: "+err.Error(), nil)
	} else {
		log.Info("Deployment Successful")
		CreateLogEntry(submissionDataMap, "up", "successful")
		_ = chatPostMessage(submissionDataMap["channel"].(string),
			"Deployment of "+submissionDataMap["git_repo"].(string)+" on "+
				submissionDataMap["server_name"].(string)+" with subdomain "+
				submissionDataMap["subdomain"].(string)+" Successful\n\n  "+
				"See logs at: "+ServerURL+"/logs/deploy/"+callbackID+".txt", nil)
	}
}

// internaldeploy deploys the given app on the server specified.
func internaldeploy(submissionData map[string]interface{}) ([]byte, error) {
	gitRepoURL := submissionData["git_repo"].(string)
	serverName := submissionData["server_name"].(string)
	subdomain := submissionData["subdomain"].(string)
	access := submissionData["access"].(string)
	branch := DefaultBranch

	status, err := GetStatus(gitRepoURL)
	if err != nil {
		return nil, fmt.Errorf("cannot get current status of service(%s) - %v", gitRepoURL, err)
	}

	var output []byte
	if status == "stopped" {
		log.Infof("calling %s to deploy %s on %s", DeployScriptName, gitRepoURL, serverName)
		SetStatus(gitRepoURL, "deploying")
		output, err = exec.Command(DeployScriptName, "-n", "-u",
			gitRepoURL, "-b", branch, "-m", serverName, "-s", subdomain, "-a",
			access).CombinedOutput()
		if err != nil {
			SetStatus(gitRepoURL, "stopped")
		} else {
			SetCurrent(gitRepoURL, "running", subdomain, access, serverName)
		}
	} else if status == "running" {
		log.Infof("service(%s) is already running - %v", gitRepoURL, DeployScriptName)
		output = []byte("Service is already running!")
		err = errors.New("already running")
	} else if status == "stopping" {
		log.Infof("service(%s) is stopping. Can't deploy. - %v", gitRepoURL, DeployScriptName)
		output = []byte("Service is stopping. Please wait for the process to be completed and try again.")
		err = errors.New("cannot deploy while service is stopping")
	} else if status == "deploying" {
		log.Infof("service(%s) is being deployed - %v", gitRepoURL, DeployScriptName)
		output = []byte("Service is being deployed. Cannot start another deploy instance.")
		err = errors.New("already deploying")
	}
	return output, err
}

func getServers() error {
	log.Info("executing docker-machine to get list of servers")
	cmd := "docker-machine ls --filter state=Running | tail -n+2 | awk '{print $1}'"
	output, err := exec.Command("bash", "-c", cmd).CombinedOutput()
	if err != nil {
		log.Errorf("cannot get information about servers - %v")
	}
	machines := strings.Split(strings.TrimSpace(string(output)), "\n")
	for _, el := range machines {
		ServersList = append(ServersList, Server{
			Name:       strings.Title(strings.TrimSpace(el)),
			DeployName: strings.TrimSpace(el),
		})
	}
	log.Infof("servers present in docker-machine ls: %v", ServersList)
	serverOptions := make(map[string]interface{})
	serverOptions["options"] = ServersList
	ServerOptionsByte, err = json.Marshal(serverOptions)
	if err != nil {
		log.Errorf("cannot marshall server options - %v", err)
	}
	return nil
}
