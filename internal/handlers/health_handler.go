package handlers

import (
	"net/http"

	"ticket-system/internal/httpx"
)

// HealthHandler serves the public health check.
type HealthHandler struct{}

// NewHealthHandler is the constructor. Go convention: `NewX` returns an X.
func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

func (h *HealthHandler) Check(w http.ResponseWriter, r *http.Request) {
	httpx.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
