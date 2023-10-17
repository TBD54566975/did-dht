package server

import (
	"github.com/gin-gonic/gin"

	"github.com/TBD54566975/did-dht-method/pkg/service"
)

type PKARRRouter struct {
	service *service.PKARRService
}

func NewPKARRRouter(service *service.PKARRService) (*PKARRRouter, error) {
	return &PKARRRouter{service: service}, nil
}

func (r *PKARRRouter) PublishPKARR(c *gin.Context) {

}

func (r *PKARRRouter) GetPKARR(c *gin.Context) {

}
