package utils

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/krzysztofkolcz/my-http-server/internal/constants"
	"github.com/krzysztofkolcz/my-http-server/internal/errs"
)

var (
	ErrExtractTenantID   = errors.New("could not extract tenant ID from context")
	ErrGetRequestID      = errors.New("no requestID found in context")
	ErrExtractClientData = errors.New("could not extract client data from context")
	ErrTenantInvalid     = errors.New("invalid tenant")
)

type key string

const requestID = key("requestID")

func InjectRequestID(ctx context.Context) context.Context {
	return context.WithValue(ctx, requestID, uuid.NewString())
}

func GetRequestID(ctx context.Context) (string, error) {
	requestID, ok := ctx.Value(requestID).(string)
	if !ok || requestID == "" {
		return "", ErrGetRequestID
	}

	return requestID, nil
}

func ExtractTenantID(ctx context.Context) (string, error) {
	// tenantID, ok := ctx.Value(nethttp.TenantKey).(string)
	// if !ok || tenantID == "" {
	// 	return "", errs.Wrap(ErrExtractTenantID, nethttp.ErrTenantInvalid)
	// }

	tenantID, ok := ctx.Value(constants.TenantCtxKey).(string)
	if !ok || tenantID == "" {
		return "", errs.Wrap(ErrExtractTenantID, ErrTenantInvalid)
	}

	return tenantID, nil
}
