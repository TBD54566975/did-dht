package server

import (
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"

	"did-dht/pkg/db"
)

type Server struct {
	*http.Server
	db       db.Storage
	handler  *gin.Engine
	shutdown chan os.Signal
}

// NewServer returns a new instance of Server with the given db and host.
func NewServer(db db.Storage, host string, shutdown chan os.Signal) (*Server, error) {
	if !db.IsOpen() {
		return nil, errors.New("db must be open")
	}

	gin.ForceConsoleColor()
	middlewares := gin.HandlersChain{
		gin.Recovery(),
		gin.Logger(),
		gin.ErrorLogger(),
	}
	handler := gin.New()
	handler.Use(middlewares...)

	return &Server{
		Server: &http.Server{
			Addr:              host,
			Handler:           handler,
			ReadTimeout:       time.Second * 5,
			ReadHeaderTimeout: time.Second * 5,
			WriteTimeout:      time.Second * 5,
		},
		db:       db,
		handler:  handler,
		shutdown: shutdown,
	}, nil
}
