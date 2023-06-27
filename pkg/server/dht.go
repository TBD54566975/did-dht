package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"

	"did-dht/pkg/service"
)

type DHTRouter struct {
	service *service.DHTService
}

func NewDHTRouter(service *service.DHTService) (*DHTRouter, error) {
	if service == nil {
		return nil, errors.New("service cannot be nil")
	}
	return &DHTRouter{service: service}, nil
}

type AddDHTRecordRequest struct {
	RequesterID   string `json:"requesterId" validate:"required"`
	RequesterName string `json:"requesterName" validate:"required"`
	DID           string `json:"did" validate:"required"`
	Endpoint      string `json:"endpoint" validate:"required"`
}

func (r AddDHTRecordRequest) toServiceRequest() service.DHTMessage {
	return service.DHTMessage{
		Requester: service.Requester{
			ID:   r.RequesterID,
			Name: r.RequesterName,
		},
		Record: service.Record{
			DID:      r.DID,
			Endpoint: r.Endpoint,
		},
	}
}

func (r *DHTRouter) AddDHTRecord(c *gin.Context) {
	var request AddDHTRecordRequest
	if err := Decode(c.Request, &request); err != nil {
		LoggingRespondErrWithMsg(c, err, "invalid add dht record request", http.StatusBadRequest)
		return
	}

	if err := r.service.Gossip(c, request.toServiceRequest()); err != nil {
		LoggingRespondErrWithMsg(c, err, "failed to gossip", http.StatusInternalServerError)
		return
	}

	Respond(c, "success", http.StatusOK)
}

type ReadDHTRecordRequest struct {
	DID string `json:"did" validate:"required"`
}

type ReadDHTRecordsRequest struct{}

type RemoveDHTRecordRequest struct {
	Requester string `json:"requester" validate:"required"`
	DID       string `json:"did" validate:"required"`
}
