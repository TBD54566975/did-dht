package service

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/TBD54566975/did-dht-method/config"
	"github.com/TBD54566975/did-dht-method/pkg/storage"
)

func TestGatewayService(t *testing.T) {
	svc := newGatewayService(t)
	require.NotEmpty(t, svc)

	t.Run("Publish and Get a DID", func(t *testing.T) {

	})
}

func newGatewayService(t *testing.T) GatewayService {
	defaultConfig := config.GetDefaultConfig()
	db, err := storage.NewStorage(defaultConfig.ServerConfig.DBFile)
	require.NoError(t, err)
	require.NotEmpty(t, db)
	pkarrService, err := NewPkarrService(&defaultConfig, db)
	require.NoError(t, err)
	require.NotEmpty(t, pkarrService)

	gatewayService, err := NewGatewayService(&defaultConfig, db, pkarrService)
	require.NoError(t, err)
	require.NotEmpty(t, gatewayService)
	return *gatewayService
}
