package history

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

func getHistory(history *map[string]Service) error {
	bytes, err := ioutil.ReadFile(historyFile)
	if err != nil {
		return fmt.Errorf("cannot read historyFile(%s) - %v", historyFile, err)
	}
	json.Unmarshal(bytes, history)
	return nil
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
