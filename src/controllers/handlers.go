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

// deployAction : Used to represent the action passed to a handler
type deployAction struct {
	command    string
	callbackID string
	data       map[string]interface{}
}

// ActionHandler handles the request from slack to perform an action
func ActionHandler(w http.ResponseWriter, r *http.Request) {
	data, code, err := slack.ParseAction(r)
	if err != nil {
		w.WriteHeader(code)
		if code != 200 {
			log.Errorf("cannot get action to perform - %v", err)
		}
	}
	callbackID := data["callbackID"].(string)
	switch {
	case strings.Contains(callbackID, "redeploy"):
		params := deployAction{"redeploy", callbackID, data["data"].(map[string]interface{})}
		go redeploy(&params)
	case strings.Contains(callbackID, "deploy"):
		params := deployAction{"deploy", callbackID, data["data"].(map[string]interface{})}
		go deploy(&params)
	case strings.Contains(callbackID, "stop"):
		params := deployAction{"stop", callbackID, data["data"].(map[string]interface{})}
		go stop(&params)
	case strings.Contains(callbackID, "logs"):
		params := deployAction{"logs", callbackID, data["data"].(map[string]interface{})}
		go logs(&params)
	default:
		w.WriteHeader(400)
		log.Errorf("Invalid callbackID: %v", callbackID)
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
