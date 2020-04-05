package main

import (
	"net/http"

	log "github.com/sirupsen/logrus"
)

func main() {
	log.Info("Beginning Initialization")
	initialize()
	log.Info("Initialization completed successfully")

	// Slack related HTTP handlers
	http.HandleFunc("/slack/commands/deploy/", deployCommandHandler)
	http.HandleFunc("/slack/commands/stop/", stopCommandHandler)
	http.HandleFunc("/slack/commands/logs/", logsCommandHandler)
	http.HandleFunc("/slack/interactive/request/", requestHandler)
	http.HandleFunc("/slack/interactive/data-options/", dataOptionsHandler)

	// Github New repo creation HTTP handler
	http.HandleFunc("/github/repo/", repoHandler)

	// General status and history HTTP handlers
	http.HandleFunc("/logs/", logHandler)
	http.HandleFunc("/status/", statusHandler)
	http.HandleFunc("/history/", historyHandler)

	// General Health checking handlers
	http.HandleFunc("/", okHandler)
	http.HandleFunc("/healthz", okHandler)

	log.Infof("Starting HTTP Server on :%s", Port)
	log.Fatal(http.ListenAndServe(":"+Port, nil))
}
