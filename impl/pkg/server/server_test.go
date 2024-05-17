package server

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/goccy/go-json"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"github.com/TBD54566975/did-dht/config"
	"github.com/TBD54566975/did-dht/pkg/dht"
)

const (
	testServerURL = "https://diddht-service.com"
)

func TestHealthCheckAPI(t *testing.T) {
	shutdown := make(chan os.Signal, 1)
	serviceConfig, err := config.LoadConfig("")
	serviceConfig.ServerConfig.StorageURI = "bolt://health-check.db"
	serviceConfig.ServerConfig.BaseURL = testServerURL
	assert.NoError(t, err)

	server, err := NewServer(serviceConfig, shutdown, dht.NewTestDHT(t))
	assert.NoError(t, err)
	assert.NotEmpty(t, server)

	defer server.Close()

	req := httptest.NewRequest(http.MethodGet, testServerURL+"/health", nil)
	w := httptest.NewRecorder()

	c := newRequestContext(w, req)
	Health(c)
	assert.True(t, is2xxResponse(w.Code))

	var resp GetHealthCheckResponse
	err = json.NewDecoder(w.Body).Decode(&resp)
	assert.NoError(t, err)
	assert.Equal(t, HealthOK, resp.Status)

	shutdown <- os.Interrupt
}

// Is2xxResponse returns true if the given status code is a 2xx response
func is2xxResponse(statusCode int) bool {
	return statusCode/100 == 2
}

// construct a context value as expected by our handler
func newRequestContext(w http.ResponseWriter, req *http.Request) *gin.Context {
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	return c
}

// construct a context value with query params as expected by our handler
func newRequestContextWithParams(w http.ResponseWriter, req *http.Request, params map[string]string) *gin.Context {
	c := newRequestContext(w, req)
	for k, v := range params {
		c.AddParam(k, v)
	}
	return c
}
