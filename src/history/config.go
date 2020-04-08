package history

import "time"

// All the deploy and stop requests are logged in historyFile
// templateFile is html template for viewing running services
const (
	historyFile  = "/etc/nginx/history.json"
	templateFile = "status_template.html"
)

// ActionInstance type stores one log entry for a deploy or stop request
type ActionInstance struct {
	Action    string    `json:"action"`
	Subdomain string    `json:"subdomain"`
	Server    string    `json:"server"`
	Access    string    `json:"access"`
	Result    string    `json:"result"`
	Timestamp time.Time `json:"timestamp"`
}

// State stores the current state of service
type State struct {
	Status    string `json:"status"`
	Subdomain string `json:"subdomain"`
	Access    string `json:"access"`
	Server    string `json:"server"`
}

// Service stores the history and current state of a service
type Service struct {
	Actions []ActionInstance `json:"actions"`
	Current State            `json:"current"`
}
