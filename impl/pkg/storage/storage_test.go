package storage_test

import (
	"net/url"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/TBD54566975/did-dht-method/pkg/storage"
	"github.com/TBD54566975/did-dht-method/pkg/storage/db/bolt"
	"github.com/TBD54566975/did-dht-method/pkg/storage/db/postgres"
)

func TestNewStoragePostgres(t *testing.T) {
	uri := os.Getenv("TEST_DB")
	if uri == "" {
		t.SkipNow()
	}

	u, err := url.Parse(uri)
	require.NoError(t, err)
	if u.Scheme != "postgres" {
		t.SkipNow()
	}

	db, err := storage.NewStorage(uri)
	require.NoError(t, err)
	assert.IsType(t, postgres.Postgres(""), db)
}

func TestNewStorageBolt(t *testing.T) {
	db, err := storage.NewStorage("bolt:///tmp/bolt.db")
	require.NoError(t, err)
	assert.IsType(t, &bolt.Bolt{}, db)
}

func TestNewStorageUnsupported(t *testing.T) {
	db, err := storage.NewStorage("imaginaryDB://a:b@c/d")
	require.Error(t, err)
	assert.Nil(t, db)
}
