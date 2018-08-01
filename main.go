package main

import (
	"log"
	"net/http"
)

var jsonInteractive = []byte(`{
	"trigger_id" : "xxxxx",
	"dialog": {
		"callback_id": "deploy-",
		"title": "Deploy App",
		"submit_label": "Deploy",
		"elements": [
			{
				"type": "select",
				"label": "Github Repository",
				"name": "git_repo",
				"data_source": "external"
			},
			{
				"type": "select",
				"label": "Server Name",
				"name": "server_name",
				"data_source": "external"
			}
			]
	}
}`)

func main() {

	initialize()

	http.HandleFunc("/slack/commands/deploy/", deployHandler)
	http.HandleFunc("/slack/interactive/request/", requestHandler)
	http.HandleFunc("/slack/interactive/data-options/", dataOptionsHandler)

	http.HandleFunc("/github/repo/", repoHandler)

	log.Fatal(http.ListenAndServe(":7777", nil))
}
