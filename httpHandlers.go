package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

func deployCommandHandler(w http.ResponseWriter, r *http.Request) {
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
	fmt.Println(r.Form["payload"][0])

	fmt.Fprint(w, "")

	var formPayload interface{}
	json.Unmarshal([]byte(r.Form["payload"][0]), &formPayload)
	formPayloadMap := formPayload.(map[string]interface{})

	fmt.Println(formPayloadMap["action_ts"].(string))

	mesg := make(map[string]interface{})
	mesg["channel"] = "CBT56106S"
	mesg["text"] = "Deployement in Progress"
	// mesg["ts"] = formPayloadMap["action_ts"].(string)
	// mesg["as_user"] = true

	// fmt.Println(mesg["ts"])

	mesgByte, _ := json.Marshal(mesg)

	req, err := http.NewRequest("POST", "https://slack.com/api/chat.postMessage", bytes.NewBuffer(mesgByte))
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
	fmt.Println(respBody)
	// if (respBody.(map[string]interface{}))["ok"] == false {
	// 	fmt.Fprintf(w, "Some error occured, please try again later")
	// }

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
