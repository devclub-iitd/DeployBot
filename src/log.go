package main

import (
	"encoding/json"
	"io/ioutil"
	"time"
)

var historyFile string = "/etc/nginx/history.json"

type ActionInstance struct {
	Action    string    `json:"action"`
	Subdomain string    `json:"subdomain"`
	Server    string    `json:"server"`
	Access    string    `json:"access"`
	Result    string    `json:"result"`
	Timestamp time.Time `json:"timestamp"`
}

type Service struct {
	Actions []ActionInstance  `json:"actions"`
	Current map[string]string `json:"current"`
}

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
		return val.Current["status"]
	}
	return "stopped"
}

func GetCurrent(service string) map[string]string {
	bytes, _ := ioutil.ReadFile(historyFile)
	var history map[string]Service
	json.Unmarshal([]byte(bytes), &history)

	if val, ok := history[service]; ok {
		return val.Current
	}

	return map[string]string{
		"status":    "stopped",
		"subdomain": "",
		"access":    "",
		"server":    "",
	}
}

func SetStatus(service, status string) {
	bytes, _ := ioutil.ReadFile(historyFile)
	var history map[string]Service
	json.Unmarshal([]byte(bytes), &history)

	if val, ok := history[service]; ok {
		val.Current["status"] = status
		history[service] = val
	} else {
		history[service] = Service{[]ActionInstance{},
			map[string]string{
				"status":    status,
				"access":    "",
				"subdomain": "",
				"server":    "",
			}}
	}

	file, _ := json.Marshal(history)
	_ = ioutil.WriteFile(historyFile, file, 0644)
}

func SetCurrent(service, status, subdomain, access, server string) {
	bytes, _ := ioutil.ReadFile(historyFile)
	var history map[string]Service
	json.Unmarshal([]byte(bytes), &history)

	current := map[string]string{
		"status":    status,
		"server":    server,
		"access":    access,
		"subdomain": subdomain,
	}
	if val, ok := history[service]; ok {
		val.Current = current
		history[service] = val
	} else {
		history[service] = Service{[]ActionInstance{}, current}
	}

	file, _ := json.Marshal(history)
	_ = ioutil.WriteFile(historyFile, file, 0644)
}
