package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"path"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"
)

func okHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "OK. Healthy!\n") // send healthy data
}

func deployCommandHandler(w http.ResponseWriter, r *http.Request) {

	log.Infof("/deploy command called on slack")

	if validateRequestSlack(r) {
		log.Info("Request verification from slack: SUCCESS")
	} else {
		log.Warn("Request verification from slack: FAILED")
		w.WriteHeader(403)
		return
	}

	r.ParseForm()
	triggerID := r.Form["trigger_id"][0]

	var f interface{}
	err := json.Unmarshal(DeployDialog, &f)
	if err != nil {
		panic(err)
	}
	dialogJSON := f.(map[string]interface{})
	DeployCount++
	dialogJSON["callback_id"] = "deploy-" + strconv.Itoa(DeployCount)

	dialogMesg := make(map[string]interface{})
	dialogMesg["trigger_id"] = triggerID
	dialogMesg["dialog"] = dialogJSON

	log.Info("Created a dialog Message, Beginning to send it")
	if dialogOpen(dialogMesg) == false {
		log.Warn("Some error occured")
		w.WriteHeader(500)
	}

}

func stopCommandHandler(w http.ResponseWriter, r *http.Request) {

	log.Infof("/stop command called on slack")

	if validateRequestSlack(r) {
		log.Infof("Request verification from slack: SUCCESS")
	} else {
		log.Warn("Request verification from slack: FAILED")
		w.WriteHeader(403)
		return
	}

	r.ParseForm()
	triggerID := r.Form["trigger_id"][0]

	var f interface{}
	err := json.Unmarshal(StopDialog, &f)
	if err != nil {
		panic(err)
	}
	dialogJSON := f.(map[string]interface{})
	StopCount++
	dialogJSON["callback_id"] = "stop-" + strconv.Itoa(StopCount)

	dialogMesg := make(map[string]interface{})
	dialogMesg["trigger_id"] = triggerID
	dialogMesg["dialog"] = dialogJSON

	log.Info("Created a dialog Message, Beginning to send it")
	if dialogOpen(dialogMesg) == false {
		log.Warn("Some error occured")
		w.WriteHeader(500)
	}
}

func requestHandler(w http.ResponseWriter, r *http.Request) {

	log.Info("Recieved a request from slack")
	if validateRequestSlack(r) {
		log.Info("Request verification from slack: SUCCESS")
	} else {
		log.Warn("Request verification from slack: FAILED")
		w.WriteHeader(403)
		return
	}

	r.ParseForm()
	fmt.Fprint(w, "")
	w.WriteHeader(200)

	var formPayload interface{}
	json.Unmarshal([]byte(r.Form["payload"][0]), &formPayload)
	formPayloadMap := formPayload.(map[string]interface{})

	submissionDataMap := formPayloadMap["submission"].(map[string]interface{})
	log.Infof(submissionDataMap["git_repo"].(string))
	if strings.Contains(formPayloadMap["callback_id"].(string), "deploy") {
		go deployGoRoutine(formPayloadMap["callback_id"].(string), submissionDataMap)
	} else {
		go stopGoRoutine(formPayloadMap["callback_id"].(string), submissionDataMap)
	}
}

func dataOptionsHandler(w http.ResponseWriter, r *http.Request) {

	log.Info("Recieved a data options request from slack")
	if validateRequestSlack(r) {
		log.Info("Request verification from slack: SUCCESS")
	} else {
		log.Warn("Request verification from slack: FAILED")
		w.WriteHeader(403)
		return
	}

	r.ParseForm()
	var payload interface{}
	err := json.Unmarshal([]byte(r.Form["payload"][0]), &payload)
	if err != nil {
		panic(err)
	}
	payloadMap, _ := payload.(map[string]interface{})
	optionType, _ := payloadMap["name"].(string)

	log.Infof("Data-options requested %s", optionType)
	if optionType == "server_name" {
		// getServers()
		w.Write(ServerOptionsByte)
	} else if optionType == "git_repo" {
		getGitRepos()
		w.Write(RepoOptionsByte)
	}
	log.Info("Data options response sent")
}

func repoHandler(w http.ResponseWriter, r *http.Request) {

	log.Info("Recieved a repository action event")

	if validateRequestGit(r) {
		log.Info("Request verification from Github: SUCCESS")
	} else {
		log.Info("Request verification from Github: FAILED")
		w.WriteHeader(403)
		return
	}

	b, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	// Unmarshal
	var msg interface{}
	err = json.Unmarshal(b, &msg)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	repoName, repoURL := parseRepoEvent(msg)

	if repoURL == "None" {
		fmt.Fprintf(w, "")
		return
	}

	log.Infof("Beginning Initialization of %s", repoURL)

	gitOut, err := addHooks(repoURL)

	writeToFile(path.Join(LogDir, "git", repoName+".txt"), string(gitOut))

	if err != nil {
		log.Error("Initialization of git repo FAILED")
		_ = chatPostMessage(SlackAllHooksChannelID,
			"Initialization of new repo - "+repoURL+" FAILED\n\n  "+
				"See logs at: "+ServerURL+"/logs/git/"+repoName+".txt", nil)
	} else {
		log.Info("Initialization of git repo SUCCESS")
		_ = chatPostMessage(SlackAllHooksChannelID,
			"Initialization of new repo - "+repoURL+" SUCCESS\n\n  "+
				"See logs at: "+ServerURL+"/logs/git/"+repoName+".txt", nil)
	}
}

func logHandler(w http.ResponseWriter, r *http.Request) {

	filename := strings.TrimPrefix(r.URL.Path, "/logs/")
	filename = strings.TrimSuffix(filename, "/")

	log.Infof("Serving file %s for /log request", path.Join(LogDir, filename))

	http.ServeFile(w, r, path.Join(LogDir, filename))

}

func historyHandler(w http.ResponseWriter, r *http.Request) {

	bytes, _ := ioutil.ReadFile(historyFile)

	w.Header().Set("Content-Type", "application/json")
	w.Write(bytes)
}

func statusHandler(w http.ResponseWriter, r *http.Request) {

	bytes, _ := ioutil.ReadFile(historyFile)
	history := make(map[string]Service)
	json.Unmarshal([]byte(bytes), &history)

	tmpl, err := template.ParseFiles(templateFile)
	if err != nil {
		log.Error(err.Error)
	}
	tmpl.Execute(w, history)
}
