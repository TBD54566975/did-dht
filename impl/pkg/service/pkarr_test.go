package service

import (
	"context"
	"testing"

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
		_, err := svc.PublishPKARR(context.Background(), PublishPKARRRequest{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "validation for 'V' failed on the 'required' tag")
	})

	t.Run("test get non existent record", func(t *testing.T) {
		got, err := svc.GetPKARR(context.Background(), "test")
		assert.NoError(t, err)
		assert.Nil(t, got)
	})

	t.Run("test put and get record", func(t *testing.T) {
		// create a did doc as a packet to store
		sk, doc, err := did.GenerateDIDDHT(did.CreateDIDDHTOpts{})
		require.NoError(t, err)
		require.NotEmpty(t, doc)

		packet, err := did.DHT(doc.ID).ToDNSPacket(*doc, nil)
		assert.NoError(t, err)
		assert.NotEmpty(t, packet)

		putMsg, err := dht.CreatePKARRPublishRequest(sk, *packet)
		require.NoError(t, err)
		require.NotEmpty(t, putMsg)

		id, err := svc.PublishPKARR(context.Background(), PublishPKARRRequest{
			V:   putMsg.V.([]byte),
			K:   *putMsg.K,
			Sig: putMsg.Sig,
			Seq: putMsg.Seq,
		})
		assert.NoError(t, err)
		assert.NotEmpty(t, id)

		got, err := svc.GetPKARR(context.Background(), id)
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
