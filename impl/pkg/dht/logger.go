package dht

import (
	"os"

	"github.com/anacrolix/log"
	"github.com/sirupsen/logrus"
)

type logHandler struct{}

func (logHandler) Handle(record log.Record) {
	logrus.SetFormatter(&logrus.JSONFormatter{})
	logrus.SetOutput(&delimitedWriter{})
	switch record.Level {
	case log.Debug:
		logrus.WithFields(logrus.Fields{
			"names": record.Names,
		}).Debug(record.Msg.String())
	case log.Info:
		logrus.WithFields(logrus.Fields{
			"names": record.Names,
		}).Info(record.Msg.String())
	case log.Warning:
		logrus.WithFields(logrus.Fields{
			"names": record.Names,
		}).Warn(record.Msg.String())
	default:
		logrus.WithFields(logrus.Fields{
			"names": record.Names,
		}).Error(record.Msg.String())
	}
}

type delimitedWriter struct{}

func (w *delimitedWriter) Write(p []byte) (n int, err error) {
	// Write the log message and append a newline character
	return os.Stdout.Write(append(p, '\n'))
}
