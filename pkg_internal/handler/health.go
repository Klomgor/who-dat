package handler

import (
	"net/http"
	"time"

	"github.com/lissy93/who-dat/pkg_internal/response"
	"github.com/lissy93/who-dat/pkg/types"
)

var startTime = time.Now()

// HealthHandler handles health check requests
type HealthHandler struct {
	version string
}

// NewHealthHandler creates a new health check handler
func NewHealthHandler(version string) *HealthHandler {
	return &HealthHandler{
		version: version,
	}
}

func (h *HealthHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	uptime := time.Since(startTime).Seconds()

	healthResponse := types.HealthResponse{
		Status:  "ok",
		Version: h.version,
		Uptime:  int64(uptime),
	}

	response.JSON(w, http.StatusOK, healthResponse)
}
