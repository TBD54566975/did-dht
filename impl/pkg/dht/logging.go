package dht

import (
	"github.com/anacrolix/log"
	"github.com/sirupsen/logrus"
)

func init() {
	log.Default.Handlers = []log.Handler{logrusHandler{}}
}

type logrusHandler struct{}

func (logrusHandler) Handle(record log.Record) {
	entry := logrus.WithFields(logrus.Fields{"names": record.Names})
	msg := record.Msg.String()

	switch record.Level {
	case log.Debug:
		entry.Debug(msg)
	case log.Info:
		entry.Info(msg)
	case log.Warning:
		entry.Warn(msg)
	default:
		entry.Error(msg)
	}
}
