package utils

import (
	"context"
	"errors"

	"github.com/C5383717/my-todo/internal/errs"
	"github.com/google/uuid"
)

var (
	ErrGetRequestID = errors.New("no requestID found in context")
)

type key string

const requestID = key("requestID")

func InjectRequestID(ctx context.Context) context.Context {
	return context.WithValue(ctx, requestID, uuid.NewString())
}

func GetRequestID(ctx context.Context) (string, error) {
	id, ok := ctx.Value(requestID).(string)
	if !ok || id == "" {
		return "", errs.Wrap(ErrGetRequestID, nil)
	}

	return id, nil
}
