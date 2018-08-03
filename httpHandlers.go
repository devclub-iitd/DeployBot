package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"path"
	"strconv"
	"strings"
)

func deployCommandHandler(w http.ResponseWriter, r *http.Request) {

	if validateRequestSlack(r) {
		fmt.Println("Request verification from slack SUCCESS")
	} else {
		fmt.Println("Request verification from slack FAILED")
		w.WriteHeader(403)
		return
	}

	r.ParseForm()
	triggerID := r.Form["trigger_id"][0]

	var f interface{}
	err := json.Unmarshal(DialogMenu, &f)
	if err != nil {
		panic(err)
	}
	dialogJSON := f.(map[string]interface{})
	DeployCount++
	dialogJSON["callback_id"] = "deploy-" + strconv.Itoa(DeployCount)

	dialogMesg := make(map[string]interface{})
	dialogMesg["trigger_id"] = triggerID
	dialogMesg["dialog"] = dialogJSON

	if dialogOpen(dialogMesg) == false {
		fmt.Fprintf(w, "Some error occured")
		w.WriteHeader(500)
	}

}

func requestHandler(w http.ResponseWriter, r *http.Request) {

	fmt.Println("Got a deploy request")
	if validateRequestSlack(r) {
		fmt.Println("Request verification from slack SUCCESS")
	} else {
		fmt.Println("Request verification from slack FAILED")
		w.WriteHeader(403)
		return
	}

	r.ParseForm()
	fmt.Fprint(w, "")

	var formPayload interface{}
	json.Unmarshal([]byte(r.Form["payload"][0]), &formPayload)
	formPayloadMap := formPayload.(map[string]interface{})

	submissionDataMap := formPayloadMap["submission"].(map[string]interface{})

	if chatPostMessage(submissionDataMap["channel"].(string),
		"Deployment in Progress", nil) == false {
		fmt.Fprintf(w, "Some error occured")
		w.WriteHeader(500)
		return
	}
	fmt.Println("Beginning Deployment of ", submissionDataMap["git_repo"])
	deployOut, err := DeployApp(submissionDataMap)

	writeToFile(path.Join(LogDir, "deploy",
		formPayloadMap["callback_id"].(string)+".txt"), string(deployOut))
	if err != nil {
		if chatPostMessage(submissionDataMap["channel"].(string),
			"Deployment of "+submissionDataMap["git_repo"].(string)+
				" FAILED\n\n  "+"See logs at: "+ServerURL+"/logs/deploy/"+
				formPayloadMap["callback_id"].(string)+".txt", nil) == false {

			fmt.Println("Deployment Failed")
			fmt.Fprintf(w, "Some error occured")
			w.WriteHeader(500)
			return
		}
	} else {
		if chatPostMessage(submissionDataMap["channel"].(string),
			"Deployment of "+submissionDataMap["git_repo"].(string)+
				" Successful\n\n  "+"See logs at: "+ServerURL+"/logs/deploy/"+
				formPayloadMap["callback_id"].(string)+".txt", nil) == false {
			fmt.Println("Deployment Succeeded")
			fmt.Fprintf(w, "Some error occured")
			w.WriteHeader(500)
			return
		}
	}

}

func dataOptionsHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	var payload interface{}
	err := json.Unmarshal([]byte(r.Form["payload"][0]), &payload)
	if err != nil {
		panic(err)
	}
	payloadMap, _ := payload.(map[string]interface{})
	optionType, _ := payloadMap["name"].(string)

	if optionType == "server_name" {
		w.Write(ServerOptionsByte)
	} else if optionType == "git_repo" {
		w.Write(RepoOptionsByte)
	}
}

func repoHandler(w http.ResponseWriter, r *http.Request) {

	fmt.Println("Got a Repository creation event")

	if validateRequestGit(r) {
		fmt.Println("Request verification from Github SUCCESS")
	} else {
		fmt.Println("Request verification from Github FAILED")
		w.WriteHeader(403)
		return
	}

	b, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	// Unmarshal
	var msg interface{}
	err = json.Unmarshal(b, &msg)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	repoName, repoURL := parseRepoEvent(msg)

	if repoURL == "None" {
		fmt.Fprintf(w, "")
		return
	}

	fmt.Println("Beginning Initialization of ", repoURL)

	gitOut, err := addHooks(repoURL)

	writeToFile(path.Join(LogDir, "git", repoName+".txt"), string(gitOut))

	if err != nil {
		_ = chatPostMessage(SlackGeneralChannelID,
			"Initialization of new repo - "+repoURL+" FAILED\n\n  "+
				"See logs at: "+ServerURL+"/logs/git/"+repoName+".txt", nil)
		fmt.Println("Initialization Failed")

	} else {
		_ = chatPostMessage(SlackGeneralChannelID,
			"Initialization of new repo - "+repoURL+" SUCCESS\n\n  "+
				"See logs at: "+ServerURL+"/logs/git/"+repoName+".txt", nil)
		fmt.Println("Initialization Succeeded")
	}
}

func logHandler(w http.ResponseWriter, r *http.Request) {

	filename := strings.TrimPrefix(r.URL.Path, "/logs/")
	filename = strings.TrimSuffix(filename, "/")

	fmt.Println(path.Join(LogDir, filename))
	http.ServeFile(w, r, path.Join(LogDir, filename))

}
