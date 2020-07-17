package history

import (
	"bytes"
	"fmt"
	"html/template"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/devclub-iitd/DeployBot/src/helper"
	log "github.com/sirupsen/logrus"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// All the deploy and stop requests are logged in historyFile
// templateFile is html template for viewing running services
const (
	templateFile         = "status_template.html"
	actionsInMemory      = 20
	healthChecksInMemory = 50
	TimeFormatString     = "Mon Jan _2 15:04:05 2006"
)

var (
	historyFile     string
	healthCheckFile string
	stateFile       string
	// serverURL is the URL of the server
	serverURL string
	domain    string

	statusTemplate *template.Template

	// historyLogger is the zap logger that is used to log actions in a structured log format
	historyLogger *zap.SugaredLogger
	// healthLogger is the zap logger that is used to log actions in a structured log format
	healthLogger *zap.SugaredLogger
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
	buffer.WriteString(a.Timestamp.Format(TimeFormatString))
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
		zap.String("timestamp", a.Timestamp.Format(TimeFormatString)),
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

type EmbedField struct {
	Name   string `json:"name"`
	Value  string `json:"value"`
	Inline bool   `json:"inline"`
}

func (a *ActionInstance) EmbedFields() map[string]interface{} {
	m := map[string]interface{}{"action": a.Action, "result": a.Result}
	if a.Result != "" {
		fields := []interface{}{
			EmbedField{"Result", a.Result, true},
			EmbedField{"Repo", a.RepoURL, false},
			EmbedField{"Logs", fmt.Sprintf("%s/logs/%s", serverURL, a.LogPath), false},
		}
		if a.Action == "deploy" && a.Result == "success" {
			fields = append(fields,
				EmbedField{"URL", fmt.Sprintf("https://%s.%s", a.Subdomain, domain), false},
			)
		}
		m["fields"] = fields
		return m
	}
	fields := []interface{}{
		EmbedField{"User", a.User, true},
		EmbedField{"Repo", a.RepoURL, false},
	}
	if a.Action == "deploy" {
		fields = append(fields,
			EmbedField{"Subdomain", a.Subdomain, true},
			EmbedField{"Server", a.Server, true},
			EmbedField{"Access", a.Access, true},
		)
	}
	m["fields"] = fields
	return m
}

// HealthCheck type stores the result of a healthcheck
type HealthCheck struct {
	Timestamp time.Time `json:"timestamp"`
	RepoURL   string    `json:"repo_url"`
	QueryURL  string    `json:"query_url"`
	Code      int       `json:"code"`
	// Response will also be the error string if the code is non-2xx or there is a HTTP protocol error
	Response string `json:"response"`
}

// Fields returns a slice of zap fields to use in the sugar logger
func (hc *HealthCheck) Fields() []interface{} {
	return []interface{}{
		zap.String("timestamp", hc.Timestamp.Format(TimeFormatString)),
		zap.String("repo_url", hc.RepoURL),
		zap.String("query_url", hc.QueryURL),
		zap.Int("code", hc.Code),
		zap.String("response", hc.Response),
	}
}

// State stores the current state of service
type State struct {
	Timestamp time.Time `json:"timestamp"` // last update of the state message
	Status    string    `json:"status"`
	Subdomain string    `json:"subdomain"`
	Access    string    `json:"access"`
	Server    string    `json:"server"`
	Health    string    `json:"health"`
	Code      int       `json:"health_code"`
}

// Service stores the history and current state of a service
type Service struct {
	Actions      []*ActionInstance `json:"actions"`
	HealthChecks []*HealthCheck    `json:"health_checks"`
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

// newZapLogger returns a sugared logger with output to a given file, in a format we need
func newZapLogger(outfile string) (*zap.SugaredLogger, error) {
	var err error
	w := zapcore.AddSync(&lumberjack.Logger{
		Filename:   outfile,
		MaxSize:    10, // megabytes
		MaxBackups: 30,
		MaxAge:     60, // days
	})
	esink, _, err := zap.Open("stderr")
	if err != nil {
		return nil, err
	}
	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(zapcore.EncoderConfig{
			LineEnding: zapcore.DefaultLineEnding,
		}),
		w,
		zap.InfoLevel,
	)

	logger := zap.New(core, zap.ErrorOutput(esink))
	return logger.Sugar(), nil
}

func init() {
	serverURL = helper.Env("SERVER_URL", "https://listen.devclub.iitd.ac.in")
	domain = helper.Env("DOMAIN", "devclub.iitd.ac.in")
	historyFile = helper.Env("HISTORY_FILE", "/etc/nginx/logs/history.json")
	healthCheckFile = helper.Env("HEALTH_CHECK_FILE", "/etc/nginx/logs/health.json")
	stateFile = helper.Env("STATE_FILE", "/etc/nginx/logs/state.json")

	helper.CreateDirIfNotExist(path.Dir(historyFile))
	helper.CreateDirIfNotExist(path.Dir(healthCheckFile))
	helper.CreateDirIfNotExist(path.Dir(stateFile))

	if err := initState(); err != nil {
		log.Fatalf("cannot read state from %s - %v", stateFile, err)
	}

	var err error
	if historyLogger, err = newZapLogger(historyFile); err != nil {
		log.Fatalf("cannot create a new zap sugared logger for history actions - %v", err)
	}
	defer historyLogger.Sync()
	log.Info("history actions logger construction succeeded")

	if healthLogger, err = newZapLogger(healthCheckFile); err != nil {
		log.Fatalf("cannot create a new zap sugared logger for health checks- %v", err)
	}
	defer healthLogger.Sync()
	log.Info("health check logger construction succeeded")

	statusTemplate, err = template.ParseFiles(templateFile)
	if err != nil {
		log.Fatalf("cannot parse the template file - %v", err)
	}
}
