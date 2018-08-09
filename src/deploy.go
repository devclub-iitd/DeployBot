package main

import (
	"os/exec"
	"path"

	log "github.com/sirupsen/logrus"
)

func deployGoRoutine(callbackID string,
	submissionDataMap map[string]interface{}) {
	if chatPostMessage(submissionDataMap["channel"].(string),
		"Deployment in Progress", nil) == false {
		log.Warn("Some error occured")
		return
	}

	log.Infof("Beginning Deployment of %s on server %s with callback_id as %s",
		submissionDataMap["git_repo"], submissionDataMap["server_name"],
		callbackID)

	deployOut, err := DeployApp(submissionDataMap)

	writeToFile(path.Join(LogDir, "deploy", callbackID+".txt"),
		string(deployOut))

	if err != nil {
		log.Error("Deployment Failed")
		_ = chatPostMessage(submissionDataMap["channel"].(string),
			"Deployment of "+submissionDataMap["git_repo"].(string)+" on "+
				submissionDataMap["server_name"].(string)+
				" FAILED\n\n  "+"See logs at: "+ServerURL+"/logs/deploy/"+
				callbackID+".txt", nil)
	} else {
		log.Info("Deployment Successful")
		_ = chatPostMessage(submissionDataMap["channel"].(string),
			"Deployment of "+submissionDataMap["git_repo"].(string)+" on "+
				submissionDataMap["server_name"].(string)+
				" Successful\n\n  "+"See logs at: "+ServerURL+"/logs/deploy/"+
				callbackID+".txt", nil)
	}
}

// DeployApp deploys the given app on the server specified.
func DeployApp(submissionData map[string]interface{}) ([]byte, error) {
	gitRepoURL := submissionData["git_repo"].(string)
	serverName := submissionData["server_name"].(string)
	subdomain := submissionData["subdomain"].(string)
	branch := DefaultBranch

	log.Infof("Calling %s to deploy", DeployScriptName)
	output, err := exec.Command(DeployScriptName, "-n", "-u",
		gitRepoURL, "-b", branch, "-m", serverName, "-s", subdomain).CombinedOutput()
	return output, err
}
