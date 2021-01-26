package controllers

import (
	"os/exec"

	"github.com/devclub-iitd/DeployBot/src/history"
	log "github.com/sirupsen/logrus"
)

// NginxRegenerate - Regenerate nginx entries for all deployed services
func NginxRegenerate() (string, error) {
	for repoURL, service := range *history.GetHistory() {
		state := service.Current
		branch := defaultBranch
		_, err := exec.Command(deployScriptName, "-n", "-u", repoURL, "-b", branch, "-m", state.Server, "-s", state.Subdomain, "-a", state.Access, "-r").CombinedOutput()
		if err != nil {
			state.Status = "stopped"
			return repoURL, err
		}
		log.Infof("Nginx entry regenerated for repoURL: %s, subdomain: %s, branch: %s", repoURL, state.Subdomain, branch)
		state.Status = "running"

	}
	return "", nil
}
