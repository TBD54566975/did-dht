package service

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	ssiutil "github.com/TBD54566975/ssi-sdk/util"
	"github.com/allegro/bigcache/v3"
	"github.com/anacrolix/torrent/bencode"
	"github.com/goccy/go-json"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/tv42/zbase32"

	"github.com/TBD54566975/did-dht-method/internal/util"

	"github.com/TBD54566975/did-dht-method/config"
	dhtint "github.com/TBD54566975/did-dht-method/internal/dht"
	"github.com/TBD54566975/did-dht-method/pkg/dht"
	"github.com/TBD54566975/did-dht-method/pkg/storage"
	"github.com/TBD54566975/did-dht-method/pkg/telemetry"
)

const recordSizeLimitBytes = 1000

// DHTService is the service responsible for managing BEP44 DNS records in the DHT and reading/writing records
type DHTService struct {
	cfg         *config.Config
	db          storage.Storage
	dht         *dht.DHT
	cache       *bigcache.BigCache
	badGetCache *bigcache.BigCache
	scheduler   *dhtint.Scheduler
}

// NewDHTService returns a new instance of the DHT service
func NewDHTService(cfg *config.Config, db storage.Storage, d *dht.DHT) (*DHTService, error) {
	if cfg == nil {
		return nil, ssiutil.LoggingNewError("config is required")
	}

	// create and start get cache
	cacheTTL := time.Duration(cfg.DHTConfig.CacheTTLSeconds) * time.Second
	cacheConfig := bigcache.DefaultConfig(cacheTTL)
	cacheConfig.MaxEntrySize = recordSizeLimitBytes
	cacheConfig.HardMaxCacheSize = cfg.DHTConfig.CacheSizeLimitMB
	cacheConfig.CleanWindow = cacheTTL / 2
	cache, err := bigcache.New(context.Background(), cacheConfig)
	if err != nil {
		return nil, ssiutil.LoggingErrorMsg(err, "failed to instantiate cache")
	}

	// create a new cache for bad gets to prevent spamming the DHT
	cacheConfig.LifeWindow = 60 * time.Second
	cacheConfig.CleanWindow = 30 * time.Second
	badGetCache, err := bigcache.New(context.Background(), cacheConfig)
	if err != nil {
		return nil, ssiutil.LoggingErrorMsg(err, "failed to instantiate badGetCache")
	}

	// start scheduler for republishing
	scheduler := dhtint.NewScheduler()
	svc := DHTService{
		cfg:         cfg,
		db:          db,
		dht:         d,
		cache:       cache,
		badGetCache: badGetCache,
		scheduler:   &scheduler,
	}
	if err = scheduler.Schedule(cfg.DHTConfig.RepublishCRON, svc.republish); err != nil {
		return nil, ssiutil.LoggingErrorMsg(err, "failed to start republisher")
	}
	return &svc, nil
}

