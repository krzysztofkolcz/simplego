package myhttpcontroller2

import (
	"context"

	"github.com/google/uuid"
	"github.com/krzysztofkolcz/my-http-server-002/internal/api/myhttpserver2"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

func (c *APIController) GetTest(ctx context.Context,
	request myhttpserver2.GetTestRequestObject) (myhttpserver2.GetTestResponseObject, error) {
	parsed, _ := uuid.Parse("7d4ef8d4-9073-483f-a2d5-bc55ab4c9faa")
	resp := myhttpserver2.GetTest200JSONResponse{
		Id:          openapi_types.UUID(parsed),
		Description: "description",
	}

	return resp, nil

}
