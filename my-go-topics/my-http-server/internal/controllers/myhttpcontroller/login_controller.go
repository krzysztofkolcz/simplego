package myhttpcontroller

import (
	"context"

	"github.com/krzysztofkolcz/my-http-server/internal/api/myhttpserver"
)

// Login
// (POST /auth/login)
func (c *APIController) PostAuthLogin(ctx context.Context,
	request myhttpserver.PostAuthLoginRequestObject) (myhttpserver.PostAuthLoginResponseObject, error) {
	return myhttpserver.PostAuthLogin200Response{}, nil

}