// PublishDHT stores the record in the db, publishes the given DNS record to the DHT, and returns the z-base-32 encoded ID
func (s *DHTService) PublishDHT(ctx context.Context, id string, record dht.BEP44Record) error {
	ctx, span := telemetry.GetTracer().Start(ctx, "DHTService.PublishDHT")
	defer span.End()

	// make sure the key is valid
	if _, err := util.Z32Decode(id); err != nil {
		return ssiutil.LoggingCtxErrorMsgf(ctx, err, "failed to decode z-base-32 encoded ID: %s", id)
	}

	if err := record.IsValid(); err != nil {
		return err
	}

	// check if the message is already in the cache
	if got, err := s.cache.Get(id); err == nil {
		var resp dht.BEP44Response
		if err = json.Unmarshal(got, &resp); err == nil && record.Response().Equals(resp) {
			logrus.WithContext(ctx).WithField("record_id", id).Debug("resolved dht record from cache with matching response")
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
	logrus.WithContext(ctx).WithField("record", id).Debug("added dht record to cache and db")

	// return here and put it in the DHT asynchronously
	go func() {
		// Create a new context with a timeout so that the parent context does not cancel the put
		putCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if _, err = s.dht.Put(putCtx, record.Put()); err != nil {
			logrus.WithContext(ctx).WithError(err).Errorf("error from dht.Put for record: %s", id)
		} else {
			logrus.WithContext(ctx).WithField("record", id).Debug("put record to DHT")
		}
	}()

	return nil
}

// GetDHT returns the full DNS record (including sig data) for the given z-base-32 encoded ID
func (s *DHTService) GetDHT(ctx context.Context, id string) (*dht.BEP44Response, error) {
	ctx, span := telemetry.GetTracer().Start(ctx, "DHTService.GetDHT")
	defer span.End()

	// make sure the key is valid
	if _, err := util.Z32Decode(id); err != nil {
		return nil, ssiutil.LoggingCtxErrorMsgf(ctx, err, "failed to decode z-base-32 encoded ID: %s", id)
	}

	// if the key is in the badGetCache, return an error
	if _, err := s.badGetCache.Get(id); err == nil {
		return nil, ssiutil.LoggingCtxErrorMsgf(ctx, err, "bad key [%s] rate limited to prevent spam", id)
	}

	// first do a cache lookup
	if got, err := s.cache.Get(id); err == nil {
		var resp dht.BEP44Response
		if err = json.Unmarshal(got, &resp); err == nil {
			logrus.WithContext(ctx).WithField("record_id", id).Info("resolved record from cache")
			return &resp, nil
		}
		logrus.WithContext(ctx).WithError(err).WithField("record", id).Warn("failed to get record from cache, falling back to dht")
	}

	// next do a dht lookup with a timeout of 10 seconds
	getCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	got, err := s.dht.GetFull(getCtx, id)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			logrus.WithContext(ctx).WithField("record", id).Warn("dht lookup timed out, attempting to resolve from storage")
		} else {
			logrus.WithContext(ctx).WithError(err).WithField("record", id).Warn("failed to get record from dht, attempting to resolve from storage")
		}

		rawID, err := util.Z32Decode(id)
		if err != nil {
			return nil, err
		}

		record, err := s.db.ReadRecord(ctx, rawID)
		if err != nil || record == nil {
			logrus.WithContext(ctx).WithError(err).WithField("record", id).Error("failed to resolve record from storage; adding to badGetCache")

			// add the key to the badGetCache to prevent spamming the DHT
			if err = s.badGetCache.Set(id, []byte{0}); err != nil {
				logrus.WithContext(ctx).WithError(err).WithField("record", id).Error("failed to set key in badGetCache")
			}

			return nil, err
		}

		logrus.WithContext(ctx).WithField("record", id).Info("resolved record from storage")
		resp := record.Response()
		// add the record back to the cache for future lookups
		if err = s.addRecordToCache(id, record.Response()); err != nil {
			logrus.WithError(err).WithField("record", id).Error("failed to set record in cache")
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
	resp := dht.BEP44Response{
		V:   []byte(payload),
		Seq: got.Seq,
		Sig: got.Sig,
	}

	// add the record to cache, do it here to avoid duplicate calculations
	if err = s.addRecordToCache(id, resp); err != nil {
		logrus.WithContext(ctx).WithField("record", id).WithError(err).Error("failed to set record in cache")
	} else {
		logrus.WithContext(ctx).WithField("record", id).Info("added record back to cache")
	}

	return &resp, nil
}

func (s *DHTService) addRecordToCache(id string, resp dht.BEP44Response) error {
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
func (s *DHTService) republish() {
	ctx, span := telemetry.GetTracer().Start(context.Background(), "DHTService.republish")
	defer span.End()

	recordCnt, err := s.db.RecordCount(ctx)
	if err != nil {
		logrus.WithContext(ctx).WithError(err).Error("failed to get record count before republishing")
		return
	} else {
		logrus.WithContext(ctx).WithField("record_count", recordCnt).Info("republishing records")
	}

	var nextPageToken []byte
	var recordsBatch []dht.BEP44Record
	var seenRecords, batchCnt, successCnt, errCnt int32 = 0, 1, 0, 0

	for {
		recordsBatch, nextPageToken, err = s.db.ListRecords(ctx, nextPageToken, 1000)
		if err != nil {
			logrus.WithContext(ctx).WithError(err).Error("failed to list record(s) for republishing")
			return
		}
		batchSize := len(recordsBatch)
		seenRecords += int32(batchSize)
		if batchSize == 0 {
			logrus.WithContext(ctx).Info("no records to republish")
			return
		}

		logrus.WithContext(ctx).WithFields(logrus.Fields{
			"record_count": batchSize,
			"batch_number": batchCnt,
			"total_seen":   seenRecords,
		}).Infof("republishing batch [%d] of [%d] records", batchCnt, batchSize)
		batchCnt++

		var wg sync.WaitGroup
		wg.Add(batchSize)

		var batchErrCnt, batchSuccessCnt int32 = 0, 0
		for _, record := range recordsBatch {
			go func(ctx context.Context, record dht.BEP44Record) {
				defer wg.Done()

				recordID := zbase32.EncodeToString(record.Key[:])
				logrus.WithContext(ctx).Debugf("republishing record: %s", recordID)

				putCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
				defer cancel()

				if _, putErr := s.dht.Put(putCtx, record.Put()); putErr != nil {
					logrus.WithContext(putCtx).WithError(putErr).Debugf("failed to republish record: %s", recordID)
					atomic.AddInt32(&batchErrCnt, 1)
				} else {
					atomic.AddInt32(&batchSuccessCnt, 1)
				}
			}(ctx, record)
		}

		// Wait for all goroutines in this batch to finish before moving on to the next batch
		wg.Wait()

		// Update the success and error counts
		atomic.AddInt32(&successCnt, batchSuccessCnt)
		atomic.AddInt32(&errCnt, batchErrCnt)

		successRate := float64(batchSuccessCnt) / float64(batchSize)

		logrus.WithContext(ctx).WithFields(logrus.Fields{
			"batch_number": batchCnt,
			"success":      successCnt,
			"errors":       errCnt,
		}).Infof("batch [%d] completed with a [%.2f] percent success rate", batchCnt, successRate*100)

		if successRate < 0.8 {
			logrus.WithContext(ctx).WithFields(logrus.Fields{
				"batch_number": batchCnt,
				"success":      successCnt,
				"errors":       errCnt,
			}).Errorf("batch [%d] failed to meet success rate threshold; exiting republishing early", batchCnt)
			break
		}

		if nextPageToken == nil {
			break
		}
	}

	successRate := float64(successCnt) / float64(seenRecords)
	logrus.WithContext(ctx).WithFields(logrus.Fields{
		"success": seenRecords - errCnt,
		"errors":  errCnt,
		"total":   seenRecords,
	}).Infof("republishing complete with [%d] batches of [%d] total records with an [%.2f] percent success rate", batchCnt, seenRecords, successRate*100)
}

// Close closes the Mainline service gracefully
func (s *DHTService) Close() {
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
	if s.badGetCache != nil {
		if err := s.badGetCache.Close(); err != nil {
			logrus.WithError(err).Error("failed to close badGetCache")
		}
	}
	if err := s.db.Close(); err != nil {
		logrus.WithError(err).Error("failed to close db")
	}
	if s.dht != nil {
		s.dht.Close()
	}
}
