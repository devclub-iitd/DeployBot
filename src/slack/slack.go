package slack

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"sync/atomic"

	"github.com/devclub-iitd/DeployBot/src/helper"
	log "github.com/sirupsen/logrus"
	"github.com/teris-io/shortid"
)

// validateRequest validates if a request came from slack.com or not
func validateRequest(r *http.Request) bool {
	ts := r.Header["X-Slack-Request-Timestamp"][0]
	sig := r.Header["X-Slack-Signature"][0]
	bodyBytes, _ := ioutil.ReadAll(r.Body)
	dupReader := ioutil.NopCloser(bytes.NewBuffer(bodyBytes))
	r.Body = dupReader
	bodyString := string(bodyBytes)
	payload := strings.Join([]string{"v0", ts, bodyString}, ":")
	if sha, err := helper.Hash(payload, signingSecret, "sha256"); err != nil {
		log.Errorf("cannot get sha256 of payload(%s) - %v", payload, err)
	} else {
		if ("v0=" + sha) == sig {
			return true
		}
	}
	return false
}

// httpPost makes a post request to slack with the given payload, and token to a url
// It returns an error if the request did not succeed - i.e. the ok field was false in the response
func httpPost(payload map[string]interface{}, url, token string) error {
	payloadByte, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("cannot marshal payload - %v", err)
	}
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payloadByte))
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("Authorization", "Bearer "+token)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("cannot send HTTP POST request - %v", err)
	}
	defer resp.Body.Close()
	var respBody interface{}
	json.NewDecoder(resp.Body).Decode(&respBody)
	if (respBody.(map[string]interface{}))["ok"].(bool) {
		log.Info("HTTP POST request suceeded")
		return nil
	}
	return fmt.Errorf("cannot verify response, got response - %v: ", respBody.(map[string]interface{}))
}

func dialogOpen(payload map[string]interface{}) error {
	if payload == nil {
		return fmt.Errorf("payload is nil")
	}
	log.Infof("sending a HTTP POST request to initiate a dialog open")
	if err := httpPost(payload, dialogURL, accessToken); err != nil {
		return fmt.Errorf("HTTP post request failed - %v", err)
	}
	log.Infof("dialog opened sucessfully")
	return nil
}

func uuid(counter *int32) string {
	id, err := shortid.Generate()
	if err != nil {
		log.Errorf("cannot generate a uuid, falling back to counter - %v", err)
		return fmt.Sprintf("%d", atomic.LoadInt32(counter))
	}
	return id
}

// dialogHandler is a function that has the functionality for handling slack HTTP requests that require initiating a dialog
func dialogHandler(r *http.Request, d dialog, counter *int32) (status int, err error) {
	if !validateRequest(r) {
		log.Warnf("request verification from slack failed")
		return 403, nil
	}
	log.Info("request verification from slack succeeded")
	r.ParseForm()
	triggerID := r.Form["trigger_id"][0]
	var f interface{}
	if err := json.Unmarshal(d.content, &f); err != nil {
		return 500, fmt.Errorf("cannot unmarshall the dialog bytes - %v", err)
	}
	dialogJSON := f.(map[string]interface{})

	dialogJSON["callback_id"] = fmt.Sprintf("%s-%s", d.name, uuid(counter))
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

// DeployCommandHandler handles /deploy commands on slack
func DeployCommandHandler(w http.ResponseWriter, r *http.Request) {
	log.Infof("/deploy command called on slack")
	status, err := dialogHandler(r, deployDialog, &helper.DeployCount)
	w.WriteHeader(status)
	if err != nil {
		log.Errorf("error handling /deploy command - %v", err)
	}
}

// StopCommandHandler handles /stop commands on slack
func StopCommandHandler(w http.ResponseWriter, r *http.Request) {
	log.Infof("/stop command called on slack")
	status, err := dialogHandler(r, stopDialog, &helper.StopCount)
	w.WriteHeader(status)
	if err != nil {
		log.Errorf("error handling /stop command - %v", err)
	}
}

// LogsCommandHandler handles /logs commands on slack
func LogsCommandHandler(w http.ResponseWriter, r *http.Request) {
	log.Infof("/logs command called on slack")
	status, err := dialogHandler(r, logsDialog, &helper.LogsCount)
	w.WriteHeader(status)
	if err != nil {
		log.Errorf("error handling /logs command - %v", err)
	}
}

// ParseAction parses a request and returns the action to take along with its parameters
func ParseAction(r *http.Request) (map[string]interface{}, int, error) {
	if !validateRequest(r) {
		log.Warnf("request verification from slack failed")
		return nil, 403, fmt.Errorf("verification failed")
	}
	log.Info("request verification from slack succeeded")
	r.ParseForm()
	var formPayload interface{}
	json.Unmarshal([]byte(r.Form["payload"][0]), &formPayload)
	formPayloadMap := formPayload.(map[string]interface{})
    user := formPayloadMap["user"].(map[string]string)["name"]
	submissionDataMap := formPayloadMap["submission"].(map[string]interface{})
    submissionDataMap["user"] = user
	callbackID := formPayloadMap["callback_id"].(string)
	log.Infof("action requested with github repo %s with callback_id %s", submissionDataMap["git_repo"].(string), callbackID)
	switch {
	case strings.Contains(callbackID, "deploy"):
        return map[string]interface{}{"action": "deploy", "callback_id": callbackID, "data": submissionDataMap}, 200, nil
	case strings.Contains(callbackID, "stop"):
		return map[string]interface{}{"action": "stop", "callback_id": callbackID, "data": submissionDataMap}, 200, nil
	case strings.Contains(callbackID, "logs"):
		return map[string]interface{}{"action": "logs", "callback_id": callbackID, "data": submissionDataMap}, 200, nil
	}
	return nil, 400, fmt.Errorf("invalid action")
}

// OptionType parses a request and gets the type of options to return for a slack dialog
func OptionType(r *http.Request) (string, int, error) {
	if !validateRequest(r) {
		log.Warnf("request verification from slack failed")
		return "", 403, fmt.Errorf("verification failed")
	}
	log.Info("request verification from slack succeeded")
	r.ParseForm()
	var payload interface{}
	if err := json.Unmarshal([]byte(r.Form["payload"][0]), &payload); err != nil {
		log.Errorf("cannot unmarshal payload - %v", err)
		return "", 500, fmt.Errorf("cannot unmarshal payload")
	}

	payloadMap, _ := payload.(map[string]interface{})
	return payloadMap["name"].(string), 200, nil
}

// PostChatMessage is a function used to post chat message to Slack
func PostChatMessage(channelID, text string, payload map[string]interface{}) error {
	if payload == nil {
		payload = make(map[string]interface{})
	}
	payload["channel"] = channelID
	payload["text"] = text
	log.Infof("sending a HTTP POST request to post chat message to %s channel with \"%s\" as message", channelID, text)
	if err := httpPost(payload, chatPostMessageURL, accessToken); err != nil {
		return fmt.Errorf("HTTP post request failed - %v", err)
	}
	log.Infof("chat message posted succesfully")
	return nil
}
