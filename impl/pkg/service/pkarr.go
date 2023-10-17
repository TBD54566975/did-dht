package service

import (
	"context"

	"github.com/TBD54566975/ssi-sdk/util"

	"github.com/TBD54566975/did-dht-method/config"
	"github.com/TBD54566975/did-dht-method/pkg/dht"
	"github.com/TBD54566975/did-dht-method/pkg/storage"
)

type PKARRService struct {
	cfg *config.Config
	db  *storage.Storage

	dht *dht.DHT
}

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

type PutPKARRRequest struct {
	V   []byte   `json:"v" validate:"required"`
	Sig [64]byte `json:"sig" validate:"required"`
	Seq int64    `json:"seq" validate:"required"`
}

func (s *PKARRService) PublishPKARR(ctx context.Context, request PutPKARRRequest) (string, error) {
	return "", nil
}

type GetPKARRResponse struct {
	V   []byte   `json:"v" validate:"required"`
	Sig [64]byte `json:"sig" validate:"required"`
	Seq int64    `json:"seq" validate:"required"`
}

func (s *PKARRService) GetPKARR(ctx context.Context, id string) (*GetPKARRResponse, error) {
	return nil, nil
}
