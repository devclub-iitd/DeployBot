package controllers

import (
	"path"

	"github.com/devclub-iitd/DeployBot/src/helper"
	log "github.com/sirupsen/logrus"
)

const (
	// hooksScriptName is the name of script used to setup hooks
	hooksScriptName = "hooks.sh"
	// deployScriptName is the name of script used to deploy a service
	deployScriptName = "deploy.sh"
	defaultBranch    = "master"
	// stopScriptName is the name of script used to stop a service
	stopScriptName = "stop.sh"
	// logScriptName is the name of script used to get logs of a service
	logScriptName = "logs.sh"
	// cleanupScriptName is the name of script to clean all dangling docker images
	cleanupScriptName = "cleanup.sh"
	logsExpiryMins    = 10
)

var (
	// serverURL is the URL of the server
	serverURL    string
	globalDomain string
	// LogDir is the place where all the logs are stored
	logDir string
)

func init() {
	serverURL = helper.Env("SERVER_URL", "https://listen.devclub.in")
	globalDomain = helper.Env("GLOBAL_DOMAIN", "devclub.in")

	logDir = helper.Env("LOG_DIR", "/var/logs/deploybot/")

	log.Info("Setting Up Log directory")

	if err := helper.CreateDirIfNotExist(path.Join(logDir, "deploy")); err != nil {
		log.Fatalf("cannot create directory %s - %v", path.Join(logDir, "deploy"), err)
	}
	log.Infof("Log directory %s created", path.Join(logDir, "deploy"))

	if err := helper.CreateDirIfNotExist(path.Join(logDir, "stop")); err != nil {
		log.Fatalf("cannot create directory %s - %v", path.Join(logDir, "stop"), err)
	}
	log.Infof("Log directory %s created", path.Join(logDir, "stop"))

	if err := helper.CreateDirIfNotExist(path.Join(logDir, "git")); err != nil {
		log.Fatalf("cannot create directory %s - %v", path.Join(logDir, "git"), err)
	}
	log.Infof("Log directory %s created", path.Join(logDir, "git"))

	if err := helper.CreateDirIfNotExist(path.Join(logDir, "service")); err != nil {
		log.Fatalf("cannot create directory %s - %v", path.Join(logDir, "service"), err)
	}
	log.Infof("Log directory %s created", path.Join(logDir, "service"))
}
