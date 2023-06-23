package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"did-dht/config"
	"did-dht/pkg/server"
)

func main() {
	logrus.Info("Starting up...")

	if err := run(); err != nil {
		logrus.Fatalf("main: error: %s", err.Error())
	}
}

func run() error {
	// create a channel of buffer size 1 to handle shutdown.
	// buffer's size is 1 in order to ignore any additional ctrl+c
	// spamming.
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	cfg := config.GetDefaultConfig()
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
