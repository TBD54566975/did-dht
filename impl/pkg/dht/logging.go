package dht

import (
	"github.com/anacrolix/log"
	"github.com/sirupsen/logrus"
)

func init() {
	log.Default.Handlers = []log.Handler{logrusHandler{}}
}

type logrusHandler struct{}

// Handle implements the log.Handler interface for logrus.
// It intentionally downgrades the log level to reduce verbosity.
func (logrusHandler) Handle(record log.Record) {
	entry := logrus.WithFields(logrus.Fields{"names": record.Names})
	msg := record.Msg.String()

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
