package handlers

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	todoapi "github.com/C5383717/my-todo/internal/api/todo"
	"github.com/C5383717/my-todo/internal/api/write"
	"github.com/C5383717/my-todo/internal/apierrors"
	"github.com/C5383717/my-todo/internal/log"
	"github.com/C5383717/my-todo/utils"
	md "github.com/oapi-codegen/nethttp-middleware"
	slogctx "github.com/veqryn/slog-context"
)

// OAPIValidatorHandler is called when OAPI required fields are missing from Request
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

// ParamsErrorHandler is called whenever Request doesn't follow OAPI Endpoint Parameters (Path and Query).
// Must create RequestID and inject logger because middlewares weren't run yet.
func ParamsErrorHandler() func(w http.ResponseWriter, r *http.Request, err error) {
	return func(w http.ResponseWriter, r *http.Request, err error) {
		log.Info(r.Context(), "ParamsErrorHandler")
		ctx := utils.InjectRequestID(r.Context())
		requestID, _ := utils.GetRequestID(ctx)

		ctx = slogctx.With(ctx,
			slog.String("RequestID", requestID),
		)

		log.Error(ctx, "The error encountered during parameters binding", err)

		var errorResponse todoapi.ErrorMessage

		var (
			invalidFormatErr     *todoapi.InvalidParamFormatError
			requiredHeaderErr    *todoapi.RequiredHeaderError
			tooManyParametersErr *todoapi.TooManyValuesForParamError
			requiredParamErr     *todoapi.RequiredParamError
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

// RequestErrorHandlerFunc is called when Request JSON Body decoding fails
func RequestErrorHandlerFunc() func(w http.ResponseWriter, r *http.Request, err error) {
	return func(w http.ResponseWriter, r *http.Request, err error) {
		log.Info(r.Context(), "RequestErrorHandlerFunc")
		log.Error(r.Context(), "Receiving Request", err)

		write.ErrorResponse(r.Context(), w, apierrors.JSONDecodeErrorMessage())
	}
}

// ResponseErrorHandlerFunc is called when HTTP handlers return an error.
// This is where internal errors are transformed into safe API responses.
// apierrors.TransformToAPIError does the mapping — no internal error leaks.
func ResponseErrorHandlerFunc() func(w http.ResponseWriter, r *http.Request, err error) {
	return func(w http.ResponseWriter, r *http.Request, err error) {
		log.Info(r.Context(), "ResponseErrorHandlerFunc")
		log.Error(r.Context(), "Processing Response", err)

		e := apierrors.TransformToAPIError(r.Context(), err)
		write.ErrorResponse(r.Context(), w, *e)
	}
}
