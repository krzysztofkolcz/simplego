package myhttpcontroller

import (
	"context"
)

type APIController struct {
}

func NewAPIController(ctx context.Context) *APIController {
	return &APIController{}
}
