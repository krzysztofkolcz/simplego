package middleware

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"runtime/debug"

	slogctx "github.com/veqryn/slog-context"
)

func Recovery() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rec := recover(); rec != nil {
					ctx := r.Context()
					slogctx.Error(ctx, "panic recovered",
						slog.String("panic", fmt.Sprintf("%v", rec)),
						slog.String("stack", string(debug.Stack())),
					)

					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusInternalServerError)
					_ = json.NewEncoder(w).Encode(map[string]any{
						"error": map[string]any{
							"status":  500,
							"code":    "INTERNAL_SERVER_ERROR",
							"message": "Internal server error",
						},
					})
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}
