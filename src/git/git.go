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
	"os"
	"os/exec"
	"path"
	"reflect"

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

func validateActionRequest(r *http.Request) bool {
	headerValue := r.Header.Get("Authorization")
	return "Bearer " + githubActionToken == headerValue
}

// CIHandler handles the request to add CI Files to a repo
func CIHandler(w http.ResponseWriter, r *http.Request) {

	if !validateActionRequest(r) {
		log.Info("Request verification from Github Action: FAILED")
		w.WriteHeader(403)
		return
	}
	log.Info("Request verification from Github Action: SUCCESS")

	if r.Method != "POST" {
		w.WriteHeader(405)
		log.Errorf("received a %s request, expected POST for CI Handler", r.Method)
		return
	}

	r_body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(500)
		log.Errorf("cannot read request body - %v", err)
		return
	}

	var ci CIAction
	if err := json.Unmarshal(r_body, &ci); err != nil {
		w.WriteHeader(500)
		log.Errorf("cannot unmarshal request body - %v", err)
		return
	}

	var ciList []string
	// add Default checks
	ciList = append(ciList, "Default")
	v := reflect.ValueOf(ci)
	for i := 0; i < v.NumField(); i++ {
		if v.Field(i).String() == "true" {
			ciList = append(ciList, v.Type().Field(i).Name)
		}
	}
	addChecks(ciList)
	log.Info("Added checks to pre-commit-config.yaml")

	// call a script to add this file to the repo

	// get ssh_url of the repo
	repositories, err := Repos()
	if err != nil {
		log.Error("Error getting repositories %v", err)
		return
	}
	var repo Repository
	for _, repository := range repositories {
		if repository.Name == ci.Repo {
			repo = repository
		}
	}
	if repo.Name == "" {
		log.Error("Repository not found")
		return
	}

	for _, branch := range repo.Branches {
		go addCI(repo.URL, repo.Name, branch)
	}

}

func addChecks(ciList []string) {
	// combine the data of all files in ciList to a file named .pre-commit-config.yaml

	// create the file even if it exists
	os.Create(PRE_COMMIT_CONFIG)
	f, err := os.OpenFile(PRE_COMMIT_CONFIG, os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		log.Error("Error opening file:pre-commit.config")
		return
	}

	// close the file after the function returns
	defer f.Close()

	for i := 0; i < len(ciList); i++ {
		// add the data from the file to the pre-commit-config file

		var file_name string = ciList[i] + ".ci"
		var file_path string = BASE_FILES_PATH + file_name

		content, err := os.ReadFile(file_path)
		if err != nil {
			log.Error("Error reading file:" + file_path)
			return
		}

		// append the data to the pre-commit-config file
		_, err = fmt.Fprintf(f, string(content)+"\n")
		if err != nil {
			fmt.Printf("Error writing to file:%s\n", PRE_COMMIT_CONFIG)
			return
		}
	}
}

func addCI(repoURL,repoName, branchName string) {
	// call the script ciScriptName
	log.Infof("Calling script %s for repo:%s and branch:%s",ciScriptName, repoName, branchName)
	output,err := exec.Command(ciScriptName, repoURL, branchName).CombinedOutput()
	helper.WriteToFile(path.Join(logDir, "git", fmt.Sprintf("%s:%s.txt", repoName, branchName)), string(output))
	if err != nil {
		log.Errorf("Add CI to git repo - %s: FAILED - %v", repoURL, err)
	} else {
		log.Infof("Add CI to git repo - %s: SUCCESS",repoURL)
	}
}