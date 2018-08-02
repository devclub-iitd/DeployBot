package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
)

func deployCommandHandler(w http.ResponseWriter, r *http.Request) {

	if validateRequest(r) {
		fmt.Println("Request verification from slack SUCCESS")
	} else {
		fmt.Println("Request verification from slack FAILED")
		w.WriteHeader(403)
		return
	}

	r.ParseForm()
	triggerID := r.Form["trigger_id"][0]

	var f interface{}
	err := json.Unmarshal(DialogMenu, &f)
	if err != nil {
		panic(err)
	}
	dialogJSON := f.(map[string]interface{})
	DeployCount++
	dialogJSON["callback_id"] = "deploy-" + strconv.Itoa(DeployCount)

	dialogMesg := make(map[string]interface{})
	dialogMesg["trigger_id"] = triggerID
	dialogMesg["dialog"] = dialogJSON

	if dialogOpen(dialogMesg) == false {
		fmt.Fprintf(w, "Some error occured")
		w.WriteHeader(500)
	}

}

func requestHandler(w http.ResponseWriter, r *http.Request) {

	if validateRequest(r) {
		fmt.Println("Request verification from slack SUCCESS")
	} else {
		fmt.Println("Request verification from slack FAILED")
		w.WriteHeader(403)
		return
	}

	r.ParseForm()
	fmt.Fprint(w, "")

	var formPayload interface{}
	json.Unmarshal([]byte(r.Form["payload"][0]), &formPayload)
	formPayloadMap := formPayload.(map[string]interface{})

	submissionDataMap := formPayloadMap["submission"].(map[string]interface{})

	if chatPostMessage(submissionDataMap["channel"].(string), "Deployement in Progress", nil) == false {
		fmt.Fprintf(w, "Some error occured")
		w.WriteHeader(500)
	}

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

func repoHandler(w http.ResponseWriter, r *http.Request) {
	b, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	// Unmarshal
	var msg interface{}
	err = json.Unmarshal(b, &msg)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	fmt.Println(r)
	fmt.Println(msg)
	w.WriteHeader(200)
	// Need to send a message to the chat using chat.PostMessage and call the
	// actual deployment code
}
