package main

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"
)

func getGitRepos() {

	u, err := url.Parse(GithubAPIURL)
	if err != nil {
		panic(err)
	}
	u.Path = path.Join(u.Path, "orgs/"+OrganizationName+"/repos")
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
		repoURL := repoMap["html_url"].(string)

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

func parseRepoEvent(msg interface{}) string {

	payloadMap := msg.(map[string]interface{})
	action := payloadMap["action"].(string)
	if action != "created" {
		return "None"
	}
	URL := (payloadMap["repository"].(map[string]interface{}))["clone_url"].(string)
	return URL
}
