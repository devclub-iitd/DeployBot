package main

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"time"

	log "github.com/sirupsen/logrus"
)

func CreateLogEntry(submissionData map[string]interface{}, action,
	result string) {

	bytes, _ := ioutil.ReadFile(historyFile)
	history := make(map[string]Service)
	json.Unmarshal([]byte(bytes), &history)

	var service ActionInstance

	if action == "up" {
		service = ActionInstance{
			action, submissionData["subdomain"].(string),
			submissionData["server_name"].(string),
			submissionData["access"].(string), result, time.Now()}
	} else {
		service = ActionInstance{action, "", "", "", result, time.Now()}
	}

	repoHistory := history[submissionData["git_repo"].(string)]
	repoHistory.Actions = append(repoHistory.Actions, service)
	history[submissionData["git_repo"].(string)] = repoHistory

	file, _ := json.Marshal(history)
	_ = ioutil.WriteFile(historyFile, file, 0644)

}

func GetStatus(service string) string {
	bytes, _ := ioutil.ReadFile(historyFile)
	var history map[string]Service
	json.Unmarshal([]byte(bytes), &history)

	if val, ok := history[service]; ok {
		return val.Current.Status
	}
	return "stopped"
}

func GetCurrent(service string) State {
	bytes, _ := ioutil.ReadFile(historyFile)
	var history map[string]Service
	json.Unmarshal([]byte(bytes), &history)

	if val, ok := history[service]; ok {
		return val.Current
	}

	return State{"stopped", "", "", ""}
}

func SetStatus(service, status string) {
	bytes, _ := ioutil.ReadFile(historyFile)
	var history map[string]Service
	json.Unmarshal([]byte(bytes), &history)

	if val, ok := history[service]; ok {
		val.Current.Status = status
		history[service] = val
	} else {
		history[service] = Service{[]ActionInstance{},
			State{status, "", "", ""}}
	}

	file, _ := json.Marshal(history)
	_ = ioutil.WriteFile(historyFile, file, 0644)
}

func SetCurrent(service, status, subdomain, access, server string) {
	bytes, _ := ioutil.ReadFile(historyFile)
	var history map[string]Service
	json.Unmarshal([]byte(bytes), &history)

	current := State{status, subdomain, access, server}
	if val, ok := history[service]; ok {
		val.Current = current
		history[service] = val
	} else {
		history[service] = Service{[]ActionInstance{}, current}
	}

	file, _ := json.Marshal(history)
	_ = ioutil.WriteFile(historyFile, file, 0644)
}

func logsGoRoutine(callbackID string,
	submissionDataMap map[string]interface{}) {

	if chatPostMessage(submissionDataMap["channel"].(string),
		"Fetching logs...", nil) == false {
		log.Warn("Some error occured")
		return
	}

	log.Infof("Fetching logs for service %s with callback_id as %s",
		submissionDataMap["git_url"], callbackID)

	output, err := getServiceLogs(submissionDataMap)

	if err != nil {
		_ = chatPostMessage(submissionDataMap["channel"].(string),
			"Logs could not be fetched.\nERROR: "+err.Error(), nil)
	} else {
		filePath := path.Join(LogDir, "service", callbackID+".txt")
		writeToFile(filePath, string(output))

		log.Info("Starting timer for " + filePath)
		go time.AfterFunc(time.Minute*5, func() {
			os.Remove(filePath)
			log.Infof("Deleted " + filePath)
		})

		_ = chatPostMessage(submissionDataMap["channel"].(string),
			"Requested logs would be available at "+ServerURL+"/logs/service/"+
				callbackID+".txt for 5 minutes.", nil)
	}
}

func getServiceLogs(submissionDataMap map[string]interface{}) ([]byte, error) {
	gitRepoURL := submissionDataMap["git_repo"].(string)
	tailCount := submissionDataMap["tail_count"].(string)
	current := GetCurrent(gitRepoURL)
	serverName := current.Server

	if current.Status != "running" {
		log.Infof("Service %s is not running. Can't Fetch Logs.", gitRepoURL)
		return []byte(""), errors.New("service not running")
	}

	return exec.Command(LogScriptName, gitRepoURL, serverName,
		tailCount).CombinedOutput()
}
