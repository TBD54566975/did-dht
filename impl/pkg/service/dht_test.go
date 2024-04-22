package service

import (
	"context"
	"fmt"
	"os"
	"testing"

	anacrolixdht "github.com/anacrolix/dht/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/TBD54566975/did-dht-method/config"
	"github.com/TBD54566975/did-dht-method/internal/did"
	"github.com/TBD54566975/did-dht-method/pkg/dht"
	"github.com/TBD54566975/did-dht-method/pkg/storage"
)

func TestDHTService(t *testing.T) {
	svc := newDHTService(t, "a")

	t.Run("test put bad record", func(t *testing.T) {
		err := svc.PublishDHT(context.Background(), "", dht.BEP44Record{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "validation for 'Value' failed on the 'required' tag")
	})

	t.Run("test get non existent record", func(t *testing.T) {
		got, err := svc.GetDHT(context.Background(), "test")
		assert.NoError(t, err)
		assert.Nil(t, got)
	})

	t.Run("test get record with invalid ID", func(t *testing.T) {
		got, err := svc.GetDHT(context.Background(), "---")
		assert.ErrorContains(t, err, "illegal z-base-32 data at input byte 0")
		assert.Nil(t, got)
	})

	t.Run("test record with a bad signature", func(t *testing.T) {
		// create a did doc as a packet to store
		sk, doc, err := did.GenerateDIDDHT(did.CreateDIDDHTOpts{})
		require.NoError(t, err)
		require.NotEmpty(t, doc)

		d := did.DHT(doc.ID)
		packet, err := d.ToDNSPacket(*doc, nil, nil)
		assert.NoError(t, err)
		assert.NotEmpty(t, packet)

		putMsg, err := dht.CreateDNSPublishRequest(sk, *packet)
		require.NoError(t, err)
		require.NotEmpty(t, putMsg)

		suffix, err := d.Suffix()
		require.NoError(t, err)
		err = svc.PublishDHT(context.Background(), suffix, dht.RecordFromBEP44(putMsg))
		assert.NoError(t, err)

		// invalidate the signature
		putMsg.Sig[0] = 0
		err = svc.PublishDHT(context.Background(), suffix, dht.RecordFromBEP44(putMsg))
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "signature is invalid")
	})

	t.Run("test put and get record", func(t *testing.T) {
		// create a did doc as a packet to store
		sk, doc, err := did.GenerateDIDDHT(did.CreateDIDDHTOpts{})
		require.NoError(t, err)
		require.NotEmpty(t, doc)

		d := did.DHT(doc.ID)
		packet, err := d.ToDNSPacket(*doc, nil, nil)
		assert.NoError(t, err)
		assert.NotEmpty(t, packet)

		putMsg, err := dht.CreateDNSPublishRequest(sk, *packet)
		require.NoError(t, err)
		require.NotEmpty(t, putMsg)

		suffix, err := d.Suffix()
		require.NoError(t, err)
		err = svc.PublishDHT(context.Background(), suffix, dht.RecordFromBEP44(putMsg))
		assert.NoError(t, err)

		got, err := svc.GetDHT(context.Background(), suffix)
		assert.NoError(t, err)
		assert.NotEmpty(t, got)
		assert.Equal(t, putMsg.V, got.V)
		assert.Equal(t, putMsg.Sig, got.Sig)
		assert.Equal(t, putMsg.Seq, got.Seq)
	})

	t.Run("test get uncached record", func(t *testing.T) {
		// create a did doc as a packet to store
		sk, doc, err := did.GenerateDIDDHT(did.CreateDIDDHTOpts{})
		require.NoError(t, err)
		require.NotEmpty(t, doc)

		d := did.DHT(doc.ID)
		packet, err := d.ToDNSPacket(*doc, nil, nil)
		require.NoError(t, err)
		require.NotEmpty(t, packet)

		putMsg, err := dht.CreateDNSPublishRequest(sk, *packet)
		require.NoError(t, err)
		require.NotEmpty(t, putMsg)

		suffix, err := d.Suffix()
		require.NoError(t, err)
		err = svc.PublishDHT(context.Background(), suffix, dht.RecordFromBEP44(putMsg))
		require.NoError(t, err)

		// remove it from the cache so the get tests the uncached lookup path
		err = svc.cache.Delete(suffix)
		require.NoError(t, err)

		got, err := svc.GetDHT(context.Background(), suffix)
		assert.NoError(t, err)
		assert.NotEmpty(t, got)
		assert.Equal(t, putMsg.V, got.V)
		assert.Equal(t, putMsg.Sig, got.Sig)
		assert.Equal(t, putMsg.Seq, got.Seq)
	})

	t.Run("test get record with invalid ID", func(t *testing.T) {
		got, err := svc.GetDHT(context.Background(), "uqaj3fcr9db6jg6o9pjs53iuftyj45r46aubogfaceqjbo6pp9sy")
		assert.NoError(t, err)
		assert.Empty(t, got)

		// try it again to make sure the cache is working
		got, err = svc.GetDHT(context.Background(), "uqaj3fcr9db6jg6o9pjs53iuftyj45r46aubogfaceqjbo6pp9sy")
		assert.ErrorContains(t, err, "rate limited to prevent spam")
		assert.Empty(t, got)
	})

	t.Cleanup(func() { svc.Close() })
}

