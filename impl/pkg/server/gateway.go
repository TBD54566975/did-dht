package server

import (
	"net/http"

	"github.com/TBD54566975/ssi-sdk/did"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"

	"github.com/TBD54566975/did-dht-method/internal/util"
	"github.com/TBD54566975/did-dht-method/pkg/service"
)

type GatewayRouter struct {
	service *service.GatewayService
}

func NewGatewayRouter(service *service.GatewayService) (*GatewayRouter, error) {
	return &GatewayRouter{service: service}, nil
}

// PublishDIDRequest represents a request to publish a DID
type PublishDIDRequest struct {
	Sig            string `json:"sig" validate:"required"`
	Seq            int64  `json:"seq" validate:"required"`
	V              string `json:"v" validate:"required"`
	RetentionProof string `json:"retention_proof,omitempty"`
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

// PublishDID godoc
// @Summary		Publish a DID document
// @Description	Publish a DID document to the DHT
// @Tags		DID
// @Accept		json
// @Param		id		path	string	true	"ID of the record to get"
// @Success		202 {object}    PublishDIDRequest
// @Failure		400	{string}	string	"Invalid request body"
// @Failure		401	{string}	string	"Invalid signature"
// @Failure		409	{string}	string	"DID already exists with a higher sequence number"
// @Failure		500	{string}	string	"Internal server error"
// @Router		/dids/{id} [put]
// TODO(gabe) support historical document storage https://github.com/TBD54566975/did-dht-method/issues/16
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
	if err := r.service.PublishDID(c, req.toServiceRequest(*id)); err != nil {
		if errors.Is(err, &util.InvalidSignatureError{}) {
			Respond(c, nil, http.StatusUnauthorized)
			return
		}

		if errors.Is(err, &util.HigherSequenceNumberError{}) {
			Respond(c, nil, http.StatusConflict)
			return
		}

		LoggingRespondErrWithMsg(c, err, "failed to publish did", http.StatusInternalServerError)
	}

	Respond(c, nil, http.StatusAccepted)
}

// GetDIDResponse represents a response containing a DID document, types, and sequence numbers.
type GetDIDResponse struct {
	DID             did.Document `json:"did" validate:"required"`
	Types           []int        `json:"types,omitempty"`
	SequenceNumbers []int        `json:"sequence_numbers,omitempty"`
}

// GetDID godoc
// @Summary		Get a DID document
// @Description	Get a DID document
// @Tags		DID
// @Accept		json
// @Param		id		path	string	true	"ID of the record to get"
// @Success		200 {object}    GetDIDResponse
// @Failure		400	{string}	string	"Invalid request"
// @Failure		404	{string}	string	"DID not found"
// @Failure		500	{string}	string	"Internal server error"
// @Router		/dids/{id} [get]
// TODO(gabe) support historical queries https://github.com/TBD54566975/did-dht-method/issues/16
func (r *GatewayRouter) GetDID(c *gin.Context) {
	id := GetParam(c, IDParam)
	if id == nil || *id == "" {
		LoggingRespondErrMsg(c, "missing id param", http.StatusBadRequest)
		return
	}

	resp, err := r.service.GetDID(*id)
	if err != nil {
		LoggingRespondErrWithMsg(c, err, "failed to get did", http.StatusInternalServerError)
		return
	}

	if resp == nil {
		LoggingRespondErrMsg(c, "did not found", http.StatusNotFound)
		return
	}

	Respond(c, GetDIDResponse(*resp), http.StatusOK)
}

// GetTypesResponse represents a response containing a list of supported types and their names.
type GetTypesResponse struct {
	Types []service.TypeMapping `json:"types,omitempty"`
}

// GetTypes godoc
// @Summary		Get a list of supported types
// @Description	Get a list of supported types
// @Tags		DID
// @Accept		json
// @Success		200 {object}    GetTypesResponse
// @Failure		404	{string}	string	"Type indexing is not supported by this gateway"
// @Router		/dids/types [get]
func (r *GatewayRouter) GetTypes(c *gin.Context) {
	resp, err := r.service.GetTypes()
	if err != nil {
		LoggingRespondErrWithMsg(c, err, "failed to get types", http.StatusInternalServerError)
		return
	}

	if resp == nil {
		LoggingRespondErrMsg(c, "types not supported", http.StatusNotFound)
		return
	}

	Respond(c, GetTypesResponse(*resp), http.StatusOK)
}

// GetDIDsForTypeResponse represents a response containing a list of DIDs for a given type.
type GetDIDsForTypeResponse struct {
	DIDs []string `json:"dids,omitempty"`
}

// GetDIDsForType godoc
// @Summary		Get a list of DIDs for a given type
// @Description	Get a list of DIDs for a given type
// @Tags		DID
// @Accept		json
// @Success		200 {object}    GetDIDsForTypeResponse
// @Failure		404	{string}	string	"Type not found"
// @Failure		500	{string}	string	"Internal server error"
// @Router		/dids/types/{id} [get]
func (r *GatewayRouter) GetDIDsForType(c *gin.Context) {
	id := GetParam(c, IDParam)
	if id == nil || *id == "" {
		LoggingRespondErrMsg(c, "missing id param", http.StatusBadRequest)
		return
	}

	resp, err := r.service.GetDIDsForType(service.GetDIDsForTypeRequest{Type: *id})
	if err != nil {
		if errors.Is(err, &util.TypeNotFoundError{}) {
			LoggingRespondErrMsg(c, "type not found", http.StatusNotFound)
			return
		}

		LoggingRespondErrWithMsg(c, err, "failed to get dids for type", http.StatusInternalServerError)
		return
	}

	Respond(c, GetDIDsForTypeResponse(*resp), http.StatusOK)
}

// GetDifficultyResponse represents a response containing the current difficulty for the gateway's retention proof feature.
type GetDifficultyResponse struct {
	Hash       string `json:"hash" validate:"required"`
	Difficulty int    `json:"difficulty" validate:"required"`
}

// GetDifficulty godoc
// @Summary		Get the current difficulty for the gateway's retention proof feature
// @Description	Get the current difficulty for the gateway's retention proof feature
// @Tags		DID
// @Accept		json
// @Success		200 {object}    int
// @Failure		404	{string}	string	"Retention proofs are not supported by this gateway"
// @Failure		500	{string}	string	"Internal server error"
// @Router		/difficulty [get]
func (r *GatewayRouter) GetDifficulty(c *gin.Context) {
	resp, err := r.service.GetDifficulty()
	if err != nil {
		LoggingRespondErrWithMsg(c, err, "failed to get difficulty", http.StatusInternalServerError)
		return
	}

	if resp == nil {
		LoggingRespondErrMsg(c, "retention proofs not supported", http.StatusNotFound)
		return
	}

	Respond(c, GetDifficultyResponse(*resp), http.StatusOK)
}
