package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"

	"github.com/TBD54566975/did-dht-method/pkg/service"
)

type GatewayRouter struct {
	service *service.GatewayService
}

func NewGatewayRouter(service *service.GatewayService) (*GatewayRouter, error) {
	return &GatewayRouter{service: service}, nil
}

type PublishDIDRequest struct {
	Sig            string `json:"sig" validate:"required"`
	Seq            int    `json:"seq" validate:"required"`
	V              string `json:"v" validate:"required"`
	RetentionProof int    `json:"retention_proof,omitempty"`
}

func (p PublishDIDRequest) toServiceRequest(did string) service.PublishDIDRequest {
	return service.PublishDIDRequest{
		DID:            did,
		Sig:            p.Sig,
		Seq:            p.Seq,
		V:              p.V,
		RetentionProof: p.RetentionProof,
	}
}

func (r *GatewayRouter) PublishDID(c *gin.Context) {
	id := GetParam(c, IDParam)
	if id == nil || *id == "" {
		LoggingRespondErrMsg(c, "missing id param", http.StatusBadRequest)
		return
	}

	var req PublishDIDRequest
	if err := Decode(c.Request, &req); err != nil {
		LoggingRespondErrWithMsg(c, err, "failed to decode request", http.StatusBadRequest)
		return
	}

	// three possible errors
	// 1. invalid signature
	// 2. did already exists with a higher sequence number
	// 3. internal service error
	if err := r.service.PublishDID(req.toServiceRequest(*id)); err != nil {
		if errors.Is(err, &InvalidSignatureError{}) {
			Respond(c, nil, http.StatusUnauthorized)
			return
		}

		if errors.Is(err, &HigherSequenceNumberError{}) {
			Respond(c, nil, http.StatusConflict)
			return
		}

		LoggingRespondErrWithMsg(c, err, "failed to publish did", http.StatusInternalServerError)
	}

	Respond(c, nil, http.StatusAccepted)
}

func (r *GatewayRouter) GetDID(c *gin.Context) {
	id := GetParam(c, IDParam)
	if id == nil || *id == "" {
		LoggingRespondErrMsg(c, "missing id param", http.StatusBadRequest)
		return
	}
}

func (r *GatewayRouter) GetTypes(c *gin.Context) {

}

func (r *GatewayRouter) GetDIDsForType(c *gin.Context) {

}

func (r *GatewayRouter) GetDifficulty(c *gin.Context) {

}
