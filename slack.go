package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"
)

func validateRequestSlack(r *http.Request) bool {

	ts := r.Header["X-Slack-Request-Timestamp"][0]
	sig := r.Header["X-Slack-Signature"][0]
	bodyBytes, _ := ioutil.ReadAll(r.Body)
	dupReader := ioutil.NopCloser(bytes.NewBuffer(bodyBytes))
	r.Body = dupReader

	bodyString := string(bodyBytes)

	stringsToJoin := []string{}
	stringsToJoin = append(stringsToJoin, "v0")
	stringsToJoin = append(stringsToJoin, ts)
	stringsToJoin = append(stringsToJoin, bodyString)

	sha := getHash(strings.Join(stringsToJoin, ":"), SlackSigningSecret, "sha256")

	if ("v0=" + sha) == sig {
		return true
	}
	return false
}

func chatPostMessage(channelID string, text string, payload map[string]interface{}) bool {

	if payload == nil {
		payload = make(map[string]interface{})
	}

	payload["channel"] = channelID
	payload["text"] = text

	payloadByte, _ := json.Marshal(payload)

	req, err := http.NewRequest("POST", SlackPostMessageURL, bytes.NewBuffer(payloadByte))
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
	if (respBody.(map[string]interface{}))["ok"].(bool) {
		return true
	}
	return false
}

func dialogOpen(payload map[string]interface{}) bool {

	if payload == nil {
		return false
	}

	payloadByte, _ := json.Marshal(payload)

	req, err := http.NewRequest("POST", SlackDialogURL, bytes.NewBuffer(payloadByte))
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
	if (respBody.(map[string]interface{}))["ok"].(bool) {
		return true
	}
	return false
}
