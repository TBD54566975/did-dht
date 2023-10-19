package server

import (
	"crypto/ed25519"
	"encoding/binary"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/TBD54566975/did-dht-method/internal/util"
	"github.com/TBD54566975/did-dht-method/pkg/service"
)

// RelayRouter is the router for the Relay API
type RelayRouter struct {
	service *service.PKARRService
}

// NewRelayRouter returns a new instance of the Relay router
func NewRelayRouter(service *service.PKARRService) (*RelayRouter, error) {
	return &RelayRouter{service: service}, nil
}

// Get godoc
//
//	@Summary		Get a PKARR from the DHT
//	@Description	Get a PKARR from the DHT
//	@Tags			Relay
//	@Accept			octet-stream
//	@Produce		octet-stream
//	@Param			id	path		string	true	"ID to get"
//	@Success		200	{array}		byte	"64 bytes sig, 8 bytes u64 big-endian seq, 0-1000 bytes of v."
//	@Failure		400	{string}	string	"Bad request"
//	@Failure		404	{string}	string	"Not found"
//	@Failure		500	{string}	string	"Internal server error"
//	@Router			/{id} [get]
func (r *RelayRouter) Get(c *gin.Context) {
	id := GetParam(c, IDParam)
	if id == nil || *id == "" {
		LoggingRespondErrMsg(c, "missing id param", http.StatusBadRequest)
		return
	}

	resp, err := r.service.GetFullPKARR(c, *id)
	if err != nil {
		LoggingRespondErrWithMsg(c, err, "failed to get pkarr", http.StatusInternalServerError)
		return
	}

	if resp == nil {
		LoggingRespondErrMsg(c, "pkarr not found", http.StatusNotFound)
		return
	}

	// Convert int64 to uint64 since binary.PutUint64 expects a uint64 value
	// according to https://github.com/Nuhvi/pkarr/blob/main/design/relays.md#get
	var seqBuf [8]byte
	binary.BigEndian.PutUint64(seqBuf[:], uint64(resp.Seq))
	partialRes := append(seqBuf[:], resp.V...)
	res := append(resp.Sig[:], partialRes...)
	RespondBytes(c, res, http.StatusOK)
}

// Put godoc
//
//	@Summary		Put a PKARR record into the DHT
//	@Description	Put a PKARR record into the DHT
//	@Tags			Relay
//	@Accept			octet-stream
//	@Param			id		path	string	true	"ID to put"
//	@Param			request	body	[]byte	true	"64 bytes sig, 8 bytes u64 big-endian seq, 0-1000 bytes of v."
//	@Success		200
//	@Failure		400	{string}	string	"Bad request"
//	@Failure		500	{string}	string	"Internal server error"
//	@Router			/{id} [put]
func (r *RelayRouter) Put(c *gin.Context) {
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
	vBytes := body[72:]
	keyBytes := [32]byte(key[:])
	bytes := body[:64]
	sigBytes := [64]byte(bytes)
	seq := int64(binary.BigEndian.Uint64(body[64:72]))
	request := service.PutPKARRRequest{
		V:   vBytes,
		K:   keyBytes,
		Sig: sigBytes,
		Seq: seq,
	}
	if _, err = r.service.PublishPKARR(c, request); err != nil {
		LoggingRespondErrWithMsg(c, err, "failed to publish pkarr request", http.StatusInternalServerError)
		return
	}

	Respond(c, nil, http.StatusOK)
}
