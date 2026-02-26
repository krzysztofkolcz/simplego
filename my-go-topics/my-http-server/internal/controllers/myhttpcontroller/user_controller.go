package myhttpcontroller

import (
	"context"

	"github.com/google/uuid"
	"github.com/krzysztofkolcz/my-http-server/internal/api/myhttpserver"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

// List users
// (GET /users)
func (c *APIController) GetUsers(ctx context.Context,
	request myhttpserver.GetUsersRequestObject) (myhttpserver.GetUsersResponseObject, error) {
	parsed, _ := uuid.Parse("7d4ef8d4-9073-483f-a2d5-bc55ab4c9faa")
	email := "some_email@some.com"
	name := "some name"
	user := myhttpserver.User{
		Id:    openapi_types.UUID(parsed),
		Email: openapi_types.Email(email),
		Name:  &name,
	}
	resp := myhttpserver.GetUsers200JSONResponse{
		Items: []myhttpserver.User{
			user,
		},
		Total: 1,
	}

	return resp, nil

}

// Create user
// (POST /users)
func (c *APIController) PostUsers(ctx context.Context,
	request myhttpserver.PostUsersRequestObject) (myhttpserver.PostUsersResponseObject, error) {
	parsed, _ := uuid.Parse("7d4ef8d4-9073-483f-a2d5-bc55ab4c9faa")
	email := "some_email@some.com"
	name := "some name"
	user := myhttpserver.User{
		Id:    openapi_types.UUID(parsed),
		Email: openapi_types.Email(email),
		Name:  &name,
	}
	resp := myhttpserver.PostUsers201JSONResponse(user)
	return resp, nil
}

// Delete user
// (DELETE /users/{id})
func (c *APIController) DeleteUsersId(ctx context.Context,
	request myhttpserver.DeleteUsersIdRequestObject) (myhttpserver.DeleteUsersIdResponseObject, error) {
	resp := myhttpserver.DeleteUsersId204Response{}
	return resp, nil

}

// Get user by ID
// (GET /users/{id})
func (c *APIController) GetUsersId(ctx context.Context,
	request myhttpserver.GetUsersIdRequestObject) (myhttpserver.GetUsersIdResponseObject, error) {
	parsed, _ := uuid.Parse("7d4ef8d4-9073-483f-a2d5-bc55ab4c9faa")
	email := "some_email@some.com"
	name := "some name"
	user := myhttpserver.User{
		Id:    openapi_types.UUID(parsed),
		Email: openapi_types.Email(email),
		Name:  &name,
	}
	resp := myhttpserver.GetUsersId200JSONResponse(user)
	return resp, nil
}

// Update user
// (PATCH /users/{id})
func (c *APIController) PatchUsersId(ctx context.Context,
	request myhttpserver.PatchUsersIdRequestObject) (myhttpserver.PatchUsersIdResponseObject, error) {
	parsed, _ := uuid.Parse("7d4ef8d4-9073-483f-a2d5-bc55ab4c9faa")
	email := "some_email@some.com"
	name := "some name"
	user := myhttpserver.User{
		Id:    openapi_types.UUID(parsed),
		Email: openapi_types.Email(email),
		Name:  &name,
	}
	resp := myhttpserver.PatchUsersId200JSONResponse(user)
	return resp, nil
}
