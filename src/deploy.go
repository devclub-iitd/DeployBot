package main

import (
	"encoding/json"
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
		_ = chatPostMessage(submissionDataMap["channel"].(string),
			"Deployment of "+submissionDataMap["git_repo"].(string)+" on "+
				submissionDataMap["server_name"].(string)+" with subdomain "+
				submissionDataMap["subdomain"].(string)+" FAILED\n\n  "+
				"See logs at: "+ServerURL+"/logs/deploy/"+callbackID+".txt", nil)
	} else {
		log.Info("Deployment Successful")
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

	log.Infof("Calling %s to deploy", DeployScriptName)
	output, err := exec.Command(DeployScriptName, "-n", "-u",
		gitRepoURL, "-b", branch, "-m", serverName, "-s", subdomain, "-a", access).CombinedOutput()
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
