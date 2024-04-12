package dht

import (
	"strings"

	"github.com/anacrolix/log"
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
	msg := strings.Replace(record.Msg.String(), "\n", "\\n", -1)
	entry.Debug(msg)
}
