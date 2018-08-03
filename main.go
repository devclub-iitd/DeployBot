package main

import (
	"log"
	"net/http"
)

func main() {

	initialize()

	http.HandleFunc("/slack/commands/deploy/", deployCommandHandler)
	http.HandleFunc("/slack/interactive/request/", requestHandler)
	http.HandleFunc("/slack/interactive/data-options/", dataOptionsHandler)

	http.HandleFunc("/github/repo/", repoHandler)

	http.HandleFunc("/logs/", logHandler)

	log.Fatal(http.ListenAndServe(":"+Port, nil))
}
