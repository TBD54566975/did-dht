package server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/goccy/go-json"
	"github.com/sirupsen/logrus"

	"did-dht/config"
	"did-dht/pkg/service"
)

type Server struct {
	*http.Server
	cfg      *config.Config
	svc      *service.DIDDHTService
	handler  *gin.Engine
	shutdown chan os.Signal
}

// NewServer returns a new instance of Server with the given db and host.
func NewServer(cfg *config.Config, shutdown chan os.Signal) (*Server, error) {
	// set up server prerequisites
	setupLogger(cfg.LogLevel)
	handler := setupHandler(cfg.Environment)
	ddtSvc, err := service.NewDIDDHTService(cfg)
	if err != nil {
		logrus.WithError(err).Error("could not instantiate did dht service")
		return nil, err
	}
	if err = ddtSvc.Start(context.Background()); err != nil {
		logrus.WithError(err).Error("could not start did dht service")
		return nil, err
	}

	handler.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})

	handler.PUT("/msg", func(c *gin.Context) {
		decoder := json.NewDecoder(c.Request.Body)
		decoder.DisallowUnknownFields()

		var msg string
		if err = decoder.Decode(&msg); err != nil {
			logrus.WithError(err).Error("could not decode request body")
			c.String(http.StatusBadRequest, "could not decode request body")
			return
		}

		if err = ddtSvc.Gossip(c.Request.Context(), msg); err != nil {
			logrus.WithError(err).Error("could not send message")
			c.String(http.StatusInternalServerError, "could not send message")
			return
		}

		c.String(http.StatusOK, "sent")
	})

	handler.GET("/info", func(c *gin.Context) {
		id, addr, peers := ddtSvc.Info()
		c.JSON(http.StatusOK, gin.H{
			"id":      id,
			"address": addr,
			"peers":   peers,
		})
	})

	handler.GET("/got", func(c *gin.Context) {

	})

	return &Server{
		Server: &http.Server{
			Addr:              fmt.Sprintf("%s:%d", cfg.APIHost, cfg.APIPort),
			Handler:           handler,
			ReadTimeout:       time.Second * 5,
			ReadHeaderTimeout: time.Second * 5,
			WriteTimeout:      time.Second * 5,
		},
		cfg:      cfg,
		svc:      ddtSvc,
		handler:  handler,
		shutdown: shutdown,
	}, nil
}

func setupLogger(level string) {
	logrus.SetFormatter(&logrus.JSONFormatter{
		DisableTimestamp: false,
		PrettyPrint:      true,
	})
	logrus.SetReportCaller(true)

	logLevel, err := logrus.ParseLevel(level)
	if err != nil {
		logrus.WithError(err).Errorf("could not parse log level<%s>, setting to info", level)
		logrus.SetLevel(logrus.InfoLevel)
	} else {
		logrus.SetLevel(logLevel)
	}
}

func setupHandler(env config.Environment) *gin.Engine {
	gin.ForceConsoleColor()
	middlewares := gin.HandlersChain{
		gin.Recovery(),
		gin.Logger(),
		gin.ErrorLogger(),
	}
	handler := gin.New()
	handler.Use(middlewares...)
	switch env {
	case config.EnvironmentDev:
		gin.SetMode(gin.DebugMode)
	case config.EnvironmentTest:
		gin.SetMode(gin.TestMode)
	case config.EnvironmentProd:
		gin.SetMode(gin.ReleaseMode)
	}
	return handler
}