func TestDHT(t *testing.T) {
	svc1 := newDHTService(t, "b")

	// create and publish a record to service1
	sk, doc, err := did.GenerateDIDDHT(did.CreateDIDDHTOpts{})
	require.NoError(t, err)
	require.NotEmpty(t, doc)
	d := did.DHT(doc.ID)
	packet, err := d.ToDNSPacket(*doc, nil, nil)
	require.NoError(t, err)
	require.NotEmpty(t, packet)
	putMsg, err := dht.CreateDNSPublishRequest(sk, *packet)
	require.NoError(t, err)
	require.NotEmpty(t, putMsg)
	suffix, err := d.Suffix()
	require.NoError(t, err)
	err = svc1.PublishDHT(context.Background(), suffix, dht.RecordFromBEP44(putMsg))
	require.NoError(t, err)

	// make sure we can get it back
	got, err := svc1.GetDHT(context.Background(), suffix)
	require.NoError(t, err)
	require.NotEmpty(t, got)
	assert.Equal(t, putMsg.V, got.V)
	assert.Equal(t, putMsg.Sig, got.Sig)
	assert.Equal(t, putMsg.Seq, got.Seq)

	// create service2 with service1 as a bootstrap peer
	svc2 := newDHTService(t, "c", anacrolixdht.NewAddr(svc1.dht.Addr()))

	// get the record via service2
	gotFrom2, err := svc2.GetDHT(context.Background(), suffix)
	require.NoError(t, err)
	require.NotEmpty(t, gotFrom2)
	assert.Equal(t, putMsg.V, gotFrom2.V)
	assert.Equal(t, putMsg.Sig, gotFrom2.Sig)
	assert.Equal(t, putMsg.Seq, gotFrom2.Seq)

	t.Cleanup(func() {
		svc1.Close()
		svc2.Close()
	})
}

func TestNoConfig(t *testing.T) {
	svc, err := NewDHTService(nil, nil, nil)
	assert.EqualError(t, err, "config is required")
	assert.Empty(t, svc)

	svc, err = NewDHTService(&config.Config{
		DHTConfig: config.DHTServiceConfig{
			CacheSizeLimitMB: -1,
		},
	}, nil, nil)
	assert.EqualError(t, err, "failed to instantiate cache: HardMaxCacheSize must be >= 0")
	assert.Nil(t, svc)

	svc, err = NewDHTService(&config.Config{
		DHTConfig: config.DHTServiceConfig{
			RepublishCRON: "not a real cron expression",
		},
	}, nil, nil)
	assert.EqualError(t, err, "failed to start republisher: gocron: cron expression failed to be parsed: failed to parse int from not: strconv.Atoi: parsing \"not\": invalid syntax")
	assert.Nil(t, svc)

	t.Cleanup(func() { svc.Close() })
}

func newDHTService(t *testing.T, id string, bootstrapPeers ...anacrolixdht.Addr) DHTService {
	defaultConfig := config.GetDefaultConfig()

	db, err := storage.NewStorage(fmt.Sprintf("bolt://diddht-test-%s.db", id))
	require.NoError(t, err)
	require.NotEmpty(t, db)

	t.Cleanup(func() { os.Remove(fmt.Sprintf("diddht-test-%s.db", id)) })

	d := dht.NewTestDHT(t, bootstrapPeers...)
	dhtService, err := NewDHTService(&defaultConfig, db, d)
	require.NoError(t, err)
	require.NotEmpty(t, dhtService)

	return *dhtService
}
