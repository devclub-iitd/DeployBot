package main

import (
	"encoding/json"
	"io/ioutil"
	"time"
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
