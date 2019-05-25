package main

import (
	"encoding/json"
	"io/ioutil"
	"time"
	"fmt"
)

var historyFile string = "/etc/nginx/history.json"

type DeployInstance struct {
	Action string `json:"action"`
	Subdomain string `json:"subdomain"`
	Server string `json:"server"`
	Result string `json:"result"`
	Status string `json:"status"`
	Timestamp time.Time `json:"timestamp"`
}

type Service []DeployInstance

func CreateLogEntry(action string, submissionData map[string]interface{}) {
	bytes, _ := ioutil.ReadFile(historyFile)
	history := make(map[string]Service)
	json.Unmarshal([]byte(bytes), &history)

	service := DeployInstance{
		action, submissionData["subdomain"].(string),
		submissionData["server_name"].(string),"in progress", "-",
		time.Now()}

	fmt.Println(history)
	fmt.Println(submissionData)

	history[submissionData["git_repo"].(string)] = append(Service{service},
		history[submissionData["git_repo"].(string)]...)

	file, _ := json.Marshal(history)
	_ = ioutil.WriteFile(historyFile, file, 0644)
}

func UpdateLogEntry(service, result string) {
	bytes, _ := ioutil.ReadFile(historyFile)
	var history map[string]Service
	json.Unmarshal([]byte(bytes), &history)

	history[service][0].Result = result

	file, _ := json.Marshal(history)
	_ = ioutil.WriteFile(historyFile, file, 0644)
}
