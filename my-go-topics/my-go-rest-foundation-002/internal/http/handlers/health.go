package handlers

import (
	"net/http"

	slogctx "github.com/veqryn/slog-context"
)

func Health(w http.ResponseWriter, r *http.Request) {
	logger := slogctx.FromCtx(r.Context())
	logger.Info("health check")

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}
