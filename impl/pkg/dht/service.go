package dht

import (
	"context"

	"github.com/TBD54566975/ssi-sdk/did"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/TBD54566975/did-dht-method/config"
	"github.com/TBD54566975/did-dht-method/pkg/storage"
)

// Service is the DID DHT service responsible for managing the DHT and reading/writing records
type Service struct {
	cfg *config.Config
	db  *storage.Storage

	dht *DHT
}

// NewService returns a new instance of the DHT service
func NewService(cfg *config.Config, db *storage.Storage) (*Service, error) {
	if cfg == nil {
		return nil, errors.New("config manager is required")
	}
	if db == nil && !db.IsOpen() {
		return nil, errors.New("storage is required be non-nil and to be open")
	}
	dht, err := NewDHT(cfg.DHTConfig.BootstrapPeers)
	if err != nil {
		logrus.WithError(err).Error("failed to instantiate dht")
		return nil, errors.Wrap(err, "failed to instantiate dht")
	}
	return &Service{
		cfg: cfg,
		db:  db,
		dht: dht,
	}, nil
}

type PublishDIDRequest struct {
}

func (s *Service) PublishDID(ctx context.Context, request PublishDIDRequest) error {
	return nil
}

func (s *Service) GetDID(ctx context.Context, did string) (*did.Document, error) {
	return nil, nil
}

func (s *Service) ListDIDs(ctx context.Context) ([]did.Document, error) {
	return nil, nil
}

func (s *Service) DeleteDID(ctx context.Context, did string) error {
	return nil
}
