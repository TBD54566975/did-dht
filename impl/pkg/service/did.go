package service

import (
	"context"

	"github.com/TBD54566975/ssi-sdk/did"
	"github.com/TBD54566975/ssi-sdk/util"
	"github.com/pkg/errors"

	"github.com/TBD54566975/did-dht-method/config"
	"github.com/TBD54566975/did-dht-method/pkg/storage"
)

// DIDService is the DID DHT service responsible for managing the DHT and reading/writing records
type DIDService struct {
	cfg   *config.Config
	db    *storage.Storage
	pkarr PKARRService
}

// NewDIDService returns a new instance of the DHT service
func NewDIDService(cfg *config.Config, db *storage.Storage, pkarr PKARRService) (*DIDService, error) {
	if cfg == nil {
		return nil, util.LoggingNewError("config is required")
	}
	if db == nil && !db.IsOpen() {
		return nil, util.LoggingNewError("storage is required be non-nil and to be open")
	}
	return &DIDService{
		cfg:   cfg,
		db:    db,
		pkarr: pkarr,
	}, nil
}

type PublishDIDRequest struct {
}

func (s *DIDService) PublishDID(_ context.Context, _ PublishDIDRequest) error {
	return errors.New("unimplemented")
}

func (s *DIDService) GetDID(_ context.Context, did string) (*did.Document, error) {
	return nil, errors.New("unimplemented")
}

func (s *DIDService) ListDIDs(_ context.Context) ([]did.Document, error) {
	return nil, errors.New("unimplemented")
}

func (s *DIDService) DeleteDID(_ context.Context, _ string) error {
	return errors.New("unimplemented")
}
