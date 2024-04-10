package main

import (
	"context"
	"os"
	"os/signal"
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
	logrus.SetFormatter(&logrus.JSONFormatter{PrettyPrint: true})
	logrus.SetReportCaller(true)
	logrus.WithField("version", config.Version).Info("starting up")

	if err := run(); err != nil {
		logrus.WithError(err).Fatal("unexpected error running server")
	}
}

func run() error {
	ctx := context.Background()

	// Load config
	configPath := config.DefaultConfigPath
	envConfigPath, present := os.LookupEnv(config.ConfigPath.String())
	if present {
		configPath = envConfigPath
	}

	logrus.WithContext(ctx).WithField("path", configPath).Info("loading config from file")
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		logrus.WithContext(ctx).Fatalf("could not instantiate config: %s", err.Error())
	}

	// set up telemetry
	if cfg.ServerConfig.Telemetry {
		if err = telemetry.SetupTelemetry(ctx); err != nil {
			logrus.WithContext(ctx).WithError(err).Fatal("error initializing telemetry")
		}
		defer telemetry.Shutdown(ctx)
	}

	// set up logger
	configureLogger(cfg.Log.Level)

	// create a channel of buffer size 1 to handle shutdown.
	// buffer's size is 1 in order to ignore any additional ctrl+c spamming.
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	d, err := dht.NewDHT(cfg.DHTConfig.BootstrapPeers)
	if err != nil {
		return util.LoggingCtxErrorMsg(ctx, err, "failed to instantiate dht")
	}

	s, err := server.NewServer(cfg, shutdown, d)
	if err != nil {
		return util.LoggingCtxErrorMsg(ctx, err, "could not start http services")
	}

	serverErrors := make(chan error, 1)
	go func() {
		logrus.WithContext(ctx).WithField("listen_address", s.Addr).Info("starting listener")
		serverErrors <- s.ListenAndServe()
	}()

	select {
	case err = <-serverErrors:
		return errors.Wrap(err, "server error")
	case sig := <-shutdown:
		logrus.WithContext(ctx).WithField("signal", sig.String()).Info("shutdown signal received")

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

// configureLogger configures the logger
func configureLogger(level string) {
	if level != "" {
		logLevel, err := logrus.ParseLevel(level)
		if err != nil {
			logrus.WithError(err).WithField("level", level).Error("could not parse log level, setting to info")
			logrus.SetLevel(logrus.InfoLevel)
		} else {
			logrus.SetLevel(logLevel)
		}
	}
}
