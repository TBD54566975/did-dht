package service

import (
	"context"

	"github.com/TBD54566975/ssi-sdk/util"
	"github.com/anacrolix/dht/v2/bep44"

	"github.com/TBD54566975/did-dht-method/config"
	"github.com/TBD54566975/did-dht-method/pkg/dht"
	"github.com/TBD54566975/did-dht-method/pkg/storage"
)

// PKARRService is the PKARR service responsible for managing the PKARR DHT and reading/writing records
type PKARRService struct {
	cfg *config.Config
	db  *storage.Storage
	dht *dht.DHT
}

// NewPKARRService returns a new instance of the PKARR service
func NewPKARRService(cfg *config.Config, db *storage.Storage) (*PKARRService, error) {
	if cfg == nil {
		return nil, util.LoggingNewError("config is required")
	}
	if db == nil && !db.IsOpen() {
		return nil, util.LoggingNewError("storage is required be non-nil and to be open")
	}
	dht, err := dht.NewDHT(cfg.DHTConfig.BootstrapPeers)
	if err != nil {
		return nil, util.LoggingErrorMsg(err, "failed to instantiate dht")
	}
	return &PKARRService{
		cfg: cfg,
		db:  db,
		dht: dht,
	}, nil
}

// PutPKARRRequest is the request to publish a PKARR record
type PutPKARRRequest struct {
	V   []byte   `json:"v" validate:"required"`
	K   [32]byte `json:"k" validate:"required"`
	Sig [64]byte `json:"sig" validate:"required"`
	Seq int64    `json:"seq" validate:"required"`
}

// PublishPKARR publishes the given PKARR to the DHT
func (s *PKARRService) PublishPKARR(ctx context.Context, request PutPKARRRequest) (string, error) {
	return s.dht.Put(ctx, bep44.Put{
		V:   request.V,
		K:   &request.K,
		Sig: request.Sig,
		Seq: request.Seq,
	})
}

// GetPKARRResponse is the response to a get PKARR request
type GetPKARRResponse struct {
	V   []byte   `json:"v" validate:"required"`
	Seq int64    `json:"seq" validate:"required"`
	Sig [64]byte `json:"sig,omitempty"`
}

// GetPKARR returns the PKARR for the given z-base-32 encoded ID
func (s *PKARRService) GetPKARR(ctx context.Context, id string) (*GetPKARRResponse, error) {
	got, err := s.dht.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	bBytes, err := got.V.MarshalBencode()
	if err != nil {
		return nil, err
	}
	return &GetPKARRResponse{
		V:   bBytes,
		Seq: got.Seq,
	}, nil
}

// GetFullPKARR returns the full PKARR (including sig data) for the given z-base-32 encoded ID
func (s *PKARRService) GetFullPKARR(ctx context.Context, id string) (*GetPKARRResponse, error) {
	got, err := s.dht.GetFull(ctx, id)
	if err != nil {
		return nil, err
	}
	bBytes, err := got.V.MarshalBencode()
	if err != nil {
		return nil, err
	}
	return &GetPKARRResponse{
		V:   bBytes,
		Seq: got.Seq,
		Sig: got.Sig,
	}, nil
}
