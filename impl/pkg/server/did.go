package server

import (
	"net/http"

	"github.com/TBD54566975/ssi-sdk/did"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"

	"github.com/TBD54566975/did-dht-method/pkg/dht"
)

const (
	DIDParam = "id"
)

type DIDDHTRouter struct {
	service *dht.Service
}

func NewDIDDHTRouter(service *dht.Service) (*DIDDHTRouter, error) {
	if service == nil {
		return nil, errors.New("service cannot be nil")
	}
	return &DIDDHTRouter{service: service}, nil
}

type PublishDIDRequest struct {
}

func (PublishDIDRequest) toServiceRequest() dht.PublishDIDRequest {
	return dht.PublishDIDRequest{}
}

// PublishDID godoc
//
//	@Summary		Publish a DID to the DHT
//	@Description	Publishes a DID to the DHT
//	@Tags			DIDDHT
//	@Accept			json
//	@Produce		json
//	@Param			request	body	PublishDIDRequest	true	"Publish DID Request"
//	@Success		202
//	@Failure		400	{string}	string	"Bad request"
//	@Failure		500	{string}	string	"Internal server error"
//	@Router			/v1/dht [put]
func (r *DIDDHTRouter) PublishDID(c *gin.Context) {
	var request PublishDIDRequest
	if err := Decode(c.Request, &request); err != nil {
		LoggingRespondErrWithMsg(c, err, "invalid add dht record request", http.StatusBadRequest)
		return
	}

	if err := r.service.PublishDID(c, request.toServiceRequest()); err != nil {
		LoggingRespondErrWithMsg(c, err, "failed to publish record", http.StatusInternalServerError)
		return
	}

	Respond(c, nil, http.StatusAccepted)
}

type GetDIDResponse struct {
	DIDDocument did.Document `json:"did,omitempty"`
}

// GetDID godoc
//
//	@Summary		Read a DID record from the DHT
//	@Description	Read a DID record from the DHT
//	@Tags			DIDDHT
//	@Accept			json
//	@Produce		json
//	@Param			id	path		string	true	"did to request"
//	@Success		200	{object}	GetDIDResponse
//	@Failure		400	{string}	string	"Bad request"
//	@Failure		500	{string}	string	"Internal server error"
//	@Router			/v1/did/{id} [get]
func (r *DIDDHTRouter) GetDID(c *gin.Context) {
	did := GetParam(c, DIDParam)
	if did == nil || *did == "" {
		LoggingRespondErrMsg(c, "missing did param", http.StatusBadRequest)
		return
	}

	resp, err := r.service.GetDID(c, *did)
	if err != nil {
		LoggingRespondErrWithMsg(c, err, "failed to query", http.StatusInternalServerError)
		return
	}

	if resp == nil {
		LoggingRespondErrMsg(c, "did not found", http.StatusNotFound)
		return
	}

	Respond(c, GetDIDResponse{DIDDocument: *resp}, http.StatusOK)
}

type ListDIDsResponse struct {
	DIDDocuments []did.Document `json:"dids,omitempty"`
}

// ListDIDs godoc
//
//	@Summary		List all DIDs from the service
//	@Description	List all DIDs from the service
//	@Tags			DIDDHT
//	@Accept			json
//	@Produce		json
//	@Success		200	{array}		ListDIDsResponse
//	@Failure		500	{string}	string	"Internal server error"
//	@Router			/v1/did [get]
func (r *DIDDHTRouter) ListDIDs(c *gin.Context) {
	resp, err := r.service.ListDIDs(c)
	if err != nil {
		LoggingRespondErrWithMsg(c, err, "failed to list DIDs", http.StatusInternalServerError)
		return
	}

	Respond(c, ListDIDsResponse{DIDDocuments: resp}, http.StatusOK)
}

type DeleteDIDRequest struct {
}

// DeleteDID godoc
//
//	@Summary		Remove a DID from the service
//	@Description	Remove a DID from the service, which stops republishing it to the DHT
//	@Tags			DIDDHT
//	@Accept			json
//	@Produce		json
//	@Param			request	body	DeleteDIDRequest	true	"Delete DID Request"
//	@Success		200
//	@Failure		400	{string}	string	"Bad request"
//	@Failure		500	{string}	string	"Internal server error"
//	@Router			/v1/dht [delete]
func (r *DIDDHTRouter) DeleteDID(c *gin.Context) {
	// TODO(gabe): validate before removing record
	var request DeleteDIDRequest
	if err := Decode(c.Request, &request); err != nil {
		LoggingRespondErrWithMsg(c, err, "invalid remove dht record request", http.StatusBadRequest)
		return
	}

	if err := r.service.DeleteDID(c, ""); err != nil {
		LoggingRespondErrWithMsg(c, err, "failed to remove", http.StatusInternalServerError)
		return
	}

	Respond(c, nil, http.StatusOK)
}
