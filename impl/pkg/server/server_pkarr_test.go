package server

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/TBD54566975/did-dht-method/config"
	"github.com/TBD54566975/did-dht-method/pkg/service"
	"github.com/TBD54566975/did-dht-method/pkg/storage"
)

func TestPKARRRouter(t *testing.T) {
	pkarrSvc := testPKARRService(t)
	pkarrRouter, err := NewPKARRRouter(&pkarrSvc)
	require.NoError(t, err)
	require.NotEmpty(t, pkarrRouter)

	pkarrRouter.PutRecord(nil)
}

func testPKARRService(t *testing.T) service.PKARRService {
	defaultConfig := config.GetDefaultConfig()
	db, err := storage.NewStorage(defaultConfig.ServerConfig.DBFile)
	require.NoError(t, err)
	require.NotEmpty(t, db)
	pkarrService, err := service.NewPKARRService(&defaultConfig, db)
	require.NoError(t, err)
	require.NotEmpty(t, pkarrService)
	return *pkarrService
}
