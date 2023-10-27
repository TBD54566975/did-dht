package storage

import (
	"encoding/base64"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/TBD54566975/did-dht-method/internal/did"
	"github.com/TBD54566975/did-dht-method/pkg/dht"
)

func TestPKARRStorage(t *testing.T) {
	db := setupBoltDB(t)
	defer db.Close()
	require.NotEmpty(t, db)

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
	encoding := base64.RawURLEncoding
	record := PKARRRecord{
		V:   encoding.EncodeToString(putMsg.V.([]byte)),
		K:   encoding.EncodeToString(putMsg.K[:]),
		Sig: encoding.EncodeToString(putMsg.Sig[:]),
		Seq: putMsg.Seq,
	}

	err = db.WriteRecord(record)
	assert.NoError(t, err)

	// read it back
	readRecord, err := db.ReadRecord(record.K)
	assert.NoError(t, err)
	assert.Equal(t, record, *readRecord)

	// list and confirm it's there
	records, err := db.ListRecords()
	assert.NoError(t, err)
	assert.NotEmpty(t, records)
	assert.Equal(t, record, records[0])
}
