package apierrors

import (
	"net/http"

	"github.com/krzysztofkolcz/my-http-server/internal/api/myhttpserver"
	"github.com/krzysztofkolcz/my-http-server/utils"
)

var ErrNoClientData = myhttpserver.DetailedError{
	Code:    "NO_CLIENT_DATA",
	Message: "Missing client data",
	Status:  http.StatusInternalServerError,
}

var userinfo = []APIErrors{
	{
		Errors:       []error{utils.ErrExtractClientData},
		ExposedError: ErrNoClientData,
	},
}
