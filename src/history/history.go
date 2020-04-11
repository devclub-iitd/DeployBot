package history

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
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
	sugar.Infow("", a.Fields()...)
	sugar.Sync()
}

// StoreAction stores an action entry for an action taken on a service
func StoreAction(a *ActionInstance) {
	mux.Lock()
	defer mux.Unlock()
	if _, ok := history[a.RepoURL]; !ok {
		history[a.RepoURL] = NewService()
	}
	history[a.RepoURL].Actions = append(history[a.RepoURL].Actions, a)
	if len(history[a.RepoURL].Actions) > actionsInMemory {
		history[a.RepoURL].Actions = history[a.RepoURL].Actions[len(history[a.RepoURL].Actions)-actionsInMemory:]
	}
	go writeAction(a)
}

// GetState returns the current state of the service
func GetState(repoURL string) State {
	mux.Lock()
	defer mux.Unlock()
	if _, ok := history[repoURL]; !ok {
		history[repoURL] = NewService()
	}
	return *history[repoURL].Current
}

// SetState sets the current state of service
func SetState(repoURL string, cur State) {
	mux.Lock()
	defer mux.Unlock()
	if _, ok := history[repoURL]; !ok {
		history[repoURL] = NewService()
	}
	cur.Timestamp = time.Now()
	history[repoURL].Current = &cur
	go BackupState()
}

// Handler handles the /history endpoint, where it dumps the whole action history
func Handler(w http.ResponseWriter, r *http.Request) {
	bytes, _ := ioutil.ReadFile(historyFile)
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
