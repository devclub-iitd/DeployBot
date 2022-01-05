package controllers

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"time"

	"github.com/devclub-iitd/DeployBot/src/discord"
	"github.com/devclub-iitd/DeployBot/src/helper"
	"github.com/devclub-iitd/DeployBot/src/history"
	"github.com/devclub-iitd/DeployBot/src/slack"
	log "github.com/sirupsen/logrus"
)

// logs runs the controller to get logs for a service
func logs(params *deployAction) {
	callbackID := params.callbackID
	data := params.data
	channelID := data["channel"].(string)
	actionLog := history.NewAction("logs", data)

	if err := slack.PostChatMessage(channelID, fmt.Sprintf("Fetching logs for %s ...", actionLog.CompleteURL), nil); err != nil {
		log.Warnf("error occured in posting message - %v", err)
		return
	}
	go discord.PostActionMessage(callbackID, actionLog.EmbedFields())
	log.Infof("Fetching logs for service %s with callback_id as %s", actionLog.CompleteURL, callbackID)

	output, err := internalLogs(data, actionLog)
	if err != nil {
		_ = slack.PostChatMessage(channelID, fmt.Sprintf("Logs for service %s could not be fetched.\nERROR: %s", actionLog.CompleteURL, err.Error()), nil)
		actionLog.Result = "failed"
	} else {
		actionLog.Result = "success"
		filePath := path.Join("service", fmt.Sprintf("%s.txt", callbackID))
		helper.WriteToFile(path.Join(logDir, filePath), string(output))
		actionLog.LogPath = filePath
		log.Info("starting timer for log file: %s", filePath)
		go time.AfterFunc(time.Minute*logsExpiryMins, func() {
			os.Remove(filePath)
			log.Infof("Deleted log file: %s", filePath)
		})
		_ = slack.PostChatMessage(channelID,
			fmt.Sprintf("Requested logs for service %s would be available at %s/logs/%s for %d minutes.", actionLog.CompleteURL, serverURL, filePath, logsExpiryMins),
			nil)
	}
	go discord.PostActionMessage(callbackID, actionLog.EmbedFields())
}

func internalLogs(data map[string]interface{}, a *history.ActionInstance) ([]byte, error) {
	var url string
	if a.Branch == defaultBranch {
		url = a.RepoURL
	} else {
		url = a.CompleteURL
	}
	tailCount := data["tail_count"].(string)
	current, _ := history.GetState(url)
	serverName := current.Server
	if current.Status != "running" {
		log.Infof("service %s is not running, cannot fetch logs", url)
		return nil, fmt.Errorf("service not running")
	}
	return exec.Command(logScriptName, url, serverName, tailCount).CombinedOutput()
}
