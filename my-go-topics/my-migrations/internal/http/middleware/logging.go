package middleware

import (
	"log/slog"
	"net/http"
	"time"

	slogctx "github.com/veqryn/slog-context"
)

func Logging() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := slogctx.With(r.Context(),
				"method", r.Method,
				"path", r.URL.Path,
			)
			r = r.WithContext(ctx)

			slogctx.Info(ctx, "request received")

			start := time.Now()
			lrw := &statusWriter{ResponseWriter: w, status: http.StatusOK}
			next.ServeHTTP(lrw, r)

			slogctx.Info(ctx, "request completed",
				slog.Int("status", lrw.status),
				slog.Duration("duration", time.Since(start)),
			)
		})
	}
}

type statusWriter struct {
	http.ResponseWriter
	status int
}

func (sw *statusWriter) WriteHeader(code int) {
	sw.status = code
	sw.ResponseWriter.WriteHeader(code)
}
