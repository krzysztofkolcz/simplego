package middleware

import (
	"net/http"

	"github.com/google/uuid"
	slogctx "github.com/veqryn/slog-context"
)

type contextKey string

const requestIDKey contextKey = "requestID"

func RequestID() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			id := uuid.New().String()
			ctx := slogctx.With(r.Context(), "requestID", id)
			w.Header().Set("X-Request-ID", id)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
