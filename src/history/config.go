package history

import (
	"bytes"
	"fmt"
	"html/template"
	"strings"
	"sync"
	"time"

	"github.com/devclub-iitd/DeployBot/src/helper"
	log "github.com/sirupsen/logrus"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// All the deploy and stop requests are logged in historyFile
// templateFile is html template for viewing running services
const (
	templateFile    = "status_template.html"
	actionsInMemory = 10
)

var (
	historyFile string
	stateFile   string
	// serverURL is the URL of the server
	serverURL string

	statusTemplate *template.Template

	// sugar is the zap logger that is used to log actions in a structured log format
	sugar *zap.SugaredLogger
)

// ActionInstance type stores one log entry for a deploy or stop request
type ActionInstance struct {
	Timestamp time.Time `json:"timestamp"`
	RepoURL   string    `json:"repo_url"`
	Action    string    `json:"action"`
	User      string    `json:"user"`
	Subdomain string    `json:"subdomain"`
	Server    string    `json:"server"`
	Access    string    `json:"access"`
	Result    string    `json:"result"`
	LogPath   string    `json:"log_path"`
}

// NewAction returns a new ActionInstance pointer with the relevant data populated from the data map
func NewAction(action string, data map[string]interface{}) *ActionInstance {
	a := &ActionInstance{
		Timestamp: time.Now(),
		Action:    action,
		RepoURL:   data["git_repo"].(string),
		User:      data["user"].(string),
	}
	if val, ok := data["subdomain"]; ok {
		a.Subdomain = val.(string)
	}
	if val, ok := data["server_name"]; ok {
		a.Server = val.(string)
	}
	if val, ok := data["access"]; ok {
		a.Access = val.(string)
	}
	return a
}

func (a *ActionInstance) String() string {
	var buffer bytes.Buffer
	buffer.WriteString("[")
	buffer.WriteString(a.Timestamp.Format(time.RFC1123))
	buffer.WriteString("] Action ")
	switch a.Action {
	case "deploy":
		buffer.WriteString(fmt.Sprintf("deploy {subdomain = %s, server = %s, access = %s }", a.Subdomain, a.Server, a.Access))
	default:
		buffer.WriteString(a.Action)
	}
	buffer.WriteString(fmt.Sprintf(" on git repo %s by user %s - ", a.RepoURL, a.User))
	if a.Result != "" {
		buffer.WriteString(strings.Title(strings.ToLower(a.Result)))
		if a.LogPath != "" {
			buffer.WriteString(fmt.Sprintf("\nSee Logs at: %s/logs/%s\n", serverURL, a.LogPath))
		}
	} else {
		buffer.WriteString("In Progress\n")
	}
	return buffer.String()
}

// Fields returns a slice of zap fields to use in the sugar logger
func (a *ActionInstance) Fields() []interface{} {
	fields := []interface{}{
		zap.Time("timestamp", a.Timestamp),
		zap.String("repo_url", a.RepoURL),
		zap.String("action", a.Action),
		zap.String("user", a.User),
		zap.String("result", a.Result),
		zap.String("log_path", a.LogPath),
	}
	if a.Action == "deploy" {
		fields = append(fields,
			zap.String("subdomain", a.Subdomain),
			zap.String("server", a.Server),
			zap.String("access", a.Access),
		)
	}
	return fields
}

// HealthCheck type stores the result of a healthcheck
type HealthCheck struct {
	Timestamp time.Time `json:"timestamp"`
	Code      int       `json:"code"`
	Response  string    `json:"response"`
}

// State stores the current state of service
type State struct {
	Timestamp time.Time `json:"timestamp"` // last update of the state message
	Status    string    `json:"status"`
	Subdomain string    `json:"subdomain"`
	Access    string    `json:"access"`
	Server    string    `json:"server"`
	Healthy   string    `json:"healthy"`
}

// Service stores the history and current state of a service
type Service struct {
	Actions      []*ActionInstance `json:"actions"`
	HealthChecks []*HealthCheck    `json:"healths"`
	Current      *State            `json:"current"`
}

// NewService returns a blank service, with the state as stopped
func NewService() *Service {
	return &Service{
		Current: &State{
			Status: "stopped",
		},
	}
}

var history = make(map[string]*Service)
var mux sync.Mutex

func init() {
	serverURL = helper.Env("SERVER_URL", "https://listen.devclub.iitd.ac.in")
	historyFile = helper.Env("HISTORY_FILE", "/etc/nginx/history.json")
	stateFile = helper.Env("STATE_FILE", "/etc/nginx/state.json")

	if err := initState(); err != nil {
		log.Fatalf("cannot read state from %s - %v", stateFile, err)
	}

	var err error
	cfg := zap.Config{
		Level:             zap.NewAtomicLevelAt(zap.InfoLevel),
		Development:       false,
		DisableCaller:     true,
		DisableStacktrace: true,
		Encoding:          "json",
		EncoderConfig: zapcore.EncoderConfig{
			LineEnding: zapcore.DefaultLineEnding,
		},
		OutputPaths:      []string{historyFile},
		ErrorOutputPaths: []string{"stderr"},
	}
	logger, err := cfg.Build()
	if err != nil {
		log.Fatalf("cannot initiazlie zap logger to log history actions - %v", err)
	}
	sugar = logger.Sugar()
	defer sugar.Sync()
	log.Info("history actions logger construction succeeded")

	statusTemplate, err = template.ParseFiles(templateFile)
	if err != nil {
		log.Fatalf("cannot parse the template file - %v", err)
	}
}
