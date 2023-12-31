package server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/TBD54566975/ssi-sdk/util"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	swaggerfiles "github.com/swaggo/files"
	ginswagger "github.com/swaggo/gin-swagger"

	"did-dht/config"
	"did-dht/docs"
	"did-dht/pkg/service/dht"
)

type Server struct {
	*http.Server
	handler *gin.Engine

	shutdown chan os.Signal

	cfg *config.Config
	svc *dht.Service
}

// NewServer returns a new instance of Server with the given db and host.
func NewServer(cfg *config.Config, shutdown chan os.Signal) (*Server, error) {
	// set up server prerequisites
	setupLogger(cfg.ServerConfig.LogLevel)
	handler := setupHandler(cfg.ServerConfig.Environment)
	ddtSvc, err := dht.NewService(cfg)
	if err != nil {
		return nil, util.LoggingErrorMsg(err, "could not instantiate did dht service")
	}
	if err = ddtSvc.Start(context.Background()); err != nil {
		return nil, util.LoggingErrorMsg(err, "could not start did dht service")
	}

	handler.GET("/health", Health)
	handler.GET("/info", Info(ddtSvc))

	// set up swagger
	docs.SwaggerInfo.Host = fmt.Sprintf("%s:%d", cfg.ServerConfig.APIHost, cfg.ServerConfig.APIPort)
	docs.SwaggerInfo.Version = "0.0.1"
	handler.StaticFile("swagger.yaml", "./docs/swagger.yaml")
	handler.GET("/swagger/*any", ginswagger.WrapHandler(swaggerfiles.Handler, ginswagger.URL("/swagger.yaml")))

	v1 := handler.Group("/v1")
	if err = DHTAPI(v1, ddtSvc); err != nil {
		return nil, util.LoggingErrorMsg(err, "could not setup dht api")
	}

	// TODO(gabe): add more routes here

	return &Server{
		Server: &http.Server{
			Addr:              fmt.Sprintf("%s:%d", cfg.ServerConfig.APIHost, cfg.ServerConfig.APIPort),
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

func DHTAPI(rg *gin.RouterGroup, service *dht.Service) error {
	dhtRouter, err := NewDHTRouter(service)
	if err != nil {
		return util.LoggingErrorMsg(err, "could not instantiate dht router")
	}

	dhtAPI := rg.Group("/dht")
	dhtAPI.PUT("", dhtRouter.AddRecord)
	dhtAPI.GET("", dhtRouter.ListRecords)
	dhtAPI.GET("/:did", dhtRouter.ReadRecord)
	dhtAPI.DELETE("/:did", dhtRouter.RemoveRecord)
	return nil
}
