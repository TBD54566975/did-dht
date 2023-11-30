package service

import (
	"context"
	"encoding/base64"
	"errors"
	"time"

	"github.com/TBD54566975/ssi-sdk/util"
	"github.com/anacrolix/dht/v2/bep44"
	"github.com/anacrolix/torrent/bencode"
	"github.com/jellydator/ttlcache/v3"
	"github.com/sirupsen/logrus"

	"github.com/TBD54566975/did-dht-method/config"
	dhtint "github.com/TBD54566975/did-dht-method/internal/dht"
	"github.com/TBD54566975/did-dht-method/pkg/dht"
	"github.com/TBD54566975/did-dht-method/pkg/storage"
)

// PkarrService is the Pkarr service responsible for managing the Pkarr DHT and reading/writing records
type PkarrService struct {
	cfg       *config.Config
	db        *storage.Storage
	dht       *dht.DHT
	cache     *ttlcache.Cache[string, storage.PkarrRecord]
	scheduler *dhtint.Scheduler
}

// NewPkarrService returns a new instance of the Pkarr service
func NewPkarrService(cfg *config.Config, db *storage.Storage) (*PkarrService, error) {
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

	// create and start cache and scheduler
	ttl := time.Duration(cfg.PkarrConfig.CacheTTLMinutes) * time.Minute
	cache := ttlcache.New[string, storage.PkarrRecord](
		ttlcache.WithTTL[string, storage.PkarrRecord](ttl),
	)
	scheduler := dhtint.NewScheduler()
	service := PkarrService{
		cfg:       cfg,
		db:        db,
		dht:       d,
		cache:     cache,
		scheduler: &scheduler,
	}
	go cache.Start()
	if err = scheduler.Schedule(cfg.PkarrConfig.RepublishCRON, service.republish); err != nil {
		return nil, util.LoggingErrorMsg(err, "failed to start republisher")
	}
	return &service, nil
}

// PublishPkarrRequest is the request to publish a Pkarr record
type PublishPkarrRequest struct {
	V   []byte   `validate:"required"`
	K   [32]byte `validate:"required"`
	Sig [64]byte `validate:"required"`
	Seq int64    `validate:"required"`
}

// isValid returns an error if the request is invalid; also validates the signature
func (p PublishPkarrRequest) isValid() error {
	if err := util.IsValidStruct(p); err != nil {
		return err
	}
	// validate the signature
	bv, err := bencode.Marshal(p.V)
	if err != nil {
		return err
	}
	if !bep44.Verify(p.K[:], nil, p.Seq, bv, p.Sig[:]) {
		return errors.New("signature is invalid")
	}
	return nil
}

func (p PublishPkarrRequest) toRecord() storage.PkarrRecord {
	encoding := base64.RawURLEncoding
	return storage.PkarrRecord{
		V:   encoding.EncodeToString(p.V),
		K:   encoding.EncodeToString(p.K[:]),
		Sig: encoding.EncodeToString(p.Sig[:]),
		Seq: p.Seq,
	}
}

// PublishPkarr stores the record in the db, publishes the given Pkarr record to the DHT, and returns the z-base-32 encoded ID
func (s *PkarrService) PublishPkarr(ctx context.Context, id string, request PublishPkarrRequest) error {
	if err := request.isValid(); err != nil {
		return err
	}

	// write to db and cache
	// TODO(gabe): if putting to the DHT fails we should note that in the db and retry later
	record := request.toRecord()
	if err := s.db.WriteRecord(record); err != nil {
		return err
	}
	s.cache.Set(id, record, ttlcache.DefaultTTL)

	// return here and put it in the DHT asynchronously
	go s.dht.Put(ctx, bep44.Put{
		V:   request.V,
		K:   &request.K,
		Sig: request.Sig,
		Seq: request.Seq,
	})

	return nil
}

// GetPkarrResponse is the response to a get Pkarr request
type GetPkarrResponse struct {
	V   []byte   `validate:"required"`
	Seq int64    `validate:"required"`
	Sig [64]byte `validate:"required"`
}

func fromPkarrRecord(record storage.PkarrRecord) (*GetPkarrResponse, error) {
	encoding := base64.RawURLEncoding
	vBytes, err := encoding.DecodeString(record.V)
	if err != nil {
		return nil, err
	}
	sigBytes, err := encoding.DecodeString(record.Sig)
	if err != nil {
		return nil, err
	}
	return &GetPkarrResponse{
		V:   vBytes,
		Seq: record.Seq,
		Sig: [64]byte(sigBytes),
	}, nil
}

// GetPkarr returns the full Pkarr record (including sig data) for the given z-base-32 encoded ID
func (s *PkarrService) GetPkarr(ctx context.Context, id string) (*GetPkarrResponse, error) {
	// first do a cache lookup
	if s.cache.Has(id) {
		cacheItem := s.cache.Get(id)
		logrus.Debugf("resolved pkarr record<%s> from cache, with ttl<%s>", id, cacheItem.TTL().String())
		return fromPkarrRecord(cacheItem.Value())
	}

	// next do a dht lookup
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
		return fromPkarrRecord(*record)
	}
	bBytes, err := got.V.MarshalBencode()
	if err != nil {
		return nil, err
	}
	var payload string
	if err = bencode.Unmarshal(bBytes, &payload); err != nil {
		return nil, err
	}
	return &GetPkarrResponse{
		V:   []byte(payload),
		Seq: got.Seq,
		Sig: got.Sig,
	}, nil
}

// TODO(gabe) make this more efficient. create a publish schedule based on each individual record, not all records
func (s *PkarrService) republish() {
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

func recordToBEP44Put(record storage.PkarrRecord) (*bep44.Put, error) {
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
