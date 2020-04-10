package main

import (
	"os"

	"github.com/devclub-iitd/DeployBot/src/helper"
	log "github.com/sirupsen/logrus"
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

	log.SetFormatter(&log.TextFormatter{
		DisableTimestamp: false,
		FullTimestamp:    true,
		DisableColors:    true,
		TimestampFormat:  "Mon Jan _2 15:04:05 2006",
	})

	// Output to stdout instead of the default stderr
	log.SetOutput(os.Stdout)

}
