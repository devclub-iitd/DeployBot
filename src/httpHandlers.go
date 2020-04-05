package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"path"
	"strings"

	log "github.com/sirupsen/logrus"
)

func okHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "OK. Healthy!\n") // send healthy data
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
		log.Error(err.Error())
	}
	tmpl.Execute(w, history)
}
