// Package options stores the generic config and returns to HTTP requests to get the state
package options

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
	"sort"
	"strings"

	"github.com/devclub-iitd/DeployBot/src/git"
	"github.com/devclub-iitd/DeployBot/src/slack"
	log "github.com/sirupsen/logrus"
)

// RepoOption is the type for the options of list of repositories
type RepoOption struct {
	Name string `json:"label"`
	URL  string `json:"value"`
}

// repoOptionsByte is the byte array of the options of repositories
var repoOptionsByte []byte

// UpdateRepos updates the internal state of repos. This is called when a new repo is added.
func UpdateRepos() {
	if err := internalUpdateRepos(); err != nil {
		log.Errorf("cannot update list of repos - %v", err)
	}
}

// UpdateRepos updates the internal state of repos. This is called when a new repo is added.
func internalUpdateRepos() error {
	repos, err := git.Repos()
	if err != nil {
		return fmt.Errorf("cannot get repos from github - %v", err)
	}
	var repoOptionsList []RepoOption
	for _, repo := range repos {
		repoOptionsList = append(repoOptionsList, RepoOption{
			Name: repo.Name,
			URL:  repo.URL,
		})
	}
	sort.Slice(repoOptionsList, func(i, j int) bool {
		return repoOptionsList[i].Name < repoOptionsList[j].Name
	})
	repoOptions := make(map[string]interface{})
	repoOptions["options"] = repoOptionsList
	repoOptionsByte, err = json.Marshal(repoOptions)
	if err != nil {
		return fmt.Errorf("cannot marshal the list of repos to bytes - %v", err)
	}
	log.Info("repo options successfully set with latest repositories")
	return nil
}

// Server is the type storing the information about our server names and their
// IP
type Server struct {
	Name       string `json:"label"`
	DeployName string `json:"value"`
}

// serverOptionsByte is the byte array of the options of server
var serverOptionsByte []byte

func servers() error {
	log.Info("executing docker-machine to get list of servers")
	cmd := "docker-machine ls --filter state=Running | tail -n+2 | awk '{print $1}'"
	output, err := exec.Command("bash", "-c", cmd).Output()
	if err != nil {
		return fmt.Errorf("cannot get information about servers - %v", err)
	}
	machines := strings.Split(strings.TrimSpace(string(output)), "\n")
	var serversList []Server
	for _, el := range machines {
		serversList = append(serversList, Server{
			Name:       strings.Title(strings.TrimSpace(el)),
			DeployName: strings.TrimSpace(el),
		})
	}
	if len(serversList) == 0 {
		return fmt.Errorf("zero servers to deploy to, not okay")
	}
	log.Infof("servers present in docker-machine ls: %v", serversList)
	serverOptions := make(map[string]interface{})
	serverOptions["options"] = serversList
	serverOptionsByte, err = json.Marshal(serverOptions)
	if err != nil {
		return fmt.Errorf("cannot marshall server options - %v", err)
	}
	return nil
}

// DataOptionsHandler handles the request from slack to return options for the dialogs
func DataOptionsHandler(w http.ResponseWriter, r *http.Request) {
	log.Info("recieved a data options request from slack")
	oType, code, err := slack.OptionType(r)
	w.WriteHeader(code)
	if err != nil {
		if code != 200 {
			log.Errorf("cannot get option type from request - %v", err)
		}
		return
	}
	log.Infof("data-options requested for option %s", oType)
	switch oType {
	case "server_name":
		w.Write(serverOptionsByte)
	case "git_repo":
		w.Write(repoOptionsByte)
	}
	log.Info("data options response sent")
}

// Initialize gets the list of repos and servers. Called from main
func Initialize() error {
	if err := internalUpdateRepos(); err != nil {
		return fmt.Errorf("cannot update repos list on initialization - %v", err)
	}
	if err := servers(); err != nil {
		return fmt.Errorf("cannot get list of servers from docker-machine - %v", err)
	}
	return nil
}

// CleanupDanglingImages - Cleans up all the dangling images
func CleanupDanglingImages() {
	app := "docker"
	arg0 := "system"
	arg1 := "prune"

	cmd := exec.Command(app, arg0, arg1)
	output, err := cmd.Output()

	log.Infof(string(output))

	if err != nil {
		log.Errorf("Error - %v", err)
	}
}
