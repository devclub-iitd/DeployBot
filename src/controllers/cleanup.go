package controllers

import (
	"os/exec"

	log "github.com/sirupsen/logrus"
)

// CleanupDanglingImages - Cleans up all the dangling images on all machines
func CleanupDanglingImages() {
	cmd := exec.Command(cleanupScriptName)
	output, err := cmd.Output()

	log.Infof(string(output))

	if err != nil {
		log.Errorf("Error while cleaning up - %v", err)
	}
}
