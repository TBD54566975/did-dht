package dht

import (
	"github.com/anacrolix/log"
	"github.com/sirupsen/logrus"
)

func init() {
	logrus.SetFormatter(&logrus.JSONFormatter{})
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
		entry.Debugf("%s\n", msg)
	case log.Info:
		entry.Infof("%s\n", msg)
	case log.Warning, log.Error:
		entry.Warnf("%s\n", msg)
	default:
		entry.Debugf("%s\n", msg)
	}
}
