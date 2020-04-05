package main

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os/exec"
	"path"

	log "github.com/sirupsen/logrus"
)

func getGitRepos() error {
	u, err := url.Parse(GithubAPIURL)
	if err != nil {
		return fmt.Errorf("cannot parse github url %s - %v", GithubAPIURL, err)
	}
	u.Path = path.Join(u.Path, "orgs/"+OrganizationName+"/repos")

	log.Infof("sending an HTTP GET request to %s to get list of repos", u.Path)
	resp, err := http.Get(u.String() + "?per_page=100")
	if err != nil {
		return fmt.Errorf("request failed - %v", err)
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
		return fmt.Errorf("cannot parse response into json - %v", err)
	}
	var repoOptionsList []RepoOption
	for _, repoInterface := range repoList.([]interface{}) {
		repoMap := repoInterface.(map[string]interface{})
		repoName := repoMap["name"].(string)
		repoURL := repoMap["ssh_url"].(string)
		Repositories = append(Repositories, Repo{
			Name: repoName,
			URL:  repoURL})
		repoOptionsList = append(repoOptionsList, RepoOption{
			Name: repoName,
			URL:  repoURL,
		})
	}
	repoOptions := make(map[string]interface{})
	repoOptions["options"] = repoOptionsList
	RepoOptionsByte, err = json.Marshal(repoOptions)
	if err != nil {
		return fmt.Errorf("cannot marshal the list of repos to bytes - %v", err)
	}
	log.Info("repo options successfully set with latest repositories")
	return nil
}

func validateRequestGit(r *http.Request) bool {
	sig := r.Header["X-Hub-Signature"][0]
	bodyBytes, _ := ioutil.ReadAll(r.Body)
	dupReader := ioutil.NopCloser(bytes.NewBuffer(bodyBytes))
	r.Body = dupReader
	bodyString := string(bodyBytes)
	if sha, err := getHash(bodyString, GithubSecret, "sha1"); err != nil {
		log.Errorf("cannot get hash of payload %s - %v", bodyString, err)
	} else {
		if ("sha1=" + sha) == sig {
			return true
		}
	}
	return false
}

func parseRepoEvent(msg interface{}) (string, string) {
	payloadMap := msg.(map[string]interface{})
	action := payloadMap["action"].(string)
	if action != "created" {
		log.Info("event type is not of repo creation")
		return "", "None"
	}
	repoMap := payloadMap["repository"].(map[string]interface{})
	repoURL := repoMap["ssh_url"].(string)
	repoName := repoMap["name"].(string)
	return repoName, repoURL
}

func addHooks(repoURL string) ([]byte, error) {
	log.Infof("Calling %s to initialize hooks for repo", HooksScriptName)
	output, err := exec.Command(HooksScriptName, repoURL).CombinedOutput()
	return output, err
}
