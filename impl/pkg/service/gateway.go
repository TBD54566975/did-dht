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

	// check to see if the DID already exists with a higher sequence number
	gotDID, err := s.db.ReadDID(req.DID)
	if err == nil && gotDID != nil {
		if gotDID.SequenceNumber > req.Seq {
			return &intutil.HigherSequenceNumberError{}
		}
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
	DID             did.Document       `json:"did" validate:"required"`
	Types           []didint.TypeIndex `json:"types,omitempty"`
	SequenceNumbers []int              `json:"sequence_numbers,omitempty"`
}

func (s *GatewayService) GetDID(id string) (*GetDIDResponse, error) {
	gotDID, err := s.db.ReadDID(id)
	if err != nil {
		return nil, err
	}

	if gotDID == nil {
		return nil, nil
	}

	return &GetDIDResponse{
		DID:             gotDID.Document,
		Types:           gotDID.Types,
		SequenceNumbers: []int{int(gotDID.SequenceNumber)},
	}, nil
}

type GetTypesResponse struct {
	Types []TypeMapping `json:"types,omitempty"`
}

type TypeMapping struct {
	TypeIndex didint.TypeIndex `json:"type_index" validate:"required"`
	Type      string           `json:"type" validate:"required"`
}

// GetTypes returns a list of supported types and their names.
// As defined by the spec's registry https://did-dht.com/registry/index.html#indexed-types
func (s *GatewayService) GetTypes() GetTypesResponse {
	return GetTypesResponse{
		Types: knownTypes,
	}
}

type ListDIDsForTypeRequest struct {
	Type didint.TypeIndex `json:"type" validate:"required"`
}

type ListDIDsForTypeResponse struct {
	DIDs []string `json:"dids,omitempty"`
}

// ListDIDsForType returns a list of DIDs for a given type.
func (s *GatewayService) ListDIDsForType(req ListDIDsForTypeRequest) (*ListDIDsForTypeResponse, error) {
	if !isKnownType(req.Type) {
		return nil, &intutil.TypeNotFoundError{}
	}
	dids, err := s.db.ListDIDsForType(req.Type)
	if err != nil {
		return nil, err
	}
	if len(dids) == 0 {
		return nil, nil
	}
	return &ListDIDsForTypeResponse{DIDs: dids}, nil
}

type GetDifficultyResponse struct {
	Hash       string `json:"hash" validate:"required"`
	Difficulty int    `json:"difficulty" validate:"required"`
}

// GetDifficulty returns the current difficulty for the gateway's retention proof feature.
// TODO(gabe): retention proof support https://github.com/TBD54566975/did-dht-method/issues/73
func (s *GatewayService) GetDifficulty() (*GetDifficultyResponse, error) {
	return nil, errors.New("not yet implemented")
}

func isKnownType(t didint.TypeIndex) bool {
	for _, knownType := range knownTypes {
		if knownType.TypeIndex == t {
			return true
		}
	}
	return false
}

var (
	knownTypes = []TypeMapping{
		{
			TypeIndex: didint.Discoverable,
			Type:      "Discoverable",
		},
		{
			TypeIndex: didint.Organization,
			Type:      "Organization",
		},
		{
			TypeIndex: didint.GovernmentOrganization,
			Type:      "Government Organization",
		},
		{
			TypeIndex: didint.Corporation,
			Type:      "Corporation",
		},
		{
			TypeIndex: didint.LocalBusiness,
			Type:      "Local Business",
		},
		{
			TypeIndex: didint.SoftwarePackage,
			Type:      "Software Package",
		},
		{
			TypeIndex: didint.WebApplication,
			Type:      "Web Application",
		},
		{
			TypeIndex: didint.FinancialInstitution,
			Type:      "Financial Institution",
		},
	}
)
