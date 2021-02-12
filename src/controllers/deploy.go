package controllers

func deploy(callbackID string, data map[string]interface{}) {
	handleDeploy(&DeployAction{"deploy", callbackID, data})
}
