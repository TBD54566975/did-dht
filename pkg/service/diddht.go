package service

import (
	"github.com/sirupsen/logrus"

	"did-dht/config"
	"did-dht/pkg/db"
)

type DIDDHTService struct {
	storage *db.Storage
}

func NewDIDDHTService(cfg config.Config) (*DIDDHTService, error) {
	storage, err := db.NewStorage(cfg.DBFile)
	if err != nil {
		logrus.WithError(err).Error("failed to instantiate storage")
		return nil, err
	}
	return &DIDDHTService{storage: storage}, nil
}
