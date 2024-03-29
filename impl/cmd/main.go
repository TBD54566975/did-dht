package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"syscall"
	"time"

	"github.com/TBD54566975/ssi-sdk/util"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/TBD54566975/did-dht-method/config"
	"github.com/TBD54566975/did-dht-method/pkg/dht"
	"github.com/TBD54566975/did-dht-method/pkg/server"
	"github.com/TBD54566975/did-dht-method/pkg/telemetry"
)

// main godoc
//
//	@title			The DID DHT Service
//	@description	The DID DHT Service
//	@contact.name	TBD
//	@contact.url	https://github.com/TBD54566975/did-dht-method
//	@contact.email	tbd-developer@squareup.com
//	@license.name	Apache 2.0
//	@license.url	http://www.apache.org/licenses/LICENSE-2.0.html
func main() {
	logrus.SetFormatter(&logrus.JSONFormatter{
		DisableTimestamp: false,
		PrettyPrint:      true,
	})
	logrus.SetReportCaller(true)

	log := logrus.NewEntry(logrus.StandardLogger()).WithField("version", config.Version)
	log.Info("starting up")

	if err := run(); err != nil {
		logrus.WithError(err).Fatal("unexpected error running server")
	}
}

func run() error {
	ctx := context.Background()
	if err := telemetry.SetupTelemetry(ctx); err != nil {
		logrus.WithError(err).Fatal("error initializing telemetry")
	}
	defer telemetry.Shutdown(ctx)

	// Load config
	configPath := config.DefaultConfigPath
	envConfigPath, present := os.LookupEnv(config.ConfigPath.String())
	if present {
		configPath = envConfigPath
	}

	logrus.WithField("path", configPath).Info("loading config from file")
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		logrus.Fatalf("could not instantiate config: %s", err.Error())
	}

	// set up logger
	if logFile := configureLogger(cfg.Log.Level, cfg.Log.Path); logFile != nil {
		defer func(logFile *os.File) {
			if err = logFile.Close(); err != nil {
				logrus.WithError(err).Error("failed to close log file")
			}
		}(logFile)
	}

	// create a channel of buffer size 1 to handle shutdown.
	// buffer's size is 1 in order to ignore any additional ctrl+c
	// spamming.
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	d, err := dht.NewDHT(cfg.DHTConfig.BootstrapPeers)
	if err != nil {
		return util.LoggingErrorMsg(err, "failed to instantiate dht")
	}

	s, err := server.NewServer(cfg, shutdown, d)
	if err != nil {
		return util.LoggingErrorMsg(err, "could not start http services")
	}

	serverErrors := make(chan error, 1)
	go func() {
		logrus.WithField("listen_address", s.Addr).Info("starting listener")
		serverErrors <- s.ListenAndServe()
	}()

	select {
	case err = <-serverErrors:
		return errors.Wrap(err, "server error")
	case sig := <-shutdown:
		logrus.WithField("signal", sig.String()).Info("shutdown signal received")

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()

		if err = s.Shutdown(ctx); err != nil {
			if err = s.Close(); err != nil {
				return err
			}
			return errors.Wrap(err, "main: failed to stop server gracefully")
		}
	}

	return nil
}

// configureLogger configures the logger to logs to the given location and returns a file pointer to a logs
// file that should be closed upon server shutdown
func configureLogger(level, location string) *os.File {
	if level != "" {
		logLevel, err := logrus.ParseLevel(level)
		if err != nil {
			logrus.WithError(err).WithField("level", level).Error("could not parse log level, setting to info")
			logrus.SetLevel(logrus.InfoLevel)
		} else {
			logrus.SetLevel(logLevel)
		}
	}

	// set logs config from config file
	var file *os.File
	var output io.Writer
	output = os.Stdout
	if location != "" {
		now := time.Now()
		filename := filepath.Join(location, fmt.Sprintf("%s-%s-%s.log", config.ServiceName, now.Format(time.DateOnly), strconv.FormatInt(now.Unix(), 10)))
		var err error
		file, err = os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			logrus.WithError(err).Warn("failed to create logs file, using default stdout")
		} else {
			output = io.MultiWriter(os.Stdout, file)
		}
	}

	logrus.SetOutput(output)
	return file
}
