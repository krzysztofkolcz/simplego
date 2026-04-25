package middleware

import (
	"net/http"

	"github.com/C5383717/my-todo/utils"
)

// InjectRequestID injects a RequestID into the context to be used by other middlewares
func InjectRequestID() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			next.ServeHTTP(w, r.WithContext(utils.InjectRequestID(r.Context())))
		})
	}
}
