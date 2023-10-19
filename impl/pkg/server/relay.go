package server

import (
	"encoding/binary"
	"net/http"

	"github.com/gin-gonic/gin"

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
//	@Param			id		path		string	true	"ID to get"
//	@Success		200		{array}	byte	"64 bytes sig, 8 bytes u64 big-endian seq, 0-1000 bytes of v."
//	@Failure		400		{string}	string	"Bad request"
//	@Failure		404		{string}	string	"Not found"
//	@Failure		500		{string}	string	"Internal server error"
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

func (r *RelayRouter) Put(c *gin.Context) {
	id := GetParam(c, IDParam)
	if id == nil || *id == "" {
		LoggingRespondErrMsg(c, "missing id param", http.StatusBadRequest)
		return
	}
}
