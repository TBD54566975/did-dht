package dht

import (
	"github.com/anacrolix/log"
	"github.com/sirupsen/logrus"
)

type logHandler struct{}

func (logHandler) Handle(record log.Record) {
	switch record.Level {
	case log.Debug:
		logrus.WithFields(logrus.Fields{
			"names": record.Names,
		}).Debug(record.Msg.Text())
	case log.Info:
		logrus.WithFields(logrus.Fields{
			"names": record.Names,
		}).Info(record.Msg.Text())
	case log.Warning:
		logrus.WithFields(logrus.Fields{
			"names": record.Names,
		}).Warn(record.Msg.Text())
	default:
		logrus.WithFields(logrus.Fields{
			"names": record.Names,
		}).Error(record.Msg.Text())
	}
}
