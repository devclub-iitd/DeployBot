package controllers

import (
	"net/http"
	"path"
	"strings"

	"github.com/devclub-iitd/DeployBot/src/git"
	"github.com/devclub-iitd/DeployBot/src/options"
	"github.com/devclub-iitd/DeployBot/src/slack"
	log "github.com/sirupsen/logrus"
)

// ActionHandler handles the request from slack to perform an action
func ActionHandler(w http.ResponseWriter, r *http.Request) {
	data, code, err := slack.ParseAction(r)
	if err != nil {
		w.WriteHeader(code)
		if code != 200 {
			log.Errorf("cannot get action to perform - %v", err)
		}
	}
	switch data["action"].(string) {
	case "deploy":
		go deploy(data["callback_id"].(string), data["data"].(map[string]interface{}))
	case "stop":
		go stop(data["callback_id"].(string), data["data"].(map[string]interface{}))
	case "logs":
		go logs(data["callback_id"].(string), data["data"].(map[string]interface{}))
	}
}

// RepoHandler handles the github create new repo requests
func RepoHandler(w http.ResponseWriter, r *http.Request) {
	log.Info("recieved a repository action event")
	repo, code, err := git.CreatedRepo(r)
	if err != nil {
		w.WriteHeader(code)
		if code != 200 {
			log.Errorf("cannot parse event - %v", err)
		}
		return
	}
	log.Infof("beginning initialization of %s", repo.URL)
	go addHooks(repo.URL, repo.Name)
	go options.UpdateRepos()
}

// LogHandler handles the requests to get logs i.e. /logs/ endpoint
func LogHandler(w http.ResponseWriter, r *http.Request) {
	filename := strings.TrimPrefix(r.URL.Path, "/logs/")
	filename = strings.TrimSuffix(filename, "/")
	log.Infof("Serving file %s for /log request", path.Join(logDir, filename))
	http.ServeFile(w, r, path.Join(logDir, filename))
}
