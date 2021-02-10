package controllers

import (
	"os/exec"

	"github.com/devclub-iitd/DeployBot/src/history"
	log "github.com/sirupsen/logrus"
)

// NginxRegenerate - Regenerate nginx entries for all deployed and running services
func NginxRegenerate() (string, error) {
	branch := defaultBranch
	for repoURL, state := range history.Services() {
		if state.Status != "running" {
			continue
		}

		kwargs := make(map[string]bool)
		kwargs["restart"] = true
		args := GetDeployArgs(repoURL, branch, state.Server, state.Subdomain, state.Access, kwargs)

		_, err := exec.Command(deployScriptName, args...).CombinedOutput()

		if err != nil {
			return repoURL, err
		}

		log.Infof("Nginx entry regenerated for repoURL: %s, subdomain: %s, branch: %s", repoURL, state.Subdomain, branch)
	}
	return "", nil
}
