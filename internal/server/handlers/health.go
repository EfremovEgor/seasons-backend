package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type Pinger interface {
	PingContext(ctx context.Context) error
}

type HealthHandler struct {
	db Pinger
}

func NewHealthHandler(db Pinger) *HealthHandler {
	return &HealthHandler{db: db}
}

type healthResponse struct {
	Status    string            `json:"status"`
	Timestamp time.Time         `json:"timestamp"`
	Services  map[string]string `json:"services"`
}

func (h *HealthHandler) Handle(c *gin.Context) {
	resp := healthResponse{
		Status:    "ok",
		Timestamp: time.Now().UTC(),
		Services:  make(map[string]string),
	}
	httpStatus := http.StatusOK

	if h.db == nil {
		resp.Services["database"] = "not configured"
	} else {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
		defer cancel()

		if err := h.db.PingContext(ctx); err != nil {
			resp.Services["database"] = "unavailable"
			resp.Status = "degraded"
			httpStatus = http.StatusServiceUnavailable
		} else {
			resp.Services["database"] = "ok"
		}
	}

	c.JSON(httpStatus, resp)
}
