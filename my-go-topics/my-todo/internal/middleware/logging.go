package middleware

import (
	"log/slog"
	"net/http"
	"time"

	logger "github.com/C5383717/my-todo/internal/log"
)

// LoggingMiddleware logs the start and end of each request, along with the duration and status code.
func LoggingMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := logger.InjectRequest(r.Context(), r)
			r = r.WithContext(ctx)

			logger.Info(ctx, "Received Request")

			start := time.Now()
			lrw := &loggingResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}
			next.ServeHTTP(lrw, r)

			duration := time.Since(start)

			logger.Info(ctx, "Request Completed",
				slog.Int("HttpStatus", lrw.statusCode),
				slog.Duration("Duration", duration),
			)
		})
	}
}

type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (lrw *loggingResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}
