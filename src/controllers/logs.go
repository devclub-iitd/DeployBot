package controllers

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"time"

	"github.com/devclub-iitd/DeployBot/src/helper"
	"github.com/devclub-iitd/DeployBot/src/history"
	"github.com/devclub-iitd/DeployBot/src/slack"
	log "github.com/sirupsen/logrus"
)

// logs runs the controller to get logs for a service
func logs(callbackID string, data map[string]interface{}) {
	repoURL := data["git_url"].(string)
	channelID := data["channel"].(string)
	if err := slack.PostChatMessage(channelID, fmt.Sprintf("Fetching logs for %s ...", repoURL), nil); err != nil {
		log.Warnf("error occured in posting message - %v", err)
		return
	}
	log.Infof("Fetching logs for service %s with callback_id as %s", repoURL, callbackID)

	output, err := internalLogs(data)
	if err != nil {
		_ = slack.PostChatMessage(channelID, fmt.Sprintf("Logs for service %s could not be fetched.\nERROR: %s", repoURL, err.Error()), nil)
	} else {
		filePath := path.Join(logDir, "service", fmt.Sprintf("%s.txt", callbackID))
		helper.WriteToFile(filePath, string(output))
		log.Info("starting timer for " + filePath)
		go time.AfterFunc(time.Minute*logsExpiryMins, func() {
			os.Remove(filePath)
			log.Infof("Deleted " + filePath)
		})
		_ = slack.PostChatMessage(channelID,
			fmt.Sprintf("Requested logs for service %s would be available at %s/logs/service/%s.txt for %d minutes.", repoURL, serverURL, callbackID, logsExpiryMins),
			nil)
	}
}

func internalLogs(data map[string]interface{}) ([]byte, error) {
	gitRepoURL := data["git_repo"].(string)
	tailCount := data["tail_count"].(string)
	current := history.GetCurrent(gitRepoURL)
	serverName := current.Server
	if current.Status != "running" {
		log.Infof("service %s is not running, cannot fetch logs", gitRepoURL)
		return nil, fmt.Errorf("service not running")
	}
	return exec.Command(logScriptName, gitRepoURL, serverName, tailCount).CombinedOutput()
}
