package main

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os/exec"
	"path"

	log "github.com/sirupsen/logrus"
)

func getGitRepos() {

	u, err := url.Parse(GithubAPIURL)
	if err != nil {
		panic(err)
	}
	u.Path = path.Join(u.Path, "orgs/"+OrganizationName+"/repos")

	log.Info("Sending a HTTP GET request to github.com to get repo lists")

	resp, err := http.Get(u.String())
	if err != nil {
		panic(err)
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

	err = json.NewDecoder(reader).Decode(&repoList)
	if err != nil {
		panic(err)
	}
	var repoOptionsList []RepoOption
	for _, repoInterface := range repoList.([]interface{}) {
		repoMap := repoInterface.(map[string]interface{})
		repoName := repoMap["name"].(string)
		repoURL := repoMap["git_url"].(string)

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
		panic(err)
	}
	log.Info("Repo Options successfully set with latest repositories")

}

func validateRequestGit(r *http.Request) bool {

	sig := r.Header["X-Hub-Signature"][0]
	bodyBytes, _ := ioutil.ReadAll(r.Body)
	dupReader := ioutil.NopCloser(bytes.NewBuffer(bodyBytes))
	r.Body = dupReader

	bodyString := string(bodyBytes)

	sha := getHash(bodyString, GithubSecret, "sha1")

	if ("sha1=" + sha) == sig {
		return true
	}
	return false
}

func parseRepoEvent(msg interface{}) (string, string) {

	payloadMap := msg.(map[string]interface{})
	action := payloadMap["action"].(string)
	if action != "created" {
		log.Info("Event type is not of repo creation")
		return "", "None"
	}
	repoMap := payloadMap["repository"].(map[string]interface{})
	repoURL := repoMap["git_url"].(string)
	repoName := repoMap["name"].(string)
	return repoName, repoURL
}

func addHooks(repoURL string) ([]byte, error) {

	log.Infof("Calling %s to initialize hooks for repo", HooksScriptName)

	output, err := exec.Command(HooksScriptName, repoURL).CombinedOutput()
	return output, err
}
