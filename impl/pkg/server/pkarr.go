package server

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/TBD54566975/did-dht-method/pkg/service"
)

const (
	IDParam = "id"
)

type PKARRRouter struct {
	service *service.PKARRService
}

func NewPKARRRouter(service *service.PKARRService) (*PKARRRouter, error) {
	return &PKARRRouter{service: service}, nil
}

type PublishPKARRRequest struct {
	V   []byte   `json:"v" validate:"required"`
	Sig [64]byte `json:"sig" validate:"required"`
	Seq int64    `json:"seq" validate:"required"`
}

type PublishPKARRResponse struct {
	ID string `json:"id"`
}

// PublishPKARR godoc
//
//	@Summary		Publish a PKARR to the DHT
//	@Description	Publishes a PKARR to the DHT
//	@Tags			PKARR
//	@Accept			json
//	@Produce		json
//	@Param			request	body	PublishPKARRRequest	true	"Publish PKARR Request"
//	@Success		202
//	@Failure		400	{string}	string	"Bad request"
//	@Failure		500	{string}	string	"Internal server error"
//	@Router			/{id} [put]
func (r *PKARRRouter) PublishPKARR(c *gin.Context) {
	var request PublishPKARRRequest
	if err := Decode(c.Request, &request); err != nil {
		LoggingRespondErrWithMsg(c, err, "invalid publish pkarr request", http.StatusBadRequest)
		return
	}

	id, err := r.service.PublishPKARR(c, service.PutPKARRRequest(request))
	if err != nil {
		LoggingRespondErrWithMsg(c, err, "failed to publish pkarr request", http.StatusInternalServerError)
		return
	}

	Respond(c, PublishPKARRResponse{ID: id}, http.StatusAccepted)
}

type GetPKARRResponse struct {
	V   []byte   `json:"v" validate:"required"`
	Sig [64]byte `json:"sig" validate:"required"`
	Seq int64    `json:"seq" validate:"required"`
}

// GetPKARR godoc
//
//	@Summary		Read a PKARR record from the DHT
//	@Description	Read a PKARR record from the DHT
//	@Tags			PKARR
//	@Accept			json
//	@Produce		json
//	@Param			id		path	string	true	"ID"
//	@Success		200		{object}	GetPKARRResponse
//	@Failure		400		{string}	string	"Bad request"
//	@Failure		404		{string}	string	"Not found"
//	@Failure		500		{string}	string	"Internal server error"
//	@Router			/{id} [get]
func (r *PKARRRouter) GetPKARR(c *gin.Context) {
	id := GetParam(c, IDParam)
	if id == nil || *id == "" {
		LoggingRespondErrMsg(c, "missing id param", http.StatusBadRequest)
		return
	}

	resp, err := r.service.GetPKARR(c, *id)
	if err != nil {
		LoggingRespondErrWithMsg(c, err, "failed to get pkarr", http.StatusInternalServerError)
		return
	}

	if resp == nil {
		LoggingRespondErrMsg(c, "pkarr not found", http.StatusNotFound)
		return
	}

	Respond(c, GetPKARRResponse(*resp), http.StatusOK)
}
