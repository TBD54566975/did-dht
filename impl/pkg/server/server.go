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
	ginlogrus "github.com/toorop/gin-logrus"
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
	svc *service.PkarrService
}

// NewServer returns a new instance of Server with the given db and host.
func NewServer(cfg *config.Config, shutdown chan os.Signal, d *dht.DHT) (*Server, error) {
	// set up server prerequisites
	handler := setupHandler(cfg.ServerConfig.Environment)

	db, err := storage.NewStorage(cfg.ServerConfig.StorageURI)
	if err != nil {
		return nil, util.LoggingErrorMsg(err, "failed to instantiate storage")
	}

	pkarrService, err := service.NewPkarrService(cfg, db, d)
	if err != nil {
		return nil, util.LoggingErrorMsg(err, "could not instantiate pkarr service")
	}

	handler.GET("/health", Health)

	// set up swagger
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

func setupHandler(env config.Environment) *gin.Engine {
	middlewares := gin.HandlersChain{
		gin.Recovery(),
		ginlogrus.Logger(logrus.StandardLogger()),
		gin.ErrorLogger(),
		otelgin.Middleware(config.ServiceName),
		CORS(),
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

// PkarrAPI sets up the relay API routes according to https://github.com/Nuhvi/pkarr/blob/main/design/relays.md
func PkarrAPI(rg *gin.RouterGroup, service *service.PkarrService) error {
	relayRouter, err := NewPkarrRouter(service)
	if err != nil {
		return util.LoggingErrorMsg(err, "could not instantiate relay router")
	}

	rg.PUT("/:id", relayRouter.PutRecord)
	rg.GET("/:id", relayRouter.GetRecord)
	return nil
}
