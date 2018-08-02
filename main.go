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

	log.Fatal(http.ListenAndServe(":7777", nil))
}
