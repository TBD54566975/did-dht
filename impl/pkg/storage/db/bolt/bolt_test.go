package bolt

import (
	"context"
	"os"
	"testing"

	"github.com/goccy/go-json"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/TBD54566975/did-dht-method/internal/did"
	"github.com/TBD54566975/did-dht-method/pkg/dht"
)

func TestBoltDB_ReadWrite(t *testing.T) {
	ctx := context.Background()

	db := getTestDB(t)

	// create a name space and a message in it
	namespace := "F1"

	team1 := "Red Bull"
	players1 := []string{"Max Verstappen", "Sergio Pérez"}
	p1Bytes, err := json.Marshal(players1)
	assert.NoError(t, err)

	err = db.write(ctx, namespace, team1, p1Bytes)
	assert.NoError(t, err)

	// get it back
	gotPlayers1, err := db.read(ctx, namespace, team1)
	assert.NoError(t, err)

	var players1Result []string
	err = json.Unmarshal(gotPlayers1, &players1Result)
	assert.NoError(t, err)
	assert.EqualValues(t, players1, players1Result)

	// get a value from a oldDHTNamespace that doesn't exist
	res, err := db.read(ctx, "bad", "worse")
	assert.NoError(t, err)
	assert.Empty(t, res)

	// get a value that doesn't exist in the oldDHTNamespace
	noValue, err := db.read(ctx, namespace, "Porsche")
	assert.NoError(t, err)
	assert.Empty(t, noValue)

	// create a second value in the oldDHTNamespace
	team2 := "McLaren"
	players2 := []string{"Lando Norris", "Daniel Ricciardo"}
	p2Bytes, err := json.Marshal(players2)
	assert.NoError(t, err)

	err = db.write(ctx, namespace, team2, p2Bytes)
	assert.NoError(t, err)

	// get all values from the oldDHTNamespace
	gotAll, err := db.readAll(namespace)
	assert.NoError(t, err)
	assert.True(t, len(gotAll) == 2)

	_, gotRedBull := gotAll[team1]
	assert.True(t, gotRedBull)

	_, gotMcLaren := gotAll[team2]
	assert.True(t, gotMcLaren)
}

func TestBoltDB_PrefixAndKeys(t *testing.T) {
	ctx := context.Background()

	db := getTestDB(t)

	namespace := "blockchains"

	// set up prefix read test

	dummyData := []byte("dummy")
	err := db.write(ctx, namespace, "bitcoin-testnet", dummyData)
	assert.NoError(t, err)

	err = db.write(ctx, namespace, "bitcoin-mainnet", dummyData)
	assert.NoError(t, err)

	err = db.write(ctx, namespace, "tezos-testnet", dummyData)
	assert.NoError(t, err)

	err = db.write(ctx, namespace, "tezos-mainnet", dummyData)
	assert.NoError(t, err)
}

func getTestDB(t *testing.T) *Bolt {
	path := "test.db"
	db, err := NewBolt(path)
	assert.NoError(t, err)
	assert.NotEmpty(t, db)

	t.Cleanup(func() {
		_ = db.Close()
		_ = os.Remove(path)
	})
	return db
}

func TestReadWrite(t *testing.T) {
	db := getTestDB(t)
	ctx := context.Background()

	beforeCnt, err := db.RecordCount(ctx)
	require.NoError(t, err)

	// create a did doc as a packet to store
	sk, doc, err := did.GenerateDIDDHT(did.CreateDIDDHTOpts{})
	require.NoError(t, err)
	require.NotEmpty(t, doc)

	packet, err := did.DHT(doc.ID).ToDNSPacket(*doc, nil, nil)
	require.NoError(t, err)
	require.NotEmpty(t, packet)

	putMsg, err := dht.CreateDNSPublishRequest(sk, *packet)
	require.NoError(t, err)
	require.NotEmpty(t, putMsg)

	r := dht.RecordFromBEP44(putMsg)

	err = db.WriteRecord(ctx, r)
	require.NoError(t, err)

	r2, err := db.ReadRecord(ctx, r.ID())
	require.NoError(t, err)

	assert.Equal(t, r.Key, r2.Key)
	assert.Equal(t, r.Value, r2.Value)
	assert.Equal(t, r.Signature, r2.Signature)
	assert.Equal(t, r.SequenceNumber, r2.SequenceNumber)

	afterCnt, err := db.RecordCount(ctx)
	assert.NoError(t, err)
	assert.Equal(t, beforeCnt+1, afterCnt)
}

func TestDBPagination(t *testing.T) {
	db := getTestDB(t)
	defer db.Close()

	ctx := context.Background()

	beforeCnt, err := db.RecordCount(ctx)
	require.NoError(t, err)

	preTestRecords, _, err := db.ListRecords(ctx, nil, 10)
	require.NoError(t, err)

	// store 10 records
	for i := 0; i < 10; i++ {
		// create a did doc as a packet to store
		sk, doc, err := did.GenerateDIDDHT(did.CreateDIDDHTOpts{})
		require.NoError(t, err)
		require.NotEmpty(t, doc)

		packet, err := did.DHT(doc.ID).ToDNSPacket(*doc, nil, nil)
		assert.NoError(t, err)
		assert.NotEmpty(t, packet)

		putMsg, err := dht.CreateDNSPublishRequest(sk, *packet)
		require.NoError(t, err)
		require.NotEmpty(t, putMsg)

		// create record
		record := dht.RecordFromBEP44(putMsg)

		err = db.WriteRecord(ctx, record)
		assert.NoError(t, err)
	}

	// store 11th document
	// create a did doc as a packet to store
	sk, doc, err := did.GenerateDIDDHT(did.CreateDIDDHTOpts{})
	require.NoError(t, err)
	require.NotEmpty(t, doc)

	packet, err := did.DHT(doc.ID).ToDNSPacket(*doc, nil, nil)
	assert.NoError(t, err)
	assert.NotEmpty(t, packet)

	putMsg, err := dht.CreateDNSPublishRequest(sk, *packet)
	require.NoError(t, err)
	require.NotEmpty(t, putMsg)

	// create eleventhRecord
	eleventhRecord := dht.RecordFromBEP44(putMsg)

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

	afterCnt, err := db.RecordCount(ctx)
	assert.NoError(t, err)
	assert.Equal(t, beforeCnt+11, afterCnt)
}

func TestNewBolt(t *testing.T) {
	b, err := NewBolt("")
	assert.Error(t, err)
	assert.Nil(t, b)

	b, err = NewBolt("bolt:///fake/path/bolt.db")
	assert.Error(t, err)
	assert.Nil(t, b)
}
