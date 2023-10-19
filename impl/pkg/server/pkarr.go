package server

import (
	"crypto/ed25519"
	"encoding/base64"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"

	"github.com/TBD54566975/did-dht-method/pkg/service"
)

const (
	IDParam = "id"
)

// PKARRRouter is the router for the PKARR API
type PKARRRouter struct {
	service *service.PKARRService
}

// NewPKARRRouter returns a new instance of the PKARR router
func NewPKARRRouter(service *service.PKARRService) (*PKARRRouter, error) {
	return &PKARRRouter{service: service}, nil
}

// PublishPKARRRequest is the request to publish a PKARR to the DHT
type PublishPKARRRequest struct {
	// V is a base64URL encoded byte array representing the V value for a BEP44 request, up to 1000 bytes
	V string `json:"v" validate:"required"`
	// K is a base64URL encoded byte array representing an ed25519 public key, 32 bytes
	K string `json:"k" validate:"required"`
	// Sig is a base64URL encoded byte array representing an ed25519 signature over the payload, 64 bytes
	Sig string `json:"sig" validate:"required"`
	// Seq is the sequence number for the record
	Seq int64 `json:"seq" validate:"required"`
}

func (p PublishPKARRRequest) toServiceRequest() (*service.PutPKARRRequest, error) {
	encoding := base64.RawURLEncoding
	vBytes, err := encoding.DecodeString(p.V)
	if err != nil {
		return nil, errors.WithMessage(err, "failed to decode v")
	}
	if len(vBytes) > 1000 {
		return nil, errors.New("v must be less than or equal to 1000 bytes")
	}
	kBytes, err := encoding.DecodeString(p.K)
	if err != nil {
		return nil, errors.WithMessage(err, "failed to decode k")
	}
	if len(kBytes) != ed25519.PublicKeySize {
		return nil, errors.New("k must be 32 bytes")
	}
	sigBytes, err := encoding.DecodeString(p.Sig)
	if err != nil {
		return nil, errors.WithMessage(err, "failed to decode sig")
	}
	if len(sigBytes) != ed25519.SignatureSize {
		return nil, errors.New("sig must be 64 bytes")
	}
	return &service.PutPKARRRequest{
		V:   vBytes,
		K:   [32]byte(kBytes),
		Sig: [64]byte(sigBytes),
		Seq: p.Seq,
	}, nil
}

// PublishPKARRResponse is the response to publishing a PKARR to the DHT
type PublishPKARRResponse struct {
	// ID is the ID of the record, which can be resolved from the DHT
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
//	@Router			/v1/pkarr [put]
func (r *PKARRRouter) PublishPKARR(c *gin.Context) {
	var request PublishPKARRRequest
	if err := Decode(c.Request, &request); err != nil {
		LoggingRespondErrWithMsg(c, err, "invalid publish pkarr request", http.StatusBadRequest)
		return
	}

	svcRequest, err := request.toServiceRequest()
	if err != nil {
		LoggingRespondErrWithMsg(c, err, "invalid publish pkarr request", http.StatusBadRequest)
		return
	}

	id, err := r.service.PublishPKARR(c, *svcRequest)
	if err != nil {
		LoggingRespondErrWithMsg(c, err, "failed to publish pkarr request", http.StatusInternalServerError)
		return
	}

	Respond(c, PublishPKARRResponse{ID: id}, http.StatusAccepted)
}

// GetPKARRResponse is the response to getting a PKARR from the DHT
type GetPKARRResponse struct {
	// V is a base64URL encoded byte array representing the V value for a BEP44 request, up to 1000 bytes
	V string `json:"v" validate:"required"`
	// Seq is the sequence number for the record
	Seq int64 `json:"seq" validate:"required"`
}

func fromServiceResponse(resp service.GetPKARRResponse) GetPKARRResponse {
	encoding := base64.RawURLEncoding
	return GetPKARRResponse{
		V:   encoding.EncodeToString(resp.V),
		Seq: resp.Seq,
	}
}

// GetPKARR godoc
//
//	@Summary		Read a PKARR record from the DHT
//	@Description	Read a PKARR record from the DHT
//	@Tags			PKARR
//	@Accept			json
//	@Produce		json
//	@Param			id	path		string	true	"ID"
//	@Success		200	{object}	GetPKARRResponse
//	@Failure		400	{string}	string	"Bad request"
//	@Failure		404	{string}	string	"Not found"
//	@Failure		500	{string}	string	"Internal server error"
//	@Router			/v1/pkarr/{id} [get]
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

	Respond(c, fromServiceResponse(*resp), http.StatusOK)
}
