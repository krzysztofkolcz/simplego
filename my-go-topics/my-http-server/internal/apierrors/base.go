package apierrors

import (
	"net/http"

	"github.com/krzysztofkolcz/my-http-server/internal/api/myhttpserver"
)

const (
	InternalServerErr = "INTERNAL_SERVER_ERROR"
	JSONDecodeErr     = "JSON_DECODE_ERROR"
	ValidationErr     = "VALIDATION_ERROR"
	UnauthorizedErr   = "UNAUTHORIZED"
	ParamsErr         = "PARAMS_ERROR"
	RequiredHeaderErr = "REQUIRED_HEADER_ERROR"
	RequiredParamErr  = "REQUIRED_PARAM_ERROR"
	ForbiddenErr      = "FORBIDDEN"
)

func InternalServerErrorMessage() myhttpserver.ErrorMessage {
	return myhttpserver.ErrorMessage{Error: myhttpserver.DetailedError{
		Code:    InternalServerErr,
		Message: "Internal server error",
		Status:  http.StatusInternalServerError,
	}}
}

func JSONDecodeErrorMessage() myhttpserver.ErrorMessage {
	return myhttpserver.ErrorMessage{Error: myhttpserver.DetailedError{
		Code:    JSONDecodeErr,
		Message: "Can't decode JSON body",
		Status:  http.StatusBadRequest,
	}}
}

func OAPIValidatorErrorMessage(message string, code int) myhttpserver.ErrorMessage {
	switch code {
	case http.StatusBadRequest:
		return myhttpserver.ErrorMessage{Error: myhttpserver.DetailedError{
			Code:    ValidationErr,
			Message: message,
			Status:  code,
		}}
	case http.StatusForbidden:
		return myhttpserver.ErrorMessage{Error: myhttpserver.DetailedError{
			Code:    ForbiddenErr,
			Message: message,
			Status:  code,
		}}
	}

	return InternalServerErrorMessage()
}

func TooManyParameters(message string) myhttpserver.ErrorMessage {
	return myhttpserver.ErrorMessage{Error: myhttpserver.DetailedError{
		Code:    ParamsErr,
		Message: message,
		Status:  http.StatusBadRequest,
	}}
}

func RequiredHeaderError(message string) myhttpserver.ErrorMessage {
	return myhttpserver.ErrorMessage{Error: myhttpserver.DetailedError{
		Code:    RequiredHeaderErr,
		Message: message,
		Status:  http.StatusBadRequest,
	}}
}

func RequiredParamError(message string) myhttpserver.ErrorMessage {
	return myhttpserver.ErrorMessage{Error: myhttpserver.DetailedError{
		Code:    RequiredParamErr,
		Message: message,
		Status:  http.StatusBadRequest,
	}}
}
