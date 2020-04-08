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
	logScriptName  = "logs.sh"
	logsExpiryMins = 10
)

var (
	// serverURL is the URL of the server
	serverURL string
	// LogDir is the place where all the logs are stored
	logDir string
)

func init() {
	serverURL = helper.Env("SERVER_URL", "https://listen.devclub.iitd.ac.in")
	logDir = helper.Env("LOG_DIR", "/var/logs/deploybot/")

	log.Info("Setting Up Log directory")
	helper.CreateDirIfNotExist(path.Join(logDir, "deploy"))
	log.Infof("Log directory %s created", path.Join(logDir, "deploy"))
	helper.CreateDirIfNotExist(path.Join(logDir, "git"))
	log.Infof("Log directory %s created", path.Join(logDir, "git"))
	helper.CreateDirIfNotExist(path.Join(logDir, "service"))
	log.Infof("Log directory %s created", path.Join(logDir, "service"))
}
