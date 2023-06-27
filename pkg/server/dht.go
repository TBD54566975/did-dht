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

type AddRecordRequest struct {
	RequesterID   string `json:"requesterId" validate:"required"`
	RequesterName string `json:"requesterName" validate:"required"`
	DID           string `json:"did" validate:"required"`
	Endpoint      string `json:"endpoint" validate:"required"`
}

func (r AddRecordRequest) toServiceRequest() service.DDTMessage {
	return service.DDTMessage{
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

type AddRecordResponse struct {
	Message string `json:"message"`
}

func (r *DHTRouter) AddRecord(c *gin.Context) {
	// TODO(gabe): validate before adding record
	var request AddRecordRequest
	if err := Decode(c.Request, &request); err != nil {
		LoggingRespondErrWithMsg(c, err, "invalid add dht record request", http.StatusBadRequest)
		return
	}

	if err := r.service.GossipRecord(c, request.toServiceRequest()); err != nil {
		LoggingRespondErrWithMsg(c, err, "failed to gossip", http.StatusInternalServerError)
		return
	}

	Respond(c, AddRecordResponse{Message: "success"}, http.StatusOK)
}

const (
	DIDParam = "did"
)

type GetRecordResponse struct {
	Record service.DDTMessage `json:"record"`
}

func (r *DHTRouter) ReadRecord(c *gin.Context) {
	did := GetParam(c, DIDParam)
	if did == nil || *did == "" {
		LoggingRespondErrMsg(c, "missing did param", http.StatusBadRequest)
		return
	}

	resp, err := r.service.QueryRecord(c, *did)
	if err != nil {
		LoggingRespondErrWithMsg(c, err, "failed to query", http.StatusInternalServerError)
		return
	}

	Respond(c, GetRecordResponse{Record: *resp}, http.StatusOK)
}

func (r *DHTRouter) ListRecords(c *gin.Context) {
	resp, err := r.service.ListRecords(c)
	if err != nil {
		LoggingRespondErrWithMsg(c, err, "failed to query", http.StatusInternalServerError)
		return
	}

	Respond(c, resp, http.StatusOK)
}

type RemoveRecordRequest struct {
	Requester string `json:"requester" validate:"required"`
	DID       string `json:"did" validate:"required"`
}

type RemoveRecordResponse struct {
	Message string `json:"message"`
}

func (r *DHTRouter) RemoveRecord(c *gin.Context) {
	// TODO(gabe): validate before removing record
	var request RemoveRecordRequest
	if err := Decode(c.Request, &request); err != nil {
		LoggingRespondErrWithMsg(c, err, "invalid remove dht record request", http.StatusBadRequest)
		return
	}

	if err := r.service.RemoveRecord(c, request.DID); err != nil {
		LoggingRespondErrWithMsg(c, err, "failed to remove", http.StatusInternalServerError)
		return
	}

	Respond(c, RemoveRecordResponse{Message: "success"}, http.StatusOK)
}
