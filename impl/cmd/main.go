package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/TBD54566975/did-dht-method/config"
	"github.com/TBD54566975/did-dht-method/pkg/server"
)

// main godoc
//
//	@title			DID DHT Service API
//	@description	The DID DHT Service
//	@contact.name	TBD
//	@contact.url	https://github.com/TBD54566975/did-dht-method/issues
//	@contact.email	tbd-developer@squareup.com
//	@license.name	Apache 2.0
//	@license.url	http://www.apache.org/licenses/LICENSE-2.0.html
//	@host			{{.Server.APIHost}}
//	@version		{{.SVN}}
func main() {
	logrus.Info("Starting up...")

	if err := run(); err != nil {
		logrus.Fatalf("main: error: %s", err.Error())
	}
}

func run() error {
	configPath := config.DefaultConfigPath
	envConfigPath, present := os.LookupEnv(config.ConfigPath.String())
	if present {
		logrus.Infof("loading config from env var path: %s", envConfigPath)
		configPath = envConfigPath
	}
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		logrus.Fatalf("could not instantiate config: %s", err.Error())
	}

	// create a channel of buffer size 1 to handle shutdown.
	// buffer's size is 1 in order to ignore any additional ctrl+c
	// spamming.
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	s, err := server.NewServer(cfg, shutdown)
	if err != nil {
		logrus.WithError(err).Error("could not start http services")
		return err
	}

	serverErrors := make(chan error, 1)
	go func() {
		logrus.Infof("main: server started and listening on -> %s", s.Addr)
		serverErrors <- s.ListenAndServe()
	}()

	select {
	case err = <-serverErrors:
		return errors.Wrap(err, "server error")
	case sig := <-shutdown:
		logrus.Infof("main: shutdown signal received -> %v", sig)

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
