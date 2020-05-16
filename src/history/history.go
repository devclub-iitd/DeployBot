package history

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/devclub-iitd/DeployBot/src/helper"
	log "github.com/sirupsen/logrus"
)

// This function is called to initialize in-memory state at startup from the file
func initState() error {
	if !helper.FileExists(stateFile) {
		log.Warnf("stateFile (%s) does not exists, starting a fresh server", stateFile)
		return nil
	}
	mux.Lock()
	defer mux.Unlock()
	bytes, err := ioutil.ReadFile(stateFile)
	if err != nil {
		return fmt.Errorf("cannot read stateFile (%s) - %v", stateFile, err)
	}
	if err := json.Unmarshal(bytes, &history); err != nil {
		return fmt.Errorf("cannot unmarshal json to history - %v", err)
	}

	for repoURL, service := range history {
		actualTag := makeTag(repoURL, *service.Current)
		if service.StateTag != actualTag {
			log.Warnf("State tag for %s in history file is \"%s\", expected \"%s\"", repoURL, service.StateTag, actualTag)
			service.StateTag = actualTag
		}
	}

	return nil
}

// BackupState is called periodically by main as a cron function and whenever state changes
func BackupState() {
	mux.Lock()
	defer mux.Unlock()
	log.Infof("backing up state to %s", stateFile)
	bytes, err := json.MarshalIndent(history, "", "  ")
	if err != nil {
		log.Errorf("cannot marshal status to history - %v", err)
		return
	}
	if err := helper.WriteToFile(stateFile, string(bytes)); err != nil {
		log.Errorf("cannot write state to file - %v", err)
	}
}

func writeAction(a *ActionInstance) {
	historyLogger.Infow("", a.Fields()...)
	historyLogger.Sync()
}

func writeHealth(hc *HealthCheck) {
	healthLogger.Infow("", hc.Fields()...)
	healthLogger.Sync()
}

// StoreAction stores an action entry for an action taken on a service
func StoreAction(a *ActionInstance) {
	mux.Lock()
	defer mux.Unlock()
	if _, ok := history[a.RepoURL]; !ok {
		history[a.RepoURL] = NewService()
	}
	s := history[a.RepoURL]
	s.Actions = append(s.Actions, a)
	if len(s.Actions) > actionsInMemory {
		s.Actions = s.Actions[len(s.Actions)-actionsInMemory:]
	}
	go writeAction(a)
}

// StoreHealth stores an health check entry on a service
func StoreHealth(hc *HealthCheck) {
	mux.Lock()
	defer mux.Unlock()
	// This should never happen, but we still check it
	if _, ok := history[hc.RepoURL]; !ok {
		history[hc.RepoURL] = NewService()
	}
	s := history[hc.RepoURL]
	s.HealthChecks = append(s.HealthChecks, hc)
	if len(s.HealthChecks) > healthChecksInMemory {
		s.HealthChecks = s.HealthChecks[len(s.HealthChecks)-healthChecksInMemory:]
	}
	s.Current.Timestamp = time.Now()
	s.Current.Code = hc.Code
	if hc.Code == 200 {
		s.Current.Health = hc.Response
	} else {
		s.Current.Health = fmt.Sprintf("Not healthy - code %d", hc.Code)
	}
	go writeHealth(hc)
}

// GetState returns the current state of the service and a tag identifying it
func GetState(repoURL string) (State, string) {
	mux.Lock()
	defer mux.Unlock()
	if _, ok := history[repoURL]; !ok {
		history[repoURL] = NewService()
	}
	return *history[repoURL].Current, history[repoURL].StateTag
}

// SetState sets the current state of service
// Compares the current state of repoURL to provided tag, and if it matches,
// sets it to reqState. Returns a tag for the new state and an error.
func SetState(repoURL string, tag string, reqState State) (string, error) {
	mux.Lock()
	defer mux.Unlock()
	if _, ok := history[repoURL]; !ok {
		history[repoURL] = NewService()
	}
	reqState.Timestamp = time.Now()
	var err error
	if reqState.Status == "deploying" && checkSubdomain(reqState.Subdomain) {
		err = errors.New("subdomain in use")
	} else if history[repoURL].StateTag != tag {
		err = fmt.Errorf("old tag provided %s, does not match the original tag %s", tag, history[repoURL].StateTag)
		tag = history[repoURL].StateTag
	} else {
		history[repoURL].Current = &reqState
		tag = makeTag(repoURL, reqState)
		history[repoURL].StateTag = tag
	}
	go BackupState()
	return tag, err
}

func serviceBytes(subdomain string) []byte {
	mux.Lock()
	defer mux.Unlock()
	for _, v := range history {
		if v.Current.Subdomain == subdomain {
			bytes, err := json.MarshalIndent(v, "", "  ")
			if err != nil {
				log.Errorf("cannot marshal state of the current service - %v", err)
				return nil
			}
			return bytes
		}
	}
	return nil
}

// Handler handles the /history/:subdomain endpoint, where it dumps the whole action history
func Handler(w http.ResponseWriter, r *http.Request) {
	p := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	var bytes []byte
	if len(p) < 2 {
		bytes, _ = ioutil.ReadFile(historyFile)
	} else if len(p) == 2 {
		bytes = serviceBytes(p[1])
	} else {
		w.WriteHeader(400)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(bytes)
}

// StatusHandler returns the present status of all the services
func StatusHandler(w http.ResponseWriter, r *http.Request) {
	historyClone := make(map[string]State)
	mux.Lock()
	for k, v := range history {
		historyClone[k] = *(v.Current)
	}
	mux.Unlock()
	statusTemplate.Execute(w, historyClone)
}

// Services returns the current status of all the services in the history map
func Services() map[string]State {
	historyClone := make(map[string]State)
	mux.Lock()
	for k, v := range history {
		historyClone[k] = *(v.Current)
	}
	mux.Unlock()
	return historyClone
}

// checkSubdomain checks if the subdomain is in use
func checkSubdomain(subdomain string) bool {
	for _, v := range history {
		if v.Current.Status != "stopped" && v.Current.Subdomain == subdomain {
			return false
		}
	}
	return true
}
