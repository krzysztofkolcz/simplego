package main

import (
	"log"
	"log/slog"
	"net/http"
	"os"
	"time"

	slogctx "github.com/veqryn/slog-context"
)

func main() {
	logger := initLogger()
	mux := http.NewServeMux()

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	loggedMux := loggingMiddleware(logger)(mux)

	log.Println("Starting server on :8080")

	err := http.ListenAndServe(":8080", loggedMux)
	if err != nil {
		log.Fatal(err)
	}
}

func initLogger() *slog.Logger {
	return slog.New(
		slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		}),
	)
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