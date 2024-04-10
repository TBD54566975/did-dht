package dht

import (
	"github.com/anacrolix/log"
	"github.com/sirupsen/logrus"
)

type logHandler struct{}

func (logHandler) Handle(record log.Record) {
	entry := logrus.WithFields(logrus.Fields{"names": record.Names})
	msg := record.Msg.String()

	switch record.Level {
	case log.Debug:
		entry.Debugf("%s\n", msg)
	case log.Info:
		entry.Infof("%s\n", msg)
	case log.Warning:
		entry.Warnf("%s\n", msg)
	default:
		entry.Errorf("%s\n", msg)
	}
}
