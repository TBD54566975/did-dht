package bolt

import (
	"context"
	"os"
	"testing"

	"github.com/TBD54566975/did-dht-method/internal/did"
	"github.com/TBD54566975/did-dht-method/pkg/dht"
	"github.com/TBD54566975/did-dht-method/pkg/pkarr"
	"github.com/goccy/go-json"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBoltDB_ReadWrite(t *testing.T) {
	db := setupBoltDB(t)

	// create a name space and a message in it
	namespace := "F1"

	team1 := "Red Bull"
	players1 := []string{"Max Verstappen", "Sergio Pérez"}
	p1Bytes, err := json.Marshal(players1)
	assert.NoError(t, err)

	err = db.write(namespace, team1, p1Bytes)
	assert.NoError(t, err)

	// get it back
	gotPlayers1, err := db.read(namespace, team1)
	assert.NoError(t, err)

	var players1Result []string
	err = json.Unmarshal(gotPlayers1, &players1Result)
	assert.NoError(t, err)
	assert.EqualValues(t, players1, players1Result)

	// get a value from a dhtNamespace that doesn't exist
	res, err := db.read("bad", "worse")
	assert.NoError(t, err)
	assert.Empty(t, res)

	// get a value that doesn't exist in the dhtNamespace
	noValue, err := db.read(namespace, "Porsche")
	assert.NoError(t, err)
	assert.Empty(t, noValue)

	// create a second value in the dhtNamespace
	team2 := "McLaren"
	players2 := []string{"Lando Norris", "Daniel Ricciardo"}
	p2Bytes, err := json.Marshal(players2)
	assert.NoError(t, err)

	err = db.write(namespace, team2, p2Bytes)
	assert.NoError(t, err)

	// get all values from the dhtNamespace
	gotAll, err := db.readAll(namespace)
	assert.NoError(t, err)
	assert.True(t, len(gotAll) == 2)

	_, gotRedBull := gotAll[team1]
	assert.True(t, gotRedBull)

	_, gotMcLaren := gotAll[team2]
	assert.True(t, gotMcLaren)
}

func TestBoltDB_PrefixAndKeys(t *testing.T) {
	db := setupBoltDB(t)

	namespace := "blockchains"

	// set up prefix read test

	dummyData := []byte("dummy")
	err := db.write(namespace, "bitcoin-testnet", dummyData)
	assert.NoError(t, err)

	err = db.write(namespace, "bitcoin-mainnet", dummyData)
	assert.NoError(t, err)

	err = db.write(namespace, "tezos-testnet", dummyData)
	assert.NoError(t, err)

	err = db.write(namespace, "tezos-mainnet", dummyData)
	assert.NoError(t, err)
}

func setupBoltDB(t *testing.T) *boltdb {
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
	record := pkarr.RecordFromBep44(putMsg)

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
