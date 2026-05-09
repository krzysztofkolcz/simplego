package router

import (
	"encoding/json"
	"net/http"

	httpapi "github.com/krzysztofkolcz/mymigrations/internal/http/api"
	"github.com/krzysztofkolcz/mymigrations/internal/http/middleware"
)

// New builds the full http.Handler:
//
//	Recovery → RequestID → Logging → OAPI strict handler
func New(strict httpapi.StrictServerInterface) http.Handler {
	return httpapi.HandlerWithOptions(
		httpapi.NewStrictHandlerWithOptions(
			strict,
			nil,
			httpapi.StrictHTTPServerOptions{
				RequestErrorHandlerFunc:  requestErrorHandler,
				ResponseErrorHandlerFunc: responseErrorHandler,
			},
		),
		httpapi.StdHTTPServerOptions{
			BaseURL:    "/v1",
			BaseRouter: http.NewServeMux(),
			Middlewares: []httpapi.MiddlewareFunc{
				middleware.Logging(),
				middleware.RequestID(),
				middleware.Recovery(),
			},
			ErrorHandlerFunc: paramsErrorHandler,
		},
	)
}

// requestErrorHandler is called when JSON body decoding fails.
func requestErrorHandler(w http.ResponseWriter, _ *http.Request, err error) {
	writeError(w, 400, "BAD_REQUEST", err.Error())
}

// responseErrorHandler is called when a handler returns a non-nil error.
func responseErrorHandler(w http.ResponseWriter, _ *http.Request, err error) {
	writeError(w, 500, "INTERNAL_SERVER_ERROR", err.Error())
}

// paramsErrorHandler is called when path/header parameter binding fails.
func paramsErrorHandler(w http.ResponseWriter, _ *http.Request, err error) {
	writeError(w, 400, "BAD_REQUEST", err.Error())
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"error": map[string]any{
			"status":  status,
			"code":    code,
			"message": message,
		},
	})
}
