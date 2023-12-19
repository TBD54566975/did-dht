package service

import (
	"context"
	"crypto/ed25519"
	"encoding/base64"

	"github.com/TBD54566975/ssi-sdk/did"
	"github.com/TBD54566975/ssi-sdk/util"
	"github.com/miekg/dns"
	"github.com/pkg/errors"
	"github.com/tv42/zbase32"

	"github.com/TBD54566975/did-dht-method/config"
	didint "github.com/TBD54566975/did-dht-method/internal/did"
	intutil "github.com/TBD54566975/did-dht-method/internal/util"
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
	Seq            int64  `json:"seq" validate:"required"`
	V              string `json:"v" validate:"required"`
	RetentionProof string `json:"retention_proof,omitempty"`
}

func (p PublishDIDRequest) toPkarrRequest(suffix string) (*PublishPkarrRequest, error) {
	keyBytes, err := zbase32.DecodeString(suffix)
	if err != nil {
		return nil, err
	}
	if len(keyBytes) != ed25519.PublicKeySize {
		return nil, errors.New("invalid key length")
	}
	encoding := base64.RawURLEncoding
	sigBytes, err := encoding.DecodeString(p.Sig)
	if err != nil {
		return nil, err
	}
	if len(sigBytes) != ed25519.SignatureSize {
		return nil, &intutil.InvalidSignatureError{}
	}
	vBytes, err := encoding.DecodeString(p.V)
	if err != nil {
		return nil, err
	}
	if len(vBytes) > 1000 {
		return nil, errors.New("v exceeds 1000 bytes")
	}
	return &PublishPkarrRequest{
		V:   vBytes,
		K:   [32]byte(keyBytes),
		Sig: [64]byte(sigBytes),
		Seq: p.Seq,
	}, nil
}

func (s *GatewayService) PublishDID(ctx context.Context, req PublishDIDRequest) error {
	id := didint.DHT(req.DID)
	suffix, err := id.Suffix()
	if err != nil {
		return err
	}
	pkarrRequest, err := req.toPkarrRequest(suffix)
	if err != nil {
		return err
	}

	// TODO(gabe): retention proof support https://github.com/TBD54566975/did-dht-method/issues/73

	// unpack as a DID Document and store metadata
	msg := new(dns.Msg)
	if err = msg.Unpack(pkarrRequest.V); err != nil {
		return errors.Wrap(err, "failed to unpack records")
	}
	doc, types, err := id.FromDNSPacket(msg)
	if err != nil {
		return errors.Wrap(err, "failed to parse DID document from DNS packet")
	}
	if err = s.db.WriteDID(storage.GatewayRecord{
		Document:       *doc,
		Types:          types,
		SequenceNumber: req.Seq,
		RetentionProof: req.RetentionProof,
	}); err != nil {
		return errors.Wrap(err, "failed to write DID document to db")
	}

	// publish to the network
	// TODO(gabe): check for conflicts with existing record sequence numbers https://github.com/TBD54566975/did-dht-method/issues/16
	if err = s.pkarr.PublishPkarr(ctx, suffix, *pkarrRequest); err != nil {
		return err
	}

	return nil
}

type GetDIDResponse struct {
	DID             did.Document `json:"did" validate:"required"`
	Types           []int        `json:"types,omitempty"`
	SequenceNumbers []int        `json:"sequence_numbers,omitempty"`
}

func (s *GatewayService) GetDID(id string) (*GetDIDResponse, error) {
	return nil, nil
}

type GetTypesResponse struct {
	Types []TypeMapping `json:"types,omitempty"`
}

type TypeMapping struct {
	TypeIndex int    `json:"type_index" validate:"required"`
	Type      string `json:"type" validate:"required"`
}

func (s *GatewayService) GetTypes() (*GetTypesResponse, error) {
	return nil, nil
}

type GetDIDsForTypeRequest struct {
	Type string `json:"type" validate:"required"`
}

type GetDIDsForTypeResponse struct {
	DIDs []string `json:"dids,omitempty"`
}

func (s *GatewayService) GetDIDsForType(req GetDIDsForTypeRequest) (*GetDIDsForTypeResponse, error) {
	return nil, nil
}

type GetDifficultyResponse struct {
	Hash       string `json:"hash" validate:"required"`
	Difficulty int    `json:"difficulty" validate:"required"`
}

func (s *GatewayService) GetDifficulty() (*GetDifficultyResponse, error) {
	return nil, nil
}
