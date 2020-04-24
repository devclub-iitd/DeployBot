package main

import (
	"fmt"
	"net/http"

	"github.com/devclub-iitd/DeployBot/src/controllers"
	"github.com/devclub-iitd/DeployBot/src/history"
	"github.com/devclub-iitd/DeployBot/src/options"
	"github.com/devclub-iitd/DeployBot/src/slack"
	"github.com/robfig/cron"
	log "github.com/sirupsen/logrus"
)

func main() {
	log.Info("beginning initialization of server")
	if err := options.Initialize(); err != nil {
		log.Fatalf("cannot initialize server - %v", err)
	}
	log.Info("initialization of server completed")

	// Slack related HTTP handlers
	http.HandleFunc("/slack/commands/deploy/", slack.DeployCommandHandler)
	http.HandleFunc("/slack/commands/stop/", slack.StopCommandHandler)
	http.HandleFunc("/slack/commands/redeploy/", slack.RedeployCommandHandler)
	http.HandleFunc("/slack/commands/logs/", slack.LogsCommandHandler)
	http.HandleFunc("/slack/interactive/request/", controllers.ActionHandler)
	http.HandleFunc("/slack/interactive/data-options/", options.DataOptionsHandler)

	// Github New repo creation HTTP handler
	http.HandleFunc("/github/repo/", controllers.RepoHandler)

	// General status and history HTTP handlers
	http.HandleFunc("/logs/", controllers.LogHandler)
	http.HandleFunc("/status/", history.StatusHandler)
	http.HandleFunc("/history/", history.Handler)

	// General Health checking handlers
	http.HandleFunc("/", okHandler)
	http.HandleFunc("/healthz", okHandler)

	c := cron.New()
	c.AddFunc("@every 5m", history.BackupState)
	c.AddFunc("@every 20m", options.UpdateRepos)
	c.AddFunc("@every 2m", controllers.HealthCheck)
	c.Start()
	defer c.Stop()

	log.Infof("starting HTTP server on %s:%s", serverURL, port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func okHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "OK. Healthy!\n") // send healthy data
}
