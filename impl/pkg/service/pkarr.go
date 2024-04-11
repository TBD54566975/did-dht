package service

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"github.com/goccy/go-json"
	"github.com/tv42/zbase32"

	ssiutil "github.com/TBD54566975/ssi-sdk/util"
	"github.com/allegro/bigcache/v3"
	"github.com/anacrolix/torrent/bencode"
	"github.com/sirupsen/logrus"

	"github.com/TBD54566975/did-dht-method/internal/util"

	"github.com/TBD54566975/did-dht-method/config"
	dhtint "github.com/TBD54566975/did-dht-method/internal/dht"
	"github.com/TBD54566975/did-dht-method/pkg/dht"
	"github.com/TBD54566975/did-dht-method/pkg/pkarr"
	"github.com/TBD54566975/did-dht-method/pkg/storage"
	"github.com/TBD54566975/did-dht-method/pkg/telemetry"
)

const recordSizeLimit = 1000

// PkarrService is the Pkarr service responsible for managing the Pkarr DHT and reading/writing records
type PkarrService struct {
	cfg       *config.Config
	db        storage.Storage
	dht       *dht.DHT
	cache     *bigcache.BigCache
	scheduler *dhtint.Scheduler
}

// NewPkarrService returns a new instance of the Pkarr service
func NewPkarrService(cfg *config.Config, db storage.Storage, d *dht.DHT) (*PkarrService, error) {
	if cfg == nil {
		return nil, ssiutil.LoggingNewError("config is required")
	}

	// create and start cache and scheduler
	cacheTTL := time.Duration(cfg.PkarrConfig.CacheTTLSeconds) * time.Second
	cacheConfig := bigcache.DefaultConfig(cacheTTL)
	cacheConfig.MaxEntrySize = recordSizeLimit
	cacheConfig.HardMaxCacheSize = cfg.PkarrConfig.CacheSizeLimitMB
	cacheConfig.CleanWindow = cacheTTL / 2
	cache, err := bigcache.New(context.Background(), cacheConfig)
	if err != nil {
		return nil, ssiutil.LoggingErrorMsg(err, "failed to instantiate cache")
	}
	scheduler := dhtint.NewScheduler()
	svc := PkarrService{
		cfg:       cfg,
		db:        db,
		dht:       d,
		cache:     cache,
		scheduler: &scheduler,
	}
	if err = scheduler.Schedule(cfg.PkarrConfig.RepublishCRON, svc.republish); err != nil {
		return nil, ssiutil.LoggingErrorMsg(err, "failed to start republisher")
	}
	return &svc, nil
}

// PublishPkarr stores the record in the db, publishes the given Pkarr record to the DHT, and returns the z-base-32 encoded ID
func (s *PkarrService) PublishPkarr(ctx context.Context, id string, record pkarr.Record) error {
	ctx, span := telemetry.GetTracer().Start(ctx, "PkarrService.PublishPkarr")
	defer span.End()

	if err := record.IsValid(); err != nil {
		return err
	}

	// check if the message is already in the cache
	if got, err := s.cache.Get(id); err == nil {
		var resp pkarr.Response
		if err = json.Unmarshal(got, &resp); err == nil && record.Response().Equals(resp) {
			logrus.WithContext(ctx).WithField("record_id", id).Debug("resolved pkarr record from cache with matching response")
			return nil
		}
	}

	// write to db and cache
	if err := s.db.WriteRecord(ctx, record); err != nil {
		return err
	}
	recordBytes, err := json.Marshal(record.Response())
	if err != nil {
		return err
	}
	if err = s.cache.Set(id, recordBytes); err != nil {
		return err
	}

	// return here and put it in the DHT asynchronously
	// TODO(gabe): consider a background process to monitor failures
	go func() {
		// Create a new context with a timeout so that the parent context does not cancel the put
		putCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if _, err = s.dht.Put(putCtx, record.BEP44()); err != nil {
			logrus.WithContext(ctx).WithError(err).Errorf("error from dht.Put for record: %s", id)
		}
	}()

	return nil
}

