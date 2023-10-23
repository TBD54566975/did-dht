package service

import (
	"context"
	"encoding/base64"

	"github.com/TBD54566975/ssi-sdk/util"
	"github.com/anacrolix/dht/v2/bep44"
	"github.com/anacrolix/torrent/bencode"
	"github.com/sirupsen/logrus"

	"github.com/TBD54566975/did-dht-method/config"
	dhtint "github.com/TBD54566975/did-dht-method/internal/dht"
	"github.com/TBD54566975/did-dht-method/pkg/dht"
	"github.com/TBD54566975/did-dht-method/pkg/storage"
)

// PKARRService is the PKARR service responsible for managing the PKARR DHT and reading/writing records
type PKARRService struct {
	cfg       *config.Config
	db        *storage.Storage
	dht       *dht.DHT
	scheduler *dhtint.Scheduler
}

// NewPKARRService returns a new instance of the PKARR service
func NewPKARRService(cfg *config.Config, db *storage.Storage) (*PKARRService, error) {
	if cfg == nil {
		return nil, util.LoggingNewError("config is required")
	}
	if db == nil && !db.IsOpen() {
		return nil, util.LoggingNewError("storage is required be non-nil and to be open")
	}
	d, err := dht.NewDHT(cfg.DHTConfig.BootstrapPeers)
	if err != nil {
		return nil, util.LoggingErrorMsg(err, "failed to instantiate dht")
	}
	scheduler := dhtint.NewScheduler()
	service := PKARRService{
		cfg:       cfg,
		db:        db,
		dht:       d,
		scheduler: &scheduler,
	}
	if err = scheduler.Schedule(cfg.PKARRConfig.RepublishCRON, service.republish); err != nil {
		return nil, util.LoggingErrorMsg(err, "failed to start republisher")
	}
	return &service, nil
}

// PublishPKARRRequest is the request to publish a PKARR record
type PublishPKARRRequest struct {
	V   []byte   `validate:"required"`
	K   [32]byte `validate:"required"`
	Sig [64]byte `validate:"required"`
	Seq int64    `validate:"required"`
}

func (p PublishPKARRRequest) toRecord() storage.PKARRRecord {
	encoding := base64.RawURLEncoding
	return storage.PKARRRecord{
		V:   encoding.EncodeToString(p.V),
		K:   encoding.EncodeToString(p.K[:]),
		Sig: encoding.EncodeToString(p.Sig[:]),
		Seq: p.Seq,
	}
}

// PublishPKARR stores the record in the db, publishes the given PKARR to the DHT, and returns the z-base-32 encoded ID
func (s *PKARRService) PublishPKARR(ctx context.Context, request PublishPKARRRequest) (string, error) {
	if err := util.IsValidStruct(request); err != nil {
		return "", err
	}
	// TODO(gabe): if putting to the DHT fails we should note that in the db and retry later
	if err := s.db.WriteRecord(request.toRecord()); err != nil {
		return "", err
	}
	return s.dht.Put(ctx, bep44.Put{
		V:   request.V,
		K:   &request.K,
		Sig: request.Sig,
		Seq: request.Seq,
	})
}

// GetPKARRResponse is the response to a get PKARR request
type GetPKARRResponse struct {
	V   []byte   `validate:"required"`
	Seq int64    `validate:"required"`
	Sig [64]byte `validate:"required"`
}

func fromPKARRRecord(record storage.PKARRRecord) (*GetPKARRResponse, error) {
	encoding := base64.RawURLEncoding
	vBytes, err := encoding.DecodeString(record.V)
	if err != nil {
		return nil, err
	}
	sigBytes, err := encoding.DecodeString(record.Sig)
	if err != nil {
		return nil, err
	}
	return &GetPKARRResponse{
		V:   vBytes,
		Seq: record.Seq,
		Sig: [64]byte(sigBytes),
	}, nil
}

// GetPKARR returns the full PKARR (including sig data) for the given z-base-32 encoded ID
func (s *PKARRService) GetPKARR(ctx context.Context, id string) (*GetPKARRResponse, error) {
	got, err := s.dht.GetFull(ctx, id)
	if err != nil {
		// try to resolve from storage before returning and error
		// if we detect this and have the record we should republish to the DHT
		logrus.WithError(err).Warnf("failed to get pkarr<%s> from dht, attempting to resolve from storage", id)
		record, err := s.db.ReadRecord(id)
		if err != nil || record == nil {
			logrus.WithError(err).Errorf("failed to resolve pkarr<%s> from storage", id)
			return nil, err
		}
		logrus.Debugf("resolved pkarr<%s> from storage", id)
		return fromPKARRRecord(*record)
	}
	bBytes, err := got.V.MarshalBencode()
	if err != nil {
		return nil, err
	}
	var payload string
	if err = bencode.Unmarshal(bBytes, &payload); err != nil {
		return nil, err
	}
	return &GetPKARRResponse{
		V:   []byte(payload),
		Seq: got.Seq,
		Sig: got.Sig,
	}, nil
}

// TODO(gabe) make this more efficient. create a publish schedule based on each individual record, not all records
func (s *PKARRService) republish() {
	allRecords, err := s.db.ListRecords()
	if err != nil {
		logrus.WithError(err).Error("failed to list record(s) for republishing")
		return
	}
	if len(allRecords) == 0 {
		logrus.Info("No records to republish")
		return
	}
	logrus.Infof("Republishing %d record(s)", len(allRecords))
	errCnt := 0
	for _, record := range allRecords {
		put, err := recordToBEP44Put(record)
		if err != nil {
			logrus.WithError(err).Error("failed to convert record to bep44 put")
			errCnt++
			continue
		}
		if _, err = s.dht.Put(context.Background(), *put); err != nil {
			logrus.WithError(err).Error("failed to republish record")
			errCnt++
			continue
		}
	}
	logrus.Infof("Republishing complete. Successfully republished %d out of %d record(s)", len(allRecords)-errCnt, len(allRecords))
}

func recordToBEP44Put(record storage.PKARRRecord) (*bep44.Put, error) {
	encoding := base64.RawURLEncoding
	vBytes, err := encoding.DecodeString(record.V)
	if err != nil {
		return nil, err
	}
	kBytes, err := encoding.DecodeString(record.K)
	if err != nil {
		return nil, err
	}
	sigBytes, err := encoding.DecodeString(record.Sig)
	if err != nil {
		return nil, err
	}
	return &bep44.Put{
		V:   vBytes,
		K:   (*[32]byte)(kBytes),
		Sig: [64]byte(sigBytes),
		Seq: record.Seq,
	}, nil
}
