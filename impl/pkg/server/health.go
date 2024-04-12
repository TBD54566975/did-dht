package server

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/TBD54566975/did-dht-method/pkg/telemetry"
)

type GetHealthCheckResponse struct {
	// Status is always equal to `OK`.
	Status string `json:"status"`
}

const (
	HealthOK string = "OK"
)

// Health godoc
//
//	@Summary		Health Check
//	@Description	Health is a simple handler that always responds with a 200 OK
//	@Tags			Health
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	GetHealthCheckResponse
//	@Router			/health [get]
func Health(c *gin.Context) {
	_, span := telemetry.GetTracer().Start(c, "HealthHTTP.Health")
	defer span.End()

	status := GetHealthCheckResponse{Status: HealthOK}
	Respond(c, status, http.StatusOK)
}