// GetPkarr returns the full Pkarr record (including sig data) for the given z-base-32 encoded ID
func (s *PkarrService) GetPkarr(ctx context.Context, id string) (*pkarr.Response, error) {
	ctx, span := telemetry.GetTracer().Start(ctx, "PkarrService.GetPkarr")
	defer span.End()

	// first do a cache lookup
	if got, err := s.cache.Get(id); err == nil {
		var resp pkarr.Response
		if err = json.Unmarshal(got, &resp); err == nil {
			logrus.WithContext(ctx).WithField("record_id", id).Debug("resolved pkarr record from cache")
			return &resp, nil
		}
		logrus.WithContext(ctx).WithError(err).WithField("record", id).Warn("failed to get pkarr record from cache, falling back to dht")
	}

	// next do a dht lookup
	got, err := s.dht.GetFull(ctx, id)
	if err != nil {
		// try to resolve from storage before returning and error
		logrus.WithContext(ctx).WithError(err).WithField("record", id).Warn("failed to get pkarr record from dht, attempting to resolve from storage")

		rawID, err := util.Z32Decode(id)
		if err != nil {
			return nil, err
		}

		record, err := s.db.ReadRecord(ctx, rawID)
		if err != nil || record == nil {
			logrus.WithContext(ctx).WithError(err).WithField("record", id).Error("failed to resolve pkarr record from storage")
			return nil, err
		}

		logrus.WithContext(ctx).WithField("record", id).Debug("resolved pkarr record from storage")
		resp := record.Response()
		if err = s.addRecordToCache(id, record.Response()); err != nil {
			logrus.WithError(err).WithField("record", id).Error("failed to set pkarr record in cache")
		}

		return &resp, err
	}

	// prepare the record for return
	bBytes, err := got.V.MarshalBencode()
	if err != nil {
		return nil, err
	}
	var payload string
	if err = bencode.Unmarshal(bBytes, &payload); err != nil {
		return nil, ssiutil.LoggingCtxErrorMsg(ctx, err, "failed to unmarshal bencoded payload")
	}
	resp := pkarr.Response{
		V:   []byte(payload),
		Seq: got.Seq,
		Sig: got.Sig,
	}

	// add the record to cache, do it here to avoid duplicate calculations
	if err = s.addRecordToCache(id, resp); err != nil {
		logrus.WithContext(ctx).WithError(err).Errorf("failed to set pkarr record[%s] in cache", id)
	}

	return &resp, nil
}

func (s *PkarrService) addRecordToCache(id string, resp pkarr.Response) error {
	recordBytes, err := json.Marshal(resp)
	if err != nil {
		return err
	}
	if err = s.cache.Set(id, recordBytes); err != nil {
		return err
	}
	return nil
}

// TODO(gabe) make this more efficient. create a publish schedule based on each individual record, not all records
func (s *PkarrService) republish() {
	ctx, span := telemetry.GetTracer().Start(context.Background(), "PkarrService.republish")
	defer span.End()

	recordCnt, err := s.db.RecordCount(ctx)
	if err != nil {
		logrus.WithContext(ctx).WithError(err).Error("failed to get record count before republishing")
	} else {
		logrus.WithContext(ctx).WithField("record_count", recordCnt).Info("republishing records")
	}

	var nextPageToken []byte
	var recordsBatch []pkarr.Record
	var seenRecords, batchCnt, successCnt, errCnt int32 = 0, 0, 0, 0
	for {
		recordsBatch, nextPageToken, err = s.db.ListRecords(ctx, nextPageToken, 1000)
		if err != nil {
			logrus.WithContext(ctx).WithError(err).Error("failed to list record(s) for republishing")
			return
		}
		seenRecords += int32(len(recordsBatch))
		if len(recordsBatch) == 0 {
			logrus.WithContext(ctx).Info("no records to republish")
			return
		}

		logrus.WithContext(ctx).WithFields(logrus.Fields{
			"record_count": len(recordsBatch),
			"batch_number": batchCnt,
			"total_seen":   seenRecords,
		}).Infof("republishing next batch of records")
		batchCnt++

		var wg sync.WaitGroup
		wg.Add(len(recordsBatch))

		for _, record := range recordsBatch {
			go func(record pkarr.Record) {
				defer wg.Done()

				recordID := zbase32.EncodeToString(record.Key[:])
				logrus.WithContext(ctx).Debugf("republishing record: %s", recordID)
				if _, err = s.dht.Put(ctx, record.BEP44()); err != nil {
					logrus.WithContext(ctx).WithError(err).Errorf("failed to republish record: %s", recordID)
					atomic.AddInt32(&errCnt, 1)
				} else {
					atomic.AddInt32(&successCnt, 1)
				}
			}(record)
		}

		wg.Wait()

		if nextPageToken == nil {
			break
		}
	}

	logrus.WithContext(ctx).WithFields(logrus.Fields{
		"success": seenRecords - errCnt,
		"errors":  errCnt,
		"total":   seenRecords,
	}).Infof("republishing complete with [%d] batches", batchCnt)
}

// Close closes the Pkarr service gracefully
func (s *PkarrService) Close() {
	if s == nil {
		return
	}
	if s.scheduler != nil {
		s.scheduler.Stop()
	}
	if s.cache != nil {
		if err := s.cache.Close(); err != nil {
			logrus.WithError(err).Error("failed to close cache")
		}
	}
	if err := s.db.Close(); err != nil {
		logrus.WithError(err).Error("failed to close db")
	}
	if s.dht != nil {
		s.dht.Close()
	}
}
