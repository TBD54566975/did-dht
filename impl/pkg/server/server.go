package server

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/TBD54566975/ssi-sdk/util"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	swaggerfiles "github.com/swaggo/files"
	ginswagger "github.com/swaggo/gin-swagger"

	"github.com/TBD54566975/did-dht-method/config"
	"github.com/TBD54566975/did-dht-method/docs"
	"github.com/TBD54566975/did-dht-method/pkg/service"
	"github.com/TBD54566975/did-dht-method/pkg/storage"
)

const (
	IDParam string = "id"
)

type Server struct {
	*http.Server
	handler *gin.Engine

	shutdown chan os.Signal

	cfg *config.Config
	svc *service.PkarrService
}

// NewServer returns a new instance of Server with the given db and host.
func NewServer(cfg *config.Config, shutdown chan os.Signal) (*Server, error) {
	// set up server prerequisites
	setupLogger(cfg.ServerConfig.LogLevel)
	handler := setupHandler(cfg.ServerConfig.Environment)

	db, err := storage.NewStorage(cfg.ServerConfig.DBFile)
	if err != nil {
		return nil, util.LoggingErrorMsg(err, "failed to instantiate storage")
	}

	pkarrService, err := service.NewPkarrService(cfg, db)
	if err != nil {
		return nil, util.LoggingErrorMsg(err, "could not instantiate pkarr service")
	}

	handler.GET("/health", Health)

	// set up swagger
	docs.SwaggerInfo.Host = fmt.Sprintf("%s:%d", cfg.ServerConfig.APIHost, cfg.ServerConfig.APIPort)
	docs.SwaggerInfo.Version = "0.0.1"
	handler.StaticFile("swagger.yaml", "./docs/swagger.yaml")
	handler.GET("/swagger/*any", ginswagger.WrapHandler(swaggerfiles.Handler, ginswagger.URL("/swagger.yaml")))

	// root relay API
	if err = PkarrAPI(&handler.RouterGroup, pkarrService); err != nil {
		return nil, util.LoggingErrorMsg(err, "could not setup pkarr API")
	}
	return &Server{
		Server: &http.Server{
			Addr:              fmt.Sprintf("%s:%d", cfg.ServerConfig.APIHost, cfg.ServerConfig.APIPort),
			Handler:           handler,
			ReadTimeout:       time.Second * 15,
			ReadHeaderTimeout: time.Second * 15,
			WriteTimeout:      time.Second * 15,
		},
		cfg:      cfg,
		svc:      pkarrService,
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
		CORS(),
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

// PkarrAPI sets up the relay API routes according to https://github.com/Nuhvi/pkarr/blob/main/design/relays.md
func PkarrAPI(rg *gin.RouterGroup, service *service.PkarrService) error {
	relayRouter, err := NewPKARRRouter(service)
	if err != nil {
		return util.LoggingErrorMsg(err, "could not instantiate relay router")
	}

	rg.PUT("/:id", relayRouter.PutRecord)
	rg.GET("/:id", relayRouter.GetRecord)
	return nil
}

// func GatewayAPI(rg *gin.RouterGroup, service *service.PkarrService) error {
// 	gatewayRouter, err := NewGatewayRouter(service)
// 	if err != nil {
// 		return util.LoggingErrorMsg(err, "could not instantiate gateway router")
// 	}
//
// 	rg.GET("/did", gatewayRouter.GetRecord)
// 	return nil
// }
