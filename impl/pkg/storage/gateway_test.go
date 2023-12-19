package storage

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/TBD54566975/did-dht-method/internal/did"
)

func TestGatewayStorage(t *testing.T) {
	db := setupBoltDB(t)
	defer db.Close()
	require.NotEmpty(t, db)

	t.Run("Read and Write DID", func(t *testing.T) {
		// create a did doc to store
		_, doc, err := did.GenerateDIDDHT(did.CreateDIDDHTOpts{})
		require.NoError(t, err)
		require.NotEmpty(t, doc)

		// create record
		record := GatewayRecord{
			Document:       *doc,
			Types:          []did.TypeIndex{1, 2, 3},
			SequenceNumber: 1,
		}

		err = db.WriteDID(record)
		assert.NoError(t, err)

		// get it back
		readRecord, err := db.ReadDID(record.Document.ID)
		assert.NoError(t, err)
		assert.Equal(t, record, *readRecord)
	})

	t.Run("Update a DID and its type indexes", func(t *testing.T) {
		// create a did doc to store
		_, doc, err := did.GenerateDIDDHT(did.CreateDIDDHTOpts{})
		require.NoError(t, err)
		require.NotEmpty(t, doc)

		// create record
		record := GatewayRecord{
			Document:       *doc,
			Types:          []did.TypeIndex{1, 2, 3},
			SequenceNumber: 1,
		}

		err = db.WriteDID(record)
		assert.NoError(t, err)

		// get types
		types, err := db.ListDIDsForType(1)
		assert.NoError(t, err)
		assert.Equal(t, []string{record.Document.ID}, types)

		types, err = db.ListDIDsForType(2)
		assert.NoError(t, err)
		assert.Equal(t, []string{record.Document.ID}, types)

		types, err = db.ListDIDsForType(3)
		assert.NoError(t, err)
		assert.Equal(t, []string{record.Document.ID}, types)

		// update record
		record.Types = []did.TypeIndex{4, 5, 6}
		record.SequenceNumber = 2
		err = db.WriteDID(record)
		assert.NoError(t, err)

		// get it back
		readRecord, err := db.ReadDID(record.Document.ID)
		assert.NoError(t, err)
		assert.Equal(t, record, *readRecord)

		// get types
		types, err = db.ListDIDsForType(1)
		assert.NoError(t, err)
		assert.Empty(t, types)

		types, err = db.ListDIDsForType(2)
		assert.NoError(t, err)
		assert.Empty(t, types)

		types, err = db.ListDIDsForType(3)
		assert.NoError(t, err)
		assert.Empty(t, types)

		types, err = db.ListDIDsForType(4)
		assert.NoError(t, err)
		assert.Equal(t, []string{record.Document.ID}, types)

		types, err = db.ListDIDsForType(5)
		assert.NoError(t, err)
		assert.Equal(t, []string{record.Document.ID}, types)

		types, err = db.ListDIDsForType(6)
		assert.NoError(t, err)
		assert.Equal(t, []string{record.Document.ID}, types)
	})

	t.Run("Multiple DIDs with Types", func(t *testing.T) {
		// create a did doc to store
		_, doc, err := did.GenerateDIDDHT(did.CreateDIDDHTOpts{})
		require.NoError(t, err)
		require.NotEmpty(t, doc)

		// create record
		record := GatewayRecord{
			Document:       *doc,
			Types:          []did.TypeIndex{1, 2, 3},
			SequenceNumber: 1,
		}

		err = db.WriteDID(record)
		assert.NoError(t, err)

		// create a did doc to store
		_, doc2, err := did.GenerateDIDDHT(did.CreateDIDDHTOpts{})
		require.NoError(t, err)
		require.NotEmpty(t, doc)

		// create record
		record2 := GatewayRecord{
			Document:       *doc2,
			Types:          []did.TypeIndex{2, 3, 4},
			SequenceNumber: 1,
		}

		err = db.WriteDID(record2)
		assert.NoError(t, err)

		// get types
		types, err := db.ListDIDsForType(1)
		assert.NoError(t, err)
		assert.Equal(t, []string{record.Document.ID}, types)

		types, err = db.ListDIDsForType(2)
		assert.NoError(t, err)
		assert.Equal(t, []string{record.Document.ID, record2.Document.ID}, types)

		types, err = db.ListDIDsForType(3)
		assert.NoError(t, err)
		assert.Equal(t, []string{record.Document.ID, record2.Document.ID}, types)

		types, err = db.ListDIDsForType(4)
		assert.NoError(t, err)
		assert.Equal(t, []string{record2.Document.ID}, types)
	})
}
