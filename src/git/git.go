package git

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"

	"github.com/devclub-iitd/DeployBot/src/helper"
	log "github.com/sirupsen/logrus"
)

// Repos queries github and gets the list of repos
func Repos() ([]Repository, error) {
	u, err := url.Parse(apiURL)
	if err != nil {
		return nil, fmt.Errorf("cannot parse github url %s - %v", apiURL, err)
	}
	u.Path = path.Join(u.Path, "orgs/"+organizationName+"/repos")

	log.Infof("sending an HTTP GET request to %s to get list of repos", u.Path)
	resp, err := http.Get(u.String() + "?per_page=100")
	if err != nil {
		return nil, fmt.Errorf("request failed - %v", err)
	}
	defer resp.Body.Close()

	var reader io.ReadCloser
	switch resp.Header.Get("Content-Encoding") {
	case "gzip":
		reader, err = gzip.NewReader(resp.Body)
		defer reader.Close()
	default:
		reader = resp.Body
	}
	var repoList interface{}
	if err := json.NewDecoder(reader).Decode(&repoList); err != nil {
		return nil, fmt.Errorf("cannot parse response into json - %v", err)
	}

	var repositories []Repository
	for _, repoInterface := range repoList.([]interface{}) {
		repoMap := repoInterface.(map[string]interface{})
		repoName := repoMap["name"].(string)
		repoURL := repoMap["ssh_url"].(string)
		repositories = append(repositories, Repository{
			Name: repoName,
			URL:  repoURL})

	}
	return repositories, nil
}

func validateRequest(r *http.Request) bool {
	sig := r.Header["X-Hub-Signature"][0]
	bodyBytes, _ := ioutil.ReadAll(r.Body)
	dupReader := ioutil.NopCloser(bytes.NewBuffer(bodyBytes))
	r.Body = dupReader
	bodyString := string(bodyBytes)
	if sha, err := helper.Hash(bodyString, githubSecret, "sha1"); err != nil {
		log.Errorf("cannot get hash of payload %s - %v", bodyString, err)
	} else {
		if ("sha1=" + sha) == sig {
			return true
		}
	}
	return false
}

func parseEvent(msg interface{}) (string, string, string) {
	payloadMap := msg.(map[string]interface{})
	action := payloadMap["action"].(string)
	if action != "created" {
		log.Info("event type is not of repo creation")
		return action, "", ""
	}
	repoMap := payloadMap["repository"].(map[string]interface{})
	repoURL := repoMap["ssh_url"].(string)
	repoName := repoMap["name"].(string)
	return action, repoName, repoURL
}

// CreatedRepo parses a new repo request from github and returns a Repository and error and status code
func CreatedRepo(r *http.Request) (*Repository, int, error) {
	if validateRequest(r) {
		log.Info("Request verification from Github: SUCCESS")
	} else {
		log.Info("Request verification from Github: FAILED")
		return nil, 403, fmt.Errorf("verification failed")
	}
	b, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		log.Errorf("cannot read request body - %v", err)
		return nil, 500, fmt.Errorf("cannot read request body")
	}
	var msg interface{}
	if err := json.Unmarshal(b, &msg); err != nil {
		log.Errorf("cannot unmarshal request body - %v", err)
		return nil, 500, fmt.Errorf("cannot unmarshal request body")
	}

	event, repoName, repoURL := parseEvent(msg)
	if event != "created" {
		return nil, 200, fmt.Errorf("not a repo creation event")
	}
	return &Repository{
		Name: repoName,
		URL:  repoURL,
	}, 200, nil
}
