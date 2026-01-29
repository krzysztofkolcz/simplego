package main

import (
	"log/slog"
	"net/http"
	"os"
	"time"

	slogctx "github.com/veqryn/slog-context"
)

func main() {
	logger := initLogger()

	mux := http.NewServeMux()
	mux.HandleFunc("/health", healthHandler)
	mux.HandleFunc("/ping", pingHandler)

	loggedMux := loggingMiddleware(logger)(mux)

	logger.Info("starting server", "port", 8080)

	err := http.ListenAndServe(":8080", loggedMux)
	if err != nil {
		logger.Error("server failed", "err", err)
	}
}


func initLogger() *slog.Logger {
	return slog.New(
		slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		}),
	)
}

func pingHandler(w http.ResponseWriter, r *http.Request){
	logger := slogctx.FromCtx(r.Context())

	logger.Info("health check called")

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ping"))
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	logger := slogctx.FromCtx(r.Context())

	logger.Info("health check called")

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}

func loggingMiddleware(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			reqLogger := logger.With(
				"method", r.Method,
				"path", r.URL.Path,
			)

			ctx := slogctx.NewCtx(r.Context(), reqLogger)

			start := time.Now()
			next.ServeHTTP(w, r.WithContext(ctx))
			duration := time.Since(start)

			reqLogger.Info("request handled",
				"duration_ms", duration.Milliseconds(),
			)
		})
	}
}