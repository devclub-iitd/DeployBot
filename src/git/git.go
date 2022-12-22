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

	client := &http.Client{}
	req, err := http.NewRequest("GET", u.String()+"?per_page=100", nil)
	if err != nil {
		return nil, fmt.Errorf("request creation failed - %v", err)
	}
	req.SetBasicAuth(url.QueryEscape(githubUserName), url.QueryEscape(githubAccessToken))

	log.Infof("sending an HTTP GET request to %s to get list of repos", u.Path)
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed - %v", err)
	}
	defer resp.Body.Close()

	var reader io.ReadCloser
	switch resp.Header.Get("Content-Encoding") {
	case "gzip":
		reader, err = gzip.NewReader(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read using gzip - %v", err)
		}
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
		branches, err := Branches(repoName)
		if err != nil {
			log.Errorf("Unable to fetch branches for repo %s, err: %v", repoURL, err)
			branches = []string{""}
		}
		repositories = append(repositories, Repository{
			Name:     repoName,
			URL:      repoURL,
			Branches: branches,
		})

	}
	return repositories, nil
}

func Branches(repoName string) ([]string, error) {
	u, err := url.Parse(apiURL)
	if err != nil {
		return nil, fmt.Errorf("cannot parse github url %s - %v", apiURL, err)
	}

	u.Path = path.Join(u.Path, "repos", organizationName, repoName, "branches")
	client := &http.Client{}
	req, err := http.NewRequest("GET", u.String()+"?per_page=100", nil)
	if err != nil {
		return nil, fmt.Errorf("request creation failed - %v", err)
	}
	req.SetBasicAuth(url.QueryEscape(githubUserName), url.QueryEscape(githubAccessToken))

	log.Infof("sending an HTTP GET request to %s to get list of branches", u.Path)
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed - %v", err)
	}
	defer resp.Body.Close()
	var reader io.ReadCloser
	switch resp.Header.Get("Content-Encoding") {
	case "gzip":
		reader, err = gzip.NewReader(resp.Body)
		defer reader.Close()
		if err != nil {
			return nil, fmt.Errorf("failed to read using gzip - %v", err)
		}
	default:
		reader = resp.Body
	}
	var branchList interface{}
	if err := json.NewDecoder(reader).Decode(&branchList); err != nil {
		return nil, fmt.Errorf("cannot parse response into json - %v", err)
	}

	var branches []string
	for _, branchInterface := range branchList.([]interface{}) {
		branchMap := branchInterface.(map[string]interface{})
		branchName := branchMap["name"].(string)
		branches = append(branches, branchName)
	}
	log.Infof("Fetched branches for %s: %v", repoName, branches)
	return branches, nil
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

func parseEvent(msg interface{}) (string, string, string, string) {
	payloadMap := msg.(map[string]interface{})
	action := payloadMap["action"].(string)
	if action != "created" {
		log.Info("event type is not of repo creation")
		return action, "", "", ""
	}
	repoMap := payloadMap["repository"].(map[string]interface{})
	repoURL := repoMap["ssh_url"].(string)
	repoName := repoMap["name"].(string)
	repoBranch := repoMap["default_branch"].(string)
	return action, repoName, repoURL, repoBranch
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

	event, repoName, repoURL, repoBranch := parseEvent(msg)
	if event != "created" {
		return nil, 200, fmt.Errorf("not a repo creation event")
	}
	return &Repository{
		Name: repoName,
		URL:  repoURL,
		Branches: []string{
			repoBranch,
		},
	}, 200, nil
}
