package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

func deployHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	triggerID := r.Form["trigger_id"][0]

	var f interface{}
	err := json.Unmarshal(jsonInteractive, &f)
	if err != nil {
		panic(err)
	}
	m := f.(map[string]interface{})
	m["trigger_id"] = triggerID
	mesg, _ := json.Marshal(m)

	req, err := http.NewRequest("POST", SlackDialogURL, bytes.NewBuffer(mesg))
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("Authorization", "Bearer "+SlackAccessToken)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	var respBody interface{}
	json.NewDecoder(resp.Body).Decode(&respBody)
	if (respBody.(map[string]interface{}))["ok"] == false {
		fmt.Fprintf(w, "Some error occured, please try again later")
	}

}

func requestHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	fmt.Println(r.Form)
	w.WriteHeader(200)
	// Need to send a message to the chat using chat.PostMessage and call the
	// actual deployment code
}

func dataOptionsHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	var payload interface{}
	err := json.Unmarshal([]byte(r.Form["payload"][0]), &payload)
	if err != nil {
		panic(err)
	}
	payloadMap, _ := payload.(map[string]interface{})
	optionType, _ := payloadMap["name"].(string)

	if optionType == "server_name" {
		w.Write(ServerOptionsByte)
	} else if optionType == "git_repo" {
		w.Write(RepoOptionsByte)
	}
}
