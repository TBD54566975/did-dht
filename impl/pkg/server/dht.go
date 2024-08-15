package server

import (
	"crypto/ed25519"
	"encoding/binary"
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"

	"github.com/TBD54566975/did-dht/internal/util"
	"github.com/TBD54566975/did-dht/pkg/dht"
	"github.com/TBD54566975/did-dht/pkg/service"
	"github.com/TBD54566975/did-dht/pkg/telemetry"
)

// DHTRouter is the router for the DHT API
type DHTRouter struct {
	service *service.DHTService
}

// NewDHTRouter returns a new instance of the DHT router
func NewDHTRouter(service *service.DHTService) (*DHTRouter, error) {
	return &DHTRouter{service: service}, nil
}

// GetRecord godoc
//
//	@Summary		GetRecord a BEP44 DNS record from the DHT
//	@Description	GetRecord a BEP44 DNS record from the DHT
//	@Tags			DHT
//	@Accept			octet-stream
//	@Produce		octet-stream
//	@Param			id	path		string	true	"ID to get"
//	@Success		200	{array}		byte	"64 bytes sig, 8 bytes u64 big-endian seq, 0-1000 bytes of v."
//	@Failure		400	{string}	string	"Bad request"
//	@Failure		404	{string}	string	"Not found"
//	@Failure		500	{string}	string	"Internal server error"
//	@Router			/{id} [get]
func (r *DHTRouter) GetRecord(c *gin.Context) {
	ctx, span := telemetry.GetTracer().Start(c, "DHTHTTP.GetRecord")
	defer span.End()

	id := GetParam(c, IDParam)
	if id == nil || *id == "" {
		LoggingRespondErrMsg(c, "missing id param", http.StatusBadRequest)
		return
	}

	// make sure the key is valid
	key, err := util.Z32Decode(*id)
	if err != nil {
		LoggingRespondErrWithMsg(c, err, fmt.Sprintf("invalid record id: %s", *id), http.StatusInternalServerError)
		return
	}
	if len(key) != ed25519.PublicKeySize {
		LoggingRespondErrMsg(c, fmt.Sprintf("invalid z32 encoded ed25519 public key: %s", *id), http.StatusBadRequest)
		return
	}

	resp, err := r.service.GetDHT(ctx, *id)
	if err != nil {
		if errors.Is(err, service.SpamError) {
			LoggingRespondErrMsg(c, fmt.Sprintf("too many requests for bad key %s", *id), http.StatusTooManyRequests)
			return
		}

		LoggingRespondErrWithMsg(c, err, fmt.Sprintf("failed to get dht record: %s", *id), http.StatusInternalServerError)
		return
	}
	if resp == nil {
		LoggingRespondErrMsg(c, fmt.Sprintf("dht record not found: %s", *id), http.StatusNotFound)
		return
	}

	// Convert int64 to uint64 since binary.PutUint64 expects a uint64 value
	var seqBuf [8]byte
	binary.BigEndian.PutUint64(seqBuf[:], uint64(resp.Seq))
	// sig:seq:v
	res := append(resp.Sig[:], append(seqBuf[:], resp.V[:]...)...)
	RespondBytes(c, res, http.StatusOK)
}

// PutRecord godoc
//
//	@Summary		PutRecord a BEP44 DNS record into the DHT
//	@Description	PutRecord a BEP44 DNS record into the DHT
//	@Tags			DHT
//	@Accept			octet-stream
//	@Param			id		path	string	true	"ID of the record to put"
//	@Param			request	body	[]byte	true	"64 bytes sig, 8 bytes u64 big-endian seq, 0-1000 bytes of v."
//	@Success		200
//	@Failure		400	{string}	string	"Bad request"
//	@Failure		500	{string}	string	"Internal server error"
//	@Router			/{id} [put]
func (r *DHTRouter) PutRecord(c *gin.Context) {
	ctx, span := telemetry.GetTracer().Start(c, "DHTHTTP.PutRecord")
	defer span.End()

	id := GetParam(c, IDParam)
	if id == nil || *id == "" {
		LoggingRespondErrMsg(c, "missing id param", http.StatusBadRequest)
		return
	}
	key, err := util.Z32Decode(*id)
	if err != nil {
		LoggingRespondErrWithMsg(c, err, fmt.Sprintf("invalid record id: %s", *id), http.StatusInternalServerError)
		return
	}
	if len(key) != ed25519.PublicKeySize {
		LoggingRespondErrMsg(c, fmt.Sprintf("invalid z32 encoded ed25519 public key: %s", *id), http.StatusBadRequest)
		return
	}

	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		LoggingRespondErrWithMsg(c, err, fmt.Sprintf("failed to read body for id: %s", *id), http.StatusInternalServerError)
		return
	}
	defer c.Request.Body.Close()

	// 64 byte signature and 8 byte sequence number
	if len(body) <= 72 {
		LoggingRespondErrMsg(c, fmt.Sprintf("invalid request body for id: %s", *id), http.StatusBadRequest)
		return
	}

	// transform the request into a service request by extracting the fields
	value := body[72:]
	sig := body[:64]
	seq := int64(binary.BigEndian.Uint64(body[64:72]))
	request, err := dht.NewBEP44Record(key, value, sig, seq)
	if err != nil {
		LoggingRespondErrWithMsg(c, err, "error parsing request", http.StatusBadRequest)
		return
	}

	if err = r.service.PublishDHT(ctx, *id, *request); err != nil {
		LoggingRespondErrWithMsg(c, err, fmt.Sprintf("failed to publish dht record: %s", *id), http.StatusInternalServerError)
		return
	}

	ResponseStatus(c, http.StatusOK)
}
