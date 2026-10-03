package handler

import (
	"context"
	"net/http"
	"time"
)

type PingService interface {
	Ping(ctx context.Context) error
}

type PingHandler struct {
	svc PingService
}

func NewPingHandler(svc PingService) *PingHandler {
	return &PingHandler{svc: svc}
}

func (h *PingHandler) Ping() http.Handler {
	fn := func(w http.ResponseWriter, req *http.Request) {
		ctx, cancel := context.WithTimeout(req.Context(), 5*time.Second)
		defer cancel()

		if err := h.svc.Ping(ctx); err != nil {
			http.Error(w, "Internal error", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	}

	return http.HandlerFunc(fn)
}
