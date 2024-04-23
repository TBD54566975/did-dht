package service

import (
	"context"
	"sync"
	"time"

	ssiutil "github.com/TBD54566975/ssi-sdk/util"
	"github.com/allegro/bigcache/v3"
	"github.com/anacrolix/torrent/bencode"
	"github.com/goccy/go-json"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

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
	logrus.WithContext(ctx).WithField("record_id", id).Debug("added dht record to cache and db")

	// return here and put it in the DHT asynchronously
	go func() {
		// Create a new context with a timeout so that the parent context does not cancel the put
		putCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if _, err = s.dht.Put(putCtx, record.Put()); err != nil {
			logrus.WithContext(ctx).WithField("record_id", id).WithError(err).Warnf("error from dht.Put for record: %s", id)
		} else {
			logrus.WithContext(ctx).WithField("record_id", id).Debug("put record to DHT")
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
		logrus.WithContext(ctx).WithError(err).WithField("record_id", id).Warn("failed to get record from cache, falling back to dht")
	}

	// next do a dht lookup with a timeout of 10 seconds
	getCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	got, err := s.dht.GetFull(getCtx, id)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			logrus.WithContext(ctx).WithField("record_id", id).Warn("dht lookup timed out, attempting to resolve from storage")
		} else {
			logrus.WithContext(ctx).WithError(err).WithField("record_id", id).Warn("failed to get record from dht, attempting to resolve from storage")
		}

		record, err := s.db.ReadRecord(ctx, id)
		if err != nil || record == nil {
			logrus.WithContext(ctx).WithError(err).WithField("record_id", id).Error("failed to resolve record from storage; adding to bad get cache")

			// add the key to the badGetCache to prevent spamming the DHT
			if err = s.badGetCache.Set(id, []byte{0}); err != nil {
				logrus.WithContext(ctx).WithError(err).WithField("record_id", id).Error("failed to set key in bad get cache")
			}

			return nil, err
		}

		logrus.WithContext(ctx).WithField("record_id", id).Info("resolved record from storage")
		resp := record.Response()
		// add the record back to the cache for future lookups
		if err = s.addRecordToCache(id, record.Response()); err != nil {
			logrus.WithError(err).WithField("record_id", id).Error("failed to set record in cache")
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
		logrus.WithContext(ctx).WithField("record_id", id).WithError(err).Error("failed to set record in cache")
	} else {
		logrus.WithContext(ctx).WithField("record_id", id).Info("added record back to cache")
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

// failedRecord is a struct to keep track of records that failed to be republished
type failedRecord struct {
	record     dht.BEP44Record
	failureCnt int
}

// TODO(gabe) make this more efficient. create a publish schedule based on each individual record, not all records
// republish republishes all records in the db
func (s *DHTService) republish() {
	ctx, span := telemetry.GetTracer().Start(context.Background(), "DHTService.republish")
	defer span.End()

	recordCnt, err := s.db.RecordCount(ctx)
	if err != nil {
		logrus.WithContext(ctx).WithError(err).Error("failed to get record count before republishing")
		return
	}
	logrus.WithContext(ctx).WithField("record_count", recordCnt).Info("republishing records")

	// republish all records in the db and handle failed records up to 3 times
	failedRecords := s.republishRecords(ctx)

	// handle failed records
	logrus.WithContext(ctx).WithField("failed_record_count", len(failedRecords)).Info("handling failed records")
	s.handleFailedRecords(ctx, failedRecords)
}

// republishRecords republishes all records in the db and returns a list of failed records to be retried
func (s *DHTService) republishRecords(ctx context.Context) []failedRecord {
	var nextPageToken []byte
	var seenRecords, batchCnt int32
	var failedRecords []failedRecord
	var recordsBatch []dht.BEP44Record
	var err error

	var wg sync.WaitGroup

	for {
		recordsBatch, nextPageToken, err = s.db.ListRecords(ctx, nextPageToken, 1000)
		if err != nil {
			logrus.WithContext(ctx).WithError(err).Error("failed to list record(s) for republishing")
			continue
		}

		batchSize := len(recordsBatch)
		seenRecords += int32(batchSize)
		if batchSize == 0 {
			logrus.WithContext(ctx).Info("no records to republish")
			break
		}

		logrus.WithContext(ctx).WithFields(logrus.Fields{
			"record_count": batchSize,
			"batch_number": batchCnt,
			"total_seen":   seenRecords,
		}).Debugf("republishing batch [%d] of [%d] records", batchCnt, batchSize)
		batchCnt++

		batchFailedRecords := s.republishBatch(ctx, &wg, recordsBatch)
		failedRecords = append(failedRecords, batchFailedRecords...)

		if nextPageToken == nil {
			break
		}
	}

	wg.Wait()

	successRate := float64(seenRecords-int32(len(failedRecords))) / float64(seenRecords) * 100
	logrus.WithContext(ctx).WithFields(logrus.Fields{
		"success": seenRecords - int32(len(failedRecords)),
		"errors":  len(failedRecords),
		"total":   seenRecords,
	}).Infof("republishing complete with [%d] batches of [%d] total records with a [%.2f] percent success rate", batchCnt, seenRecords, successRate)

	return failedRecords
}

// republishBatch republishes a batch of records and returns a list of failed records to be retried
func (s *DHTService) republishBatch(ctx context.Context, wg *sync.WaitGroup, recordsBatch []dht.BEP44Record) []failedRecord {
	failedRecordsChan := make(chan failedRecord, len(recordsBatch))
	var failedRecords []failedRecord

	for _, record := range recordsBatch {
		wg.Add(1)
		go func(record dht.BEP44Record) {
			defer wg.Done()

			id := record.ID()
			putCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
			defer cancel()

			if _, putErr := s.dht.Put(putCtx, record.Put()); putErr != nil {
				if errors.Is(putErr, context.DeadlineExceeded) {
					logrus.WithContext(putCtx).WithField("record_id", id).Debug("republish timeout exceeded")
				} else {
					logrus.WithContext(putCtx).WithField("record_id", id).WithError(putErr).Debug("failed to republish record")
				}
				failedRecordsChan <- failedRecord{
					record:     record,
					failureCnt: 1,
				}
			}
		}(record)
	}

	wg.Wait()
	close(failedRecordsChan)

	for fr := range failedRecordsChan {
		failedRecords = append(failedRecords, fr)
	}
	return failedRecords
}

// handleFailedRecords attempts to republish failed records up to 3 times
func (s *DHTService) handleFailedRecords(ctx context.Context, failedRecords []failedRecord) {
	for _, fr := range failedRecords {
		retryCount := 0
		for retryCount < 3 {
			id := fr.record.ID()
			putCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
			defer cancel()

			if _, putErr := s.dht.Put(putCtx, fr.record.Put()); putErr != nil {
				logrus.WithContext(putCtx).WithField("record_id", id).WithError(putErr).Debugf("failed to re-republish [%s], attempt: %d", id, retryCount+1)
				retryCount++
			} else {
				break
			}
		}

		if retryCount == 3 {
			id := fr.record.ID()
			logrus.WithContext(ctx).WithField("record_id", id).Error("record failed to republish after 3 attempts")
			if err := s.db.WriteFailedRecord(ctx, id); err != nil {
				logrus.WithContext(ctx).WithField("record_id", id).WithError(err).Warn("failed to write failed record to db")
			}
		}
	}

	failedRecordCnt, err := s.db.FailedRecordCount(ctx)
	if err != nil {
		logrus.WithContext(ctx).WithError(err).Error("failed to get failed record count")
		return
	}

	logrus.WithContext(ctx).WithField("failed_record_count", failedRecordCnt).Warn("total count of record that failed to republish")
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
			logrus.WithError(err).Error("failed to close bad get cache")
		}
	}
	if err := s.db.Close(); err != nil {
		logrus.WithError(err).Error("failed to close db")
	}
	if s.dht != nil {
		s.dht.Close()
	}
}
