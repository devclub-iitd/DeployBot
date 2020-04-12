package controllers

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/devclub-iitd/DeployBot/src/history"
	"github.com/nozzle/throttler"
	log "github.com/sirupsen/logrus"
)

const maxParallelHealthchecks = 10

// HealthCheck is a periodically called go function that checks health of every running service
func HealthCheck() {
	services := history.Services()
	t := throttler.New(maxParallelHealthchecks, len(services))
	for k, v := range services {
		if v.Status != "running" {
			continue
		}
		go updateHealth(k, v, t)
		t.Throttle()
	}
	if t.Err() != nil {
		log.Errorf("errors in getting healthchecks (%d) - %s", len(t.Errs()), t.Err())
		for _, err := range t.Errs() {
			log.Errorf("error in healthcheck - %v", err)
		}
	}
}

func updateHealth(repoURL string, st history.State, t *throttler.Throttler) {
	url := fmt.Sprintf("https://%s.%s/healthz", st.Subdomain, baseDomain)
	hc := &history.HealthCheck{
		Timestamp: time.Now(),
		RepoURL:   repoURL,
		QueryURL:  url,
	}
	log.Infof("querying %s to get health of %s", url, repoURL)
	resp, err := http.Get(url)
	if err != nil {
		hc.Response = err.Error()
		history.StoreHealth(hc)
		t.Done(err)
		return
	}
	defer resp.Body.Close()
	hc.Code = resp.StatusCode
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		hc.Response = err.Error()
	} else {
		hc.Response = string(body)
	}
	history.StoreHealth(hc)
	t.Done(nil)
}
