package main

import (
	"os/exec"
)

// DeployApp deploys the given app on the server specified.
func DeployApp(submissionData map[string]interface{}) ([]byte, error) {
	gitRepoURL := submissionData["git_repo"].(string)
	serverName := submissionData["server_name"].(string)
	branch := DefaultBranch

	output, err := exec.Command(DeployScriptName, "-n", "-u",
		gitRepoURL, "-b", branch, "-m", serverName).CombinedOutput()
	return output, err
}
