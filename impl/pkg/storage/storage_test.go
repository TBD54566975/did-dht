package storage_test

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/TBD54566975/did-dht-method/internal/did"
	"github.com/TBD54566975/did-dht-method/pkg/dht"
	"github.com/TBD54566975/did-dht-method/pkg/pkarr"
	"github.com/TBD54566975/did-dht-method/pkg/storage"
)

func getTestDB(t *testing.T) storage.Storage {
	uri := os.Getenv("TEST_DB")
	if uri == "" {
		uri = fmt.Sprintf("bolt://test-%d.db", rand.Int())
	}

	db, err := storage.NewStorage(uri)
	require.NoError(t, err)
	require.NotEmpty(t, db)

	return db
}
func TestPKARRStorage(t *testing.T) {
	db := getTestDB(t)
	defer db.Close()

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

	// create record
	record := pkarr.RecordFromBEP44(putMsg)

	ctx := context.Background()

	err = db.WriteRecord(ctx, record)
	assert.NoError(t, err)

	// read it back
	readRecord, err := db.ReadRecord(ctx, record.Key[:])
	assert.NoError(t, err)
	assert.Equal(t, record, *readRecord)

	// list and confirm it's there
	records, _, err := db.ListRecords(ctx, nil, 10)
	assert.NoError(t, err)
	assert.NotEmpty(t, records)
	assert.Equal(t, record, records[0])
}

func TestDBPagination(t *testing.T) {
	db := getTestDB(t)
	defer db.Close()

	ctx := context.Background()

	preTestRecords, _, err := db.ListRecords(ctx, nil, 10)
	assert.NoError(t, err)

	// store 10 records
	for i := 0; i < 10; i++ {
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

		// create record
		record := pkarr.RecordFromBEP44(putMsg)

		err = db.WriteRecord(ctx, record)
		assert.NoError(t, err)
	}

	// store 11th document
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

	// create eleventhRecord
	eleventhRecord := pkarr.RecordFromBEP44(putMsg)

	err = db.WriteRecord(ctx, eleventhRecord)
	assert.NoError(t, err)

	// read the first 10 back
	page, nextPageToken, err := db.ListRecords(ctx, nil, 10)
	assert.NoError(t, err)
	assert.Len(t, page, 10)

	page, nextPageToken, err = db.ListRecords(ctx, nextPageToken, 10+len(preTestRecords))
	assert.NoError(t, err)
	assert.Nil(t, nextPageToken)
	assert.Len(t, page, 1+len(preTestRecords))
}
