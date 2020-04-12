package main

import (
	"os"

	"github.com/devclub-iitd/DeployBot/src/helper"
	"github.com/devclub-iitd/DeployBot/src/history"
	log "github.com/sirupsen/logrus"
	prefixed "github.com/x-cray/logrus-prefixed-formatter"
)

var (
	// serverURL is the URL at which the server is listening to requests
	serverURL string
	// port is the HTTP Port on which the go code will listen
	port string
)

func init() {
	serverURL = helper.Env("SERVER_URL", "https://listen.devclub.iitd.ac.in")
	port = helper.Env("PORT", "7777")

	formatter := new(prefixed.TextFormatter)
	formatter.DisableColors = true
	formatter.FullTimestamp = true
	formatter.ForceFormatting = true
	formatter.TimestampFormat = history.TimeFormatString

	log.SetFormatter(formatter)

	// Output to stdout instead of the default stderr
	log.SetOutput(os.Stdout)

}
