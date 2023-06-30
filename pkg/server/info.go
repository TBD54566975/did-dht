package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/libp2p/go-libp2p/core/peer"

	"did-dht/pkg/service/dht"
)

type InfoResponse struct {
	ID      string    `json:"id"`
	Address string    `json:"address"`
	Topics  []string  `json:"topics"`
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
func Info(svc *dht.Service) gin.HandlerFunc {
	id, addr, topics, peers := svc.Info()
	return func(c *gin.Context) {
		Respond(c, InfoResponse{
			ID:      id,
			Address: addr,
			Topics:  topics,
			Peers:   peers,
		}, http.StatusOK)
	}
}
