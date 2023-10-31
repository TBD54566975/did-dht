package service

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/TBD54566975/did-dht-method/config"
	"github.com/TBD54566975/did-dht-method/internal/did"
	"github.com/TBD54566975/did-dht-method/pkg/dht"
	"github.com/TBD54566975/did-dht-method/pkg/storage"
)

func TestPKARRService(t *testing.T) {
	svc := newPKARRService(t)
	require.NotEmpty(t, svc)

	t.Run("test put bad record", func(t *testing.T) {
		err := svc.PublishPKARR(context.Background(), PublishPKARRRequest{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "validation for 'V' failed on the 'required' tag")
	})

	t.Run("test get non existent record", func(t *testing.T) {
		got, err := svc.GetPKARR(context.Background(), "test")
		assert.NoError(t, err)
		assert.Nil(t, got)
	})

	t.Run("test record with a bad signature", func(t *testing.T) {
		// create a did doc as a packet to store
		sk, doc, err := did.GenerateDIDDHT(did.CreateDIDDHTOpts{})
		require.NoError(t, err)
		require.NotEmpty(t, doc)

		d := did.DHT(doc.ID)
		packet, err := d.ToDNSPacket(*doc, nil)
		assert.NoError(t, err)
		assert.NotEmpty(t, packet)

		putMsg, err := dht.CreatePKARRPublishRequest(sk, *packet)
		require.NoError(t, err)
		require.NotEmpty(t, putMsg)

		err = svc.PublishPKARR(context.Background(), PublishPKARRRequest{
			V:   putMsg.V.([]byte),
			K:   *putMsg.K,
			Sig: putMsg.Sig,
			Seq: putMsg.Seq,
		})
		assert.NoError(t, err)

		// invalidate the signature
		putMsg.Sig[0] = 0
		err = svc.PublishPKARR(context.Background(), PublishPKARRRequest{
			V:   putMsg.V.([]byte),
			K:   *putMsg.K,
			Sig: putMsg.Sig,
			Seq: putMsg.Seq,
		})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "signature is invalid")
	})

	t.Run("test put and get record", func(t *testing.T) {
		// create a did doc as a packet to store
		sk, doc, err := did.GenerateDIDDHT(did.CreateDIDDHTOpts{})
		require.NoError(t, err)
		require.NotEmpty(t, doc)

		d := did.DHT(doc.ID)
		packet, err := d.ToDNSPacket(*doc, nil)
		assert.NoError(t, err)
		assert.NotEmpty(t, packet)

		putMsg, err := dht.CreatePKARRPublishRequest(sk, *packet)
		require.NoError(t, err)
		require.NotEmpty(t, putMsg)

		err = svc.PublishPKARR(context.Background(), PublishPKARRRequest{
			V:   putMsg.V.([]byte),
			K:   *putMsg.K,
			Sig: putMsg.Sig,
			Seq: putMsg.Seq,
		})
		assert.NoError(t, err)

		suffix, err := d.Suffix()
		assert.NoError(t, err)

		// wait for the record to be published
		time.Sleep(10 * time.Second)

		got, err := svc.GetPKARR(context.Background(), suffix)
		assert.NoError(t, err)
		assert.NotEmpty(t, got)
		assert.Equal(t, putMsg.V, got.V)
		assert.Equal(t, putMsg.Sig, got.Sig)
		assert.Equal(t, putMsg.Seq, got.Seq)
	})
}

func newPKARRService(t *testing.T) PKARRService {
	defaultConfig := config.GetDefaultConfig()
	db, err := storage.NewStorage(defaultConfig.ServerConfig.DBFile)
	require.NoError(t, err)
	require.NotEmpty(t, db)
	pkarrService, err := NewPKARRService(&defaultConfig, db)
	require.NoError(t, err)
	require.NotEmpty(t, pkarrService)
	return *pkarrService
}
