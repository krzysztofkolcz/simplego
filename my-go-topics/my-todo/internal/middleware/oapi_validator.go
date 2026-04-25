package middleware

import (
	"net/http"

	"github.com/C5383717/my-todo/internal/handlers"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"

	md "github.com/oapi-codegen/nethttp-middleware"
)

// OAPIMiddleware validates a Request against the OpenAPI Spec
func OAPIMiddleware(swagger *openapi3.T) func(next http.Handler) http.Handler {
	openapi3filter.RegisterBodyDecoder(
		"application/merge-patch+json",
		openapi3filter.JSONBodyDecoder,
	)

	return md.OapiRequestValidatorWithOptions(
		swagger, &md.Options{
			ErrorHandlerWithOpts: handlers.OAPIValidatorHandler,
			Options: openapi3filter.Options{
				AuthenticationFunc:    openapi3filter.NoopAuthenticationFunc,
				IncludeResponseStatus: true,
			},
			SilenceServersWarning: true,
		},
	)
}
