package service

import (
	"context"

	"github.com/TBD54566975/ssi-sdk/did"
	"github.com/TBD54566975/ssi-sdk/util"

	"github.com/TBD54566975/did-dht-method/config"
	"github.com/TBD54566975/did-dht-method/pkg/dht"
	"github.com/TBD54566975/did-dht-method/pkg/storage"
)

// DIDService is the DID DHT service responsible for managing the DHT and reading/writing records
type DIDService struct {
	cfg *config.Config
	db  *storage.Storage

	dht *dht.DHT
}

// NewDIDService returns a new instance of the DHT service
func NewDIDService(cfg *config.Config, db *storage.Storage) (*DIDService, error) {
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
	return &DIDService{
		cfg: cfg,
		db:  db,
		dht: dht,
	}, nil
}

type PublishDIDRequest struct {
}

func (s *DIDService) PublishDID(ctx context.Context, request PublishDIDRequest) error {
	return nil
}

func (s *DIDService) GetDID(ctx context.Context, did string) (*did.Document, error) {
	return nil, nil
}

func (s *DIDService) ListDIDs(ctx context.Context) ([]did.Document, error) {
	return nil, nil
}

func (s *DIDService) DeleteDID(ctx context.Context, did string) error {
	return nil
}
