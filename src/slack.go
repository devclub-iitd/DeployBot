package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"sync/atomic"

	log "github.com/sirupsen/logrus"
)

func validateRequestSlack(r *http.Request) bool {
	ts := r.Header["X-Slack-Request-Timestamp"][0]
	sig := r.Header["X-Slack-Signature"][0]
	bodyBytes, _ := ioutil.ReadAll(r.Body)
	dupReader := ioutil.NopCloser(bytes.NewBuffer(bodyBytes))
	r.Body = dupReader
	bodyString := string(bodyBytes)
	payload := strings.Join([]string{"v0", ts, bodyString}, ":")
	if sha, err := getHash(payload, SlackSigningSecret, "sha256"); err != nil {
		log.Errorf("cannot get sha256 of payload(%s) - %v", payload, err)
	} else {
		if ("v0=" + sha) == sig {
			return true
		}
	}
	return false
}

func slackPost(payload map[string]interface{}, url, token string) (bool, error) {
	payloadByte, err := json.Marshal(payload)
	if err != nil {
		return false, fmt.Errorf("cannot marshal payload - %v", err)
	}
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payloadByte))
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("Authorization", "Bearer "+token)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return false, fmt.Errorf("cannot send HTTP POST request - %v", err)
	}
	defer resp.Body.Close()
	var respBody interface{}
	json.NewDecoder(resp.Body).Decode(&respBody)
	if (respBody.(map[string]interface{}))["ok"].(bool) {
		log.Info("HTTP POST request suceeded")
		return true, nil
	}
	return false, fmt.Errorf("cannot verify response, got response - %v: ", respBody.(map[string]interface{}))
}

func chatPostMessage(channelID, text string, payload map[string]interface{}) error {
	if payload == nil {
		payload = make(map[string]interface{})
	}
	payload["channel"] = channelID
	payload["text"] = text
	log.Infof("sending a HTTP POST request to post chat message to %s channel with \"%s\" as message", channelID, text)
	if ok, err := slackPost(payload, SlackPostMessageURL, SlackAccessToken); !ok {
		return fmt.Errorf("HTTP post request failed - %v", err)
	}
	log.Infof("chat message posted succesfully")
	return nil
}

func dialogOpen(payload map[string]interface{}) error {
	if payload == nil {
		return fmt.Errorf("payload is nil")
	}
	log.Infof("sending a HTTP POST request to initiate a dialog open")
	if ok, err := slackPost(payload, SlackDialogURL, SlackAccessToken); !ok {
		return fmt.Errorf("HTTP post request failed - %v", err)
	}
	log.Infof("dialog opened sucessfully")
	return nil
}

// dialogHandler is a function that has the functionality for handling slack HTTP requests that require initiating a dialog
func dialogHandler(r *http.Request, dialog Dialog, counter *int32) (status int, err error) {
	if validateRequestSlack(r) {
		log.Info("Request verification from slack SUCCESS")
	} else {
		log.Warn("Request verification from slack: FAILED")
		return 403, nil
	}
	r.ParseForm()
	triggerID := r.Form["trigger_id"][0]
	var f interface{}
	if err := json.Unmarshal(dialog.Content, &f); err != nil {
		return 500, fmt.Errorf("cannot unmarshall the dialog bytes - %v", err)
	}
	dialogJSON := f.(map[string]interface{})
	dialogJSON["callback_id"] = fmt.Sprintf("%s-%d", dialog.Name, atomic.LoadInt32(counter))
	atomic.AddInt32(counter, 1)

	dialogMsg := make(map[string]interface{})
	dialogMsg["trigger_id"] = triggerID
	dialogMsg["dialog"] = dialogJSON

	log.Info("created a dialog Message, trying to send it")
	if err := dialogOpen(dialogMsg); err != nil {
		return 500, fmt.Errorf("unable to open dialog - %v", err)
	}
	return 200, nil
}

func deployCommandHandler(w http.ResponseWriter, r *http.Request) {
	log.Infof("/deploy command called on slack")
	status, err := dialogHandler(r, DeployDialog, &DeployCount)
	w.WriteHeader(status)
	if err != nil {
		log.Errorf("error handling /deploy command - %v", err)
	}
}

func stopCommandHandler(w http.ResponseWriter, r *http.Request) {
	log.Infof("/stop command called on slack")
	status, err := dialogHandler(r, StopDialog, &StopCount)
	w.WriteHeader(status)
	if err != nil {
		log.Errorf("error handling /stop command - %v", err)
	}
}

func logsCommandHandler(w http.ResponseWriter, r *http.Request) {
	log.Infof("/logs command called on slack")
	status, err := dialogHandler(r, LogsDialog, &LogsCount)
	w.WriteHeader(status)
	if err != nil {
		log.Errorf("error handling /logs command - %v", err)
	}
}

func requestHandler(w http.ResponseWriter, r *http.Request) {
	log.Info("recieved a request from slack")
	if validateRequestSlack(r) {
		log.Info("request verification from slack: SUCCESS")
	} else {
		log.Warn("request verification from slack: FAILED")
		w.WriteHeader(403)
		return
	}
	r.ParseForm()
	fmt.Fprint(w, "")
	w.WriteHeader(200)

	var formPayload interface{}
	json.Unmarshal([]byte(r.Form["payload"][0]), &formPayload)
	formPayloadMap := formPayload.(map[string]interface{})
	submissionDataMap := formPayloadMap["submission"].(map[string]interface{})
	callbackID := formPayloadMap["callback_id"].(string)
	log.Infof("action requested with github repo %s with callback_id %s", submissionDataMap["git_repo"].(string), callbackID)
	switch {
	case strings.Contains(callbackID, "deploy"):
		go deployGoRoutine(callbackID, submissionDataMap)
	case strings.Contains(callbackID, "stop"):
		go stopGoRoutine(callbackID, submissionDataMap)
	case strings.Contains(callbackID, "logs"):
		go logsGoRoutine(callbackID, submissionDataMap)
	}
}

func dataOptionsHandler(w http.ResponseWriter, r *http.Request) {
	log.Info("recieved a data options request from slack")
	if validateRequestSlack(r) {
		log.Info("request verification from slack: SUCCESS")
	} else {
		log.Warn("request verification from slack: FAILED")
		w.WriteHeader(403)
		return
	}
	r.ParseForm()
	var payload interface{}
	if err := json.Unmarshal([]byte(r.Form["payload"][0]), &payload); err != nil {
		log.Errorf("cannot unmarshal payload - %v", err)
	}

	payloadMap, _ := payload.(map[string]interface{})
	optionType, _ := payloadMap["name"].(string)

	log.Infof("data-options requested for option %s", optionType)
	switch optionType {
	case "server_name":
		// This is done at the time of initialization, so not needed every time
		// getServers()
		w.Write(ServerOptionsByte)
	case "git_repo":
		getGitRepos()
		w.Write(RepoOptionsByte)
	}
	log.Info("data options response sent")
}
