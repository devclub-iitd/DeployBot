package main

import (
	"encoding/json"
	"errors"
	"os/exec"
	"path"
	"strings"

	log "github.com/sirupsen/logrus"
)

func deployGoRoutine(callbackID string,
	submissionDataMap map[string]interface{}) {
	if chatPostMessage(submissionDataMap["channel"].(string),
		"Deployment in Progress", nil) == false {
		log.Warn("Some error occured")
		return
	}

	log.Infof("Beginning Deployment of %s on server %s with subdomain as %s "+
		"with callback_id as %s", submissionDataMap["git_repo"],
		submissionDataMap["server_name"], submissionDataMap["subdomain"],
		callbackID)

	deployOut, err := DeployApp(submissionDataMap)

	writeToFile(path.Join(LogDir, "deploy", callbackID+".txt"),
		string(deployOut))

	if err != nil {
		log.Error("Deployment Failed")
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

// DeployApp deploys the given app on the server specified.
func DeployApp(submissionData map[string]interface{}) ([]byte, error) {
	gitRepoURL := submissionData["git_repo"].(string)
	serverName := submissionData["server_name"].(string)
	subdomain := submissionData["subdomain"].(string)
	access := submissionData["access"].(string)
	branch := DefaultBranch

	status := GetStatus(gitRepoURL)

	var output []byte
	var err error

	if status == "stopped" {

		log.Infof("Calling %s to deploy", DeployScriptName)
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

		log.Infof("Service is already running", DeployScriptName)

		output = []byte("Service is already running!")
		err = errors.New("already running")
	} else if status == "stopping" {

		log.Infof("Service is stopping. Can't deploy.", DeployScriptName)

		output = []byte("Service is stopping. Please wait for the process to" +
			" be completed and try again.")
		err = errors.New("cannot deploy while service is stopping")
	} else {

		log.Infof("Service is being deployed.", DeployScriptName)

		output = []byte("Service is being deployed. Cannot start another" +
			" deploy instance.")
		err = errors.New("already deploying")
	}

	return output, err
}

func getServers() {

	log.Info("Executing docker-machine to get list of servers")
	cmd := "docker-machine ls --filter state=Running | tail -n+2 | awk '{print $1}'"
	output, err := exec.Command("bash", "-c", cmd).CombinedOutput()
	if err != nil {
		log.Fatal("Cannot get information about servers")
	}
	machines := strings.Split(strings.TrimSpace(string(output)), "\n")
	for _, el := range machines {
		elTrimmed := strings.TrimSpace(el)
		ServersList = append(ServersList, Server{
			Name:       strings.Title(elTrimmed),
			DeployName: elTrimmed,
		})
	}
	log.Info("Servers Present: ", ServersList)
	serverOptions := make(map[string]interface{})
	serverOptions["options"] = ServersList
	ServerOptionsByte, err = json.Marshal(serverOptions)
	if err != nil {
		panic(err)
	}
}
