package dht

import (
	"strings"

	"github.com/anacrolix/log"
	"github.com/goccy/go-json"
	"github.com/sirupsen/logrus"
)

func init() {
	log.Default.WithDefaultLevel(log.Debug)
	log.Default.Handlers = []log.Handler{logrusHandler{}}
}

type logrusHandler struct{}

// Handle implements the log.Handler interface for logrus.
// It intentionally downgrades the log level to reduce verbosity.
func (logrusHandler) Handle(record log.Record) {
	entry := logrus.WithField("names", strings.Join(record.Names, "/"))
	msg := strings.Replace(record.Msg.String(), "\n", "", -1)

	// Check if the log message is a valid JSON string
	var jsonMsg map[string]any
	if err := json.Unmarshal([]byte(msg), &jsonMsg); err == nil {
		// If the log message is a valid JSON string, escape backslashes and double quotes within the field values
		for k, v := range jsonMsg {
			if strVal, ok := v.(string); ok {
				escaped := strings.Replace(strVal, "\\", "\\\\", -1)
				escaped = strings.Replace(escaped, "\"", "\\\"", -1)
				jsonMsg[k] = escaped
			}
		}
		// Marshal the modified JSON message back to a string
		escapedMsg, _ := json.Marshal(jsonMsg)
		msg = string(escapedMsg)
	} else {
		// If the log message is not a valid JSON string, replace newline characters with empty strings
		msg = strings.Replace(msg, "\n", "", -1)
	}

	switch record.Level {
	case log.Debug:
		entry.Debug(msg)
	case log.Info:
		entry.Info(msg)
	case log.Warning, log.Error:
		entry.Warn(msg)
	default:
		entry.Debug(msg)
	}
}
