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
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"

	"github.com/TBD54566975/did-dht-method/config"
	"github.com/TBD54566975/did-dht-method/pkg/dht"
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
	svc *service.DHTService
}

// NewServer returns a new instance of Server with the given db and host.
func NewServer(cfg *config.Config, shutdown chan os.Signal, d *dht.DHT) (*Server, error) {
	// set up server prerequisites
	handler := setupHandler(cfg.ServerConfig.Environment)

	db, err := storage.NewStorage(cfg.ServerConfig.StorageURI)
	if err != nil {
		return nil, util.LoggingErrorMsg(err, "failed to instantiate storage")
	}

	recordCnt, err := db.RecordCount(context.Background())
	if err != nil {
		logrus.WithError(err).Error("failed to get record count")
	} else {
		logrus.WithField("record_count", recordCnt).Info("storage instantiated with record count")
	}

	dhtService, err := service.NewDHTService(cfg, db, d)
	if err != nil {
		return nil, util.LoggingErrorMsg(err, "could not instantiate the dht service")
	}

	handler.GET("/health", Health)

	// set up swagger
	handler.StaticFile("swagger.yaml", "./docs/swagger.yaml")
	handler.GET("/swagger/*any", ginswagger.WrapHandler(swaggerfiles.Handler, ginswagger.URL("/swagger.yaml")))

	// root relay API
	if err = DHTAPI(&handler.RouterGroup, dhtService); err != nil {
		return nil, util.LoggingErrorMsg(err, "could not setup the dht API")
	}
	return &Server{
		Server: &http.Server{
			Addr:              fmt.Sprintf("%s:%d", cfg.ServerConfig.APIHost, cfg.ServerConfig.APIPort),
			Handler:           handler,
			ReadTimeout:       time.Second * 15,
			ReadHeaderTimeout: time.Second * 10,
			WriteTimeout:      time.Second * 10,
			MaxHeaderBytes:    1 << 20,
		},
		cfg:      cfg,
		svc:      dhtService,
		handler:  handler,
		shutdown: shutdown,
	}, nil
}

func setupHandler(env config.Environment) *gin.Engine {
	gin.ForceConsoleColor()
	middlewares := gin.HandlersChain{
		otelgin.Middleware(config.ServiceName),
		gin.Recovery(),
		gin.ErrorLogger(),
		CORS(),
		logger(logrus.StandardLogger()),
	}
	logrus.WithField("environment", env).Info("configuring server for environment")
	switch env {
	case config.EnvironmentDev:
		gin.SetMode(gin.DebugMode)
	case config.EnvironmentTest:
		gin.SetMode(gin.TestMode)
	case config.EnvironmentProd:
		gin.SetMode(gin.ReleaseMode)
	}
	handler := gin.New()
	handler.Use(middlewares...)
	return handler
}

// DHTAPI sets up the relay API routes according to the spec https://did-dht.com/#gateway-api
func DHTAPI(rg *gin.RouterGroup, service *service.DHTService) error {
	dhtRouter, err := NewDHTRouter(service)
	if err != nil {
		return util.LoggingErrorMsg(err, "could not instantiate dht router")
	}

	rg.PUT("/:id", dhtRouter.PutRecord)
	rg.GET("/:id", dhtRouter.GetRecord)
	return nil
}
