package server

import (
	"crypto/ed25519"
	"encoding/binary"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/TBD54566975/did-dht-method/internal/util"
	"github.com/TBD54566975/did-dht-method/pkg/pkarr"
	"github.com/TBD54566975/did-dht-method/pkg/service"
)

// PkarrRouter is the router for the Pkarr API
type PkarrRouter struct {
	service *service.PkarrService
}

// NewPkarrRouter returns a new instance of the Relay router
func NewPkarrRouter(service *service.PkarrService) (*PkarrRouter, error) {
	return &PkarrRouter{service: service}, nil
}

// GetRecord godoc
//
//	@Summary		GetRecord a Pkarr record from the DHT
//	@Description	GetRecord a Pkarr record from the DHT
//	@Tags			Pkarr
//	@Accept			octet-stream
//	@Produce		octet-stream
//	@Param			id	path		string	true	"ID to get"
//	@Success		200	{array}		byte	"64 bytes sig, 8 bytes u64 big-endian seq, 0-1000 bytes of v."
//	@Failure		400	{string}	string	"Bad request"
//	@Failure		404	{string}	string	"Not found"
//	@Failure		500	{string}	string	"Internal server error"
//	@Router			/{id} [get]
func (r *PkarrRouter) GetRecord(c *gin.Context) {
	id := GetParam(c, IDParam)
	if id == nil || *id == "" {
		LoggingRespondErrMsg(c, "missing id param", http.StatusBadRequest)
		return
	}

	resp, err := r.service.GetPkarr(c, *id)
	if err != nil {
		LoggingRespondErrWithMsg(c, err, "failed to get pkarr record", http.StatusInternalServerError)
		return
	}
	if resp == nil {
		LoggingRespondErrMsg(c, "pkarr record not found", http.StatusNotFound)
		return
	}

	// Convert int64 to uint64 since binary.PutUint64 expects a uint64 value
	// according to https://github.com/Nuhvi/pkarr/blob/main/design/relays.md#get
	var seqBuf [8]byte
	binary.BigEndian.PutUint64(seqBuf[:], uint64(resp.Seq))
	// sig:seq:v
	res := append(resp.Sig[:], append(seqBuf[:], resp.V[:]...)...)
	RespondBytes(c, res, http.StatusOK)
}

// PutRecord godoc
//
//	@Summary		PutRecord a Pkarr record into the DHT
//	@Description	PutRecord a Pkarr record into the DHT
//	@Tags			Pkarr
//	@Accept			octet-stream
//	@Param			id		path	string	true	"ID of the record to put"
//	@Param			request	body	[]byte	true	"64 bytes sig, 8 bytes u64 big-endian seq, 0-1000 bytes of v."
//	@Success		200
//	@Failure		400	{string}	string	"Bad request"
//	@Failure		500	{string}	string	"Internal server error"
//	@Router			/{id} [put]
func (r *PkarrRouter) PutRecord(c *gin.Context) {
	id := GetParam(c, IDParam)
	if id == nil || *id == "" {
		LoggingRespondErrMsg(c, "missing id param", http.StatusBadRequest)
		return
	}
	key, err := util.Z32Decode(*id)
	if err != nil {
		LoggingRespondErrWithMsg(c, err, "failed to read id", http.StatusInternalServerError)
		return
	}
	if len(key) != ed25519.PublicKeySize {
		LoggingRespondErrMsg(c, "invalid z32 encoded ed25519 public key", http.StatusBadRequest)
		return
	}

	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		LoggingRespondErrWithMsg(c, err, "failed to read body", http.StatusInternalServerError)
		return
	}
	defer c.Request.Body.Close()

	// 64 byte signature and 8 byte sequence number
	if len(body) <= 72 {
		LoggingRespondErrMsg(c, "invalid request body", http.StatusBadRequest)
		return
	}

	// transform the request into a service request by extracting the fields
	// according to https://github.com/Nuhvi/pkarr/blob/main/design/relays.md#put
	value := body[72:]
	sig := body[:64]
	seq := int64(binary.BigEndian.Uint64(body[64:72]))
	request, err := pkarr.NewRecord(key, value, sig, seq)
	if err != nil {
		LoggingRespondErrWithMsg(c, err, "error parsing request", http.StatusBadRequest)
		return
	}

	if err = r.service.PublishPkarr(c, *id, *request); err != nil {
		LoggingRespondErrWithMsg(c, err, "failed to publish pkarr record", http.StatusInternalServerError)
		return
	}

	ResponseStatus(c, http.StatusOK)
}
