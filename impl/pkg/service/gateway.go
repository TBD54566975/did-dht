package service

import (
	"github.com/TBD54566975/ssi-sdk/util"

	"github.com/TBD54566975/did-dht-method/config"
	"github.com/TBD54566975/did-dht-method/pkg/storage"
)

type GatewayService struct {
	cfg   *config.Config
	db    *storage.Storage
	pkarr *PkarrService
}

func NewGatewayService(cfg *config.Config, db *storage.Storage, pkarrService *PkarrService) (*GatewayService, error) {
	if cfg == nil {
		return nil, util.LoggingNewError("config is required")
	}
	if db == nil && !db.IsOpen() {
		return nil, util.LoggingNewError("storage is required be non-nil and to be open")
	}
	if pkarrService == nil {
		return nil, util.LoggingNewError("pkarr service is required")
	}
	return &GatewayService{
		cfg:   cfg,
		db:    db,
		pkarr: pkarrService,
	}, nil
}

type PublishDIDRequest struct {
	DID            string `json:"did" validate:"required"`
	Sig            string `json:"sig" validate:"required"`
	Seq            int    `json:"seq" validate:"required"`
	V              string `json:"v" validate:"required"`
	RetentionProof int    `json:"retention_proof,omitempty"`
}

func (s *GatewayService) PublishDID(req PublishDIDRequest) error {
	return nil
}

type GetDIDResponse struct {
}

func (s *GatewayService) GetDID(id string) (*GetDIDResponse, error) {
	return nil, nil
}

type GetTypesResponse struct {
}

func (s *GatewayService) GetTypes() (*GetTypesResponse, error) {
	return nil, nil
}

type GetDIDsForTypeRequest struct {
}

type GetDIDsForTypeResponse struct {
}

func (s *GatewayService) GetDIDsForType(req *GetDIDsForTypeRequest) (*GetDIDsForTypeResponse, error) {
	return nil, nil
}

type GetDIDsForTypesResponse struct {
}

func (s *GatewayService) GetDifficulty() {

}
