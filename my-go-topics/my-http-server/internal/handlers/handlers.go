package handlers

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/krzysztofkolcz/my-http-server/internal/api/myhttpserver"
	"github.com/krzysztofkolcz/my-http-server/internal/api/write"
	"github.com/krzysztofkolcz/my-http-server/internal/apierrors"
	"github.com/krzysztofkolcz/my-http-server/internal/log"
	"github.com/krzysztofkolcz/my-http-server/utils"
	md "github.com/oapi-codegen/nethttp-middleware"
	slogctx "github.com/veqryn/slog-context"
)

// OAPIValidatorHandler is called when OAPI Required fields are missing from Request
/*
Wywoływany przez OAPI middleware, gdy:
- brakuje required field
- złe body
- zły content-type

server.go
	httpHandler := myhttpserver.HandlerWithOptions(
	...
		myhttpserver.StdHTTPServerOptions{
		...
			Middlewares: []myhttpserver.MiddlewareFunc{ // Middlewares are applied from last to first
				middleware.OAPIMiddleware(swagger),
				middleware.InjectRequestID(),
			},
		}
	)

./internal/middleware/oapi_validator.go

func OAPIMiddleware(swagger *openapi3.T) func(next http.Handler) http.Handler {
	return md.OapiRequestValidatorWithOptions(
		swagger, &md.Options{
			ErrorHandlerWithOpts: handlers.OAPIValidatorHandler,
			...
		},
	)
}

*/
func OAPIValidatorHandler(
	ctx context.Context,
	err error,
	w http.ResponseWriter,
	_ *http.Request,
	opts md.ErrorHandlerOpts,
) {
	log.Info(ctx, "OAPIValidatorHandler")
	log.Error(ctx, "Request does not follow OAPI contract", err)

	write.ErrorResponse(ctx, w, apierrors.OAPIValidatorErrorMessage(err.Error(), opts.StatusCode))
}

// ParamsErrorHandler is called whenever Request doesn't follow OAPI Endpoint Parameters (Path and Query)
// Must create RequestID and logger because middlewares weren't ran
func ParamsErrorHandler() func(w http.ResponseWriter, r *http.Request, err error) {
	return func(w http.ResponseWriter, r *http.Request, err error) {
		log.Info(r.Context(), "ParamsErrorHandler")
		ctx := utils.InjectRequestID(r.Context())
		requestID, _ := utils.GetRequestID(ctx)

		ctx = slogctx.With(ctx,
			slog.String("RequestID", requestID),
		)

		log.Error(ctx, "The error encountered during parameters binding", err)

		var errorResponse myhttpserver.ErrorMessage

		var (
			invalidFormatErr     *myhttpserver.InvalidParamFormatError
			requiredHeaderErr    *myhttpserver.RequiredHeaderError
			tooManyParametersErr *myhttpserver.TooManyValuesForParamError
			requiredParamErr     *myhttpserver.RequiredParamError
		)

		switch {
		case errors.As(err, &invalidFormatErr):
			errorResponse = apierrors.TooManyParameters(err.Error())
		case errors.As(err, &requiredHeaderErr):
			errorResponse = apierrors.RequiredHeaderError(requiredHeaderErr.Error())
		case errors.As(err, &tooManyParametersErr):
			errorResponse = apierrors.TooManyParameters(tooManyParametersErr.Error())
		case errors.As(err, &requiredParamErr):
			errorResponse = apierrors.RequiredParamError(requiredParamErr.Error())
		default:
			errorResponse = apierrors.InternalServerErrorMessage()
		}

		write.ErrorResponse(ctx, w, errorResponse)
	}
}

// RequestErrorHandlerFunc is called when Request JSON Body Decoding fails
/*
Wywoływany gdy:
JSON niepoprawny
Body niezgodne ze spec
*/
func RequestErrorHandlerFunc() func(w http.ResponseWriter, r *http.Request, err error) {
	return func(w http.ResponseWriter, r *http.Request, err error) {
		log.Info(r.Context(), "RequestErrorHandlerFunc")
		log.Error(r.Context(), "Receiving Request", err)

		write.ErrorResponse(r.Context(), w, apierrors.JSONDecodeErrorMessage())
	}
}

// ResponseErrorHandlerFunc is called when HTTP Handlers (Controller Functions) return invalid responses
/*
ResponseErrorHandlerFunc
Wywoływany gdy:
-kontroler zwróci błąd
-strict handler wykryje niezgodność response ze spec
Tu dzieje się magia:
e := apierrors.TransformToAPIError(r.Context(), err)
*/
func ResponseErrorHandlerFunc() func(w http.ResponseWriter, r *http.Request, err error) {
	return func(w http.ResponseWriter, r *http.Request, err error) {
		log.Info(r.Context(), "ResponseErrorHandlerFunc")
		log.Error(r.Context(), "Processing Response", err)

		e := apierrors.TransformToAPIError(r.Context(), err)
		write.ErrorResponse(r.Context(), w, *e)
	}
}
