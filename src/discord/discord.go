package discord

import (
	"net/http"
	"encoding/json"
	"fmt"
	"bytes"

	log "github.com/sirupsen/logrus"
)

func PostActionMessage(callbackID string, data map[string]interface{}) error {
	msg := newMessage(callbackID, data["action"].(string), data["result"].(string), data["fields"].([]interface{}))
	payload, err := json.Marshal(msg)
	if err != nil {
		log.Errorf("cannot marshal discord message - %v", err)
		return err
	}
	log.Infof("Posting chat message to discord: %+v", msg)
	resp, err := http.Post(postMessageHookURL, "application/json", bytes.NewBuffer(payload))
	resp.Body.Close()
	if err != nil {
		return fmt.Errorf("discord POST request failed - %v", err)
	}
	log.Infof("Message posted to discord")
	return nil
}

