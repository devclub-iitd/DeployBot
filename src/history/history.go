package history

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"text/template"
	"time"

	log "github.com/sirupsen/logrus"
)

func getHistory(history *map[string]Service) error {
	bytes, err := ioutil.ReadFile(historyFile)
	if err != nil {
		return fmt.Errorf("cannot read historyFile(%s) - %v", historyFile, err)
	}
	json.Unmarshal(bytes, &history)
	return nil
}

// CreateLogEntry creates a log entry for an action taken on a service in the history file
func CreateLogEntry(submissionData map[string]interface{}, action, result string) {
	history := make(map[string]Service)
	if err := getHistory(&history); err != nil {
		log.Errorf("cannot load history file - %v", err)
		return
	}

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

// GetStatus gets the current status of a given service aka git repo url
func GetStatus(service string) (string, error) {

	var history map[string]Service

	if val, ok := history[service]; ok {
		return val.Current.Status, nil
	}
	return "stopped", nil
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

func HistoryHandler(w http.ResponseWriter, r *http.Request) {
	bytes, _ := ioutil.ReadFile(historyFile)
	w.Header().Set("Content-Type", "application/json")
	w.Write(bytes)
}

func StatusHandler(w http.ResponseWriter, r *http.Request) {
	bytes, _ := ioutil.ReadFile(historyFile)
	history := make(map[string]Service)
	json.Unmarshal([]byte(bytes), &history)

	tmpl, err := template.ParseFiles(templateFile)
	if err != nil {
		log.Error(err.Error())
	}
	tmpl.Execute(w, history)
}
