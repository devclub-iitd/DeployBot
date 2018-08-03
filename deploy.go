package main

import (
	"os/exec"
)

// DefaultBranch to be deployed
const (
	DefaultBranch    = "master"
	DeployScriptName = "./scripts/deploy.sh"
)

// DeployApp deploys the given app on the server specified.
func DeployApp(submissionData map[string]interface{}) ([]byte, error) {
	// TODO: THIS IS UNSAFE. This will cause run time panic if not present. see
	// from :
	// https://stackoverflow.com/questions/27545270/how-to-get-value-from-map-golang
	gitRepoURL := submissionData["git_repo"].(string)
	serverName := submissionData["server_name"].(string)
	// gitRepoURL := submissionData["git_repo"] serverName :=
	// submissionData["server_name"]
	branch := DefaultBranch
	// branch := submissionData["branch"].(string) if err {
	//  // set default branch to master
	//  branch = DefaultBranch
	// }

	output, err := exec.Command(DeployScriptName, "-n", "-u",
		gitRepoURL, "-b", branch, "-m", serverName).CombinedOutput()
	if err != nil {
		return nil, err
	}
	return output, nil
}
