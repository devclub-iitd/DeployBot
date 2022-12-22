package controllers

import (
	"fmt"
	"os/exec"
	"path"

	"github.com/devclub-iitd/DeployBot/src/helper"
	"github.com/devclub-iitd/DeployBot/src/slack"
	log "github.com/sirupsen/logrus"
)

// addHooks calls the hooks script to initialize the hooks for a newly created git repo
func addHooks(repoURL, repoName string, branchName string) {
	log.Infof("Calling %s to initialize hooks for repo", hooksScriptName)
	output, err := exec.Command(hooksScriptName, repoURL, branchName).CombinedOutput()
	helper.WriteToFile(path.Join(logDir, "git", fmt.Sprintf("%s:%s.txt", repoName, branchName)), string(output))
	if err != nil {
		log.Errorf("initialization of git repo(%s) FAILED - %v", repoURL, err)
		_ = slack.PostChatMessage(slack.AllHooksChannelID,
			fmt.Sprintf("Initialization of new repo (%s) FAILED\n\n See logs at: %s/logs/git/%s.txt", repoURL, serverURL, repoName), nil)
	} else {
		log.Info("Initialization of git repo SUCCESS")
		_ = slack.PostChatMessage(slack.AllHooksChannelID,
			fmt.Sprintf("Initialization of new repo (%s) SUCCESS\n\n See logs at: %s/logs/git/%s.txt", repoURL, serverURL, repoName), nil)
	}
}
