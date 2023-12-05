package server

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/TBD54566975/did-dht-method/config"
	"github.com/TBD54566975/did-dht-method/internal/did"
	"github.com/TBD54566975/did-dht-method/pkg/dht"
	"github.com/TBD54566975/did-dht-method/pkg/service"
	"github.com/TBD54566975/did-dht-method/pkg/storage"
)

func TestPKARRRouter(t *testing.T) {
	pkarrSvc := testPKARRService(t)
	pkarrRouter, err := NewPkarrRouter(&pkarrSvc)
	require.NoError(t, err)
	require.NotEmpty(t, pkarrRouter)

	t.Run("test put record", func(t *testing.T) {
		didID, reqData := generateDIDPutRequest(t)

		w := httptest.NewRecorder()
		suffix, err := did.DHT(didID).Suffix()
		assert.NoError(t, err)
		req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("%s/%s", testServerURL, suffix), bytes.NewReader(reqData))
		c := newRequestContextWithParams(w, req, map[string]string{IDParam: suffix})

		pkarrRouter.PutRecord(c)
		assert.True(t, is2xxResponse(w.Code))
	})

	t.Run("test get record", func(t *testing.T) {
		didID, reqData := generateDIDPutRequest(t)

		w := httptest.NewRecorder()
		suffix, err := did.DHT(didID).Suffix()
		assert.NoError(t, err)
		req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("%s/%s", testServerURL, suffix), bytes.NewReader(reqData))
		c := newRequestContextWithParams(w, req, map[string]string{IDParam: suffix})

		pkarrRouter.PutRecord(c)
		assert.True(t, is2xxResponse(w.Code))

		w = httptest.NewRecorder()
		req = httptest.NewRequest(http.MethodGet, fmt.Sprintf("%s/%s", testServerURL, suffix), nil)
		c = newRequestContextWithParams(w, req, map[string]string{IDParam: suffix})

		pkarrRouter.GetRecord(c)
		assert.True(t, is2xxResponse(w.Code))

		resp, err := io.ReadAll(w.Body)
		assert.NoError(t, err)
		assert.NotEmpty(t, resp)
		assert.Equal(t, reqData, resp)
	})
}

func testPKARRService(t *testing.T) service.PkarrService {
	defaultConfig := config.GetDefaultConfig()
	db, err := storage.NewStorage(defaultConfig.ServerConfig.DBFile)
	require.NoError(t, err)
	require.NotEmpty(t, db)
	pkarrService, err := service.NewPkarrService(&defaultConfig, db)
	require.NoError(t, err)
	require.NotEmpty(t, pkarrService)
	return *pkarrService
}

func generateDIDPutRequest(t *testing.T) (string, []byte) {
	// generate a DID Document
	sk, doc, err := did.GenerateDIDDHT(did.CreateDIDDHTOpts{})
	require.NoError(t, err)
	require.NotEmpty(t, doc)

	packet, err := did.DHT(doc.ID).ToDNSPacket(*doc, nil)
	assert.NoError(t, err)
	assert.NotEmpty(t, packet)

	bep44Put, err := dht.CreatePKARRPublishRequest(sk, *packet)
	assert.NoError(t, err)
	assert.NotEmpty(t, bep44Put)

	// prepare request as sig:seq:v
	var seqBuf [8]byte
	binary.BigEndian.PutUint64(seqBuf[:], uint64(bep44Put.Seq))
	return doc.ID, append(bep44Put.Sig[:], append(seqBuf[:], bep44Put.V.([]byte)...)...)
}
