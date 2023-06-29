package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"

	"did-dht/pkg/service/dht"
)

type DHTRouter struct {
	service *dht.Service
}

func NewDHTRouter(service *dht.Service) (*DHTRouter, error) {
	if service == nil {
		return nil, errors.New("service cannot be nil")
	}
	return &DHTRouter{service: service}, nil
}

type AddRecordRequest struct {
	DID      string `json:"did" validate:"required"`
	Endpoint string `json:"endpoint" validate:"required"`
	JWS      string `json:"jws,omitempty"`
}

func (r AddRecordRequest) toServiceRequest() dht.DDTMessage {
	return dht.DDTMessage{
		Record: dht.Record{
			DID:      r.DID,
			Endpoint: r.Endpoint,
			JWS:      r.JWS,
		},
	}
}

type AddRecordResponse struct {
	Message string `json:"message"`
}

// AddRecord godoc
//
//	@Summary		Add a record to the DHT
//	@Description	Add a record to the DHT
//	@Tags			DHT
//	@Accept			json
//	@Produce		json
//	@Param			request	body		AddRecordRequest	true	"Add Record Request"
//	@Success		202		{object}	AddRecordResponse
//	@Failure		400		{string}	string	"Bad request"
//	@Failure		500		{string}	string	"Internal server error"
//	@Router			/v1/dht [put]
func (r *DHTRouter) AddRecord(c *gin.Context) {
	// TODO(gabe): validate before adding record
	var request AddRecordRequest
	if err := Decode(c.Request, &request); err != nil {
		LoggingRespondErrWithMsg(c, err, "invalid add dht record request", http.StatusBadRequest)
		return
	}

	if request.JWS == "" && cfg.

	if err := r.service.PublishRecord(c, request.toServiceRequest()); err != nil {
		LoggingRespondErrWithMsg(c, err, "failed to publish record", http.StatusInternalServerError)
		return
	}

	Respond(c, AddRecordResponse{Message: "success"}, http.StatusAccepted)
}

const (
	DIDParam = "did"
)

type GetRecordResponse struct {
	Record dht.DDTMessage `json:"record"`
}

// ReadRecord godoc
//
//	@Summary		Read a record from the DHT
//	@Description	Read a record from the DHT
//	@Tags			DHT
//	@Accept			json
//	@Produce		json
//	@Param			did	path		string	true	"did to query"
//	@Success		200	{object}	GetRecordResponse
//	@Failure		400	{string}	string	"Bad request"
//	@Failure		500	{string}	string	"Internal server error"
//	@Router			/v1/dht/{did} [get]
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

// ListRecords godoc
//
//	@Summary		List all records from the DHT
//	@Description	List all records from the DHT
//	@Tags			DHT
//	@Accept			json
//	@Produce		json
//	@Success		200	{array}		GetRecordResponse
//	@Failure		500	{string}	string	"Internal server error"
//	@Router			/v1/dht [get]
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

// RemoveRecord godoc
//
//	@Summary		Remove a record from the DHT
//	@Description	Remove a record from the DHT
//	@Tags			DHT
//	@Accept			json
//	@Produce		json
//	@Param			request	body		RemoveRecordRequest	true	"Remove Record Request"
//	@Success		200		{object}	RemoveRecordResponse
//	@Failure		400		{string}	string	"Bad request"
//	@Failure		500		{string}	string	"Internal server error"
//	@Router			/v1/dht [delete]
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
