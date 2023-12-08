package server

import (
	"github.com/gin-gonic/gin"

	"github.com/TBD54566975/did-dht-method/pkg/service"
)

type GatewayRouter struct {
	service *service.GatewayService
}

func NewGatewayRouter(service *service.GatewayService) (*GatewayRouter, error) {
	return &GatewayRouter{service: service}, nil
}

func (r *GatewayRouter) PublishDID(c *gin.Context) {

}

func (r *GatewayRouter) GetDID(c *gin.Context) {

}

func (r *GatewayRouter) GetTypes(c *gin.Context) {

}

func (r *GatewayRouter) GetDIDsForType(c *gin.Context) {

}

func (r *GatewayRouter) GetDifficulty(c *gin.Context) {
	
}
