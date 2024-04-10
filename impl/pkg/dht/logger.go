package dht

import (
	"fmt"

	"github.com/anacrolix/log"
	"github.com/sirupsen/logrus"
)

type logHandler struct{}

func (logHandler) Handle(record log.Record) {
	logrus.SetFormatter(&logrus.JSONFormatter{})

	entry := logrus.WithFields(logrus.Fields{"names": record.Names})

	switch record.Level {
	case log.Debug:
		entry.Debug(record.Msg.String())
	case log.Info:
		entry.Info(record.Msg.String())
	case log.Warning:
		entry.Warn(record.Msg.String())
	default:
		entry.Error(record.Msg.String())
	}

	// Add a newline character after each log message
	fmt.Println()
}
