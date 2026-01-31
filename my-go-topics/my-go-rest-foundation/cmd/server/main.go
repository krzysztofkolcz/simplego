package main

import (
	"log/slog"
	"net/http"
	"os"
	"time"

	app "github.com/krzysztofkolcz/my-go-rest-foundation/internal"
	apphttp "github.com/krzysztofkolcz/my-go-rest-foundation/internal/http"
	slogctx "github.com/veqryn/slog-context"
)

func main() {
	logger := slog.New(
		slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		}),
	)

	application := app.New()
	router := apphttp.NewRouter(application)


	handler := loggingMiddleware(logger)(router)

	logger.Info("server started", "port", 8080)
	http.ListenAndServe(":8080", handler)
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