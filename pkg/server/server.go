package server

import (
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
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
func NewServer(cfg config.Config, shutdown chan os.Signal) (*Server, error) {
	// set up server prerequisites
	setupLogger(cfg.LogLevel)
	handler := setupHandler(cfg.Environment)
	_, err := service.NewDIDDHTService(cfg)
	if err != nil {
		logrus.WithError(err).Error("could not instantiate did dht service")
		return nil, err
	}

	handler.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})

	return &Server{
		Server: &http.Server{
			Addr:              cfg.APIHost,
			Handler:           handler,
			ReadTimeout:       time.Second * 5,
			ReadHeaderTimeout: time.Second * 5,
			WriteTimeout:      time.Second * 5,
		},
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
