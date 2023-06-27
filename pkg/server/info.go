package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/libp2p/go-libp2p/core/peer"

	"did-dht/pkg/service"
)

type InfoResponse struct {
	ID      string    `json:"id"`
	Address string    `json:"address"`
	Peers   []peer.ID `json:"peers"`
}

// Info godoc
//
//	@Summary		Get info about the service
//	@Description	Get info about the service
//	@Tags			Info
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	InfoResponse
func Info(svc *service.DHTService) gin.HandlerFunc {
	id, addr, peers := svc.Info()
	return func(c *gin.Context) {
		Respond(c, InfoResponse{
			ID:      id,
			Address: addr,
			Peers:   peers,
		}, http.StatusOK)
	}
}
