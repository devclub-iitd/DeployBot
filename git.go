package main

import (
	"compress/gzip"
	"encoding/json"
	"io"
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
