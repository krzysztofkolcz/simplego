package apierrors

import (
	"net/http"

	todoapi "github.com/C5383717/my-todo/internal/api/todo"
)

const (
	InternalServerErr = "INTERNAL_SERVER_ERROR"
	JSONDecodeErr     = "JSON_DECODE_ERROR"
	ValidationErr     = "VALIDATION_ERROR"
	ParamsErr         = "PARAMS_ERROR"
	RequiredHeaderErr = "REQUIRED_HEADER_ERROR"
	RequiredParamErr  = "REQUIRED_PARAM_ERROR"
	ForbiddenErr      = "FORBIDDEN"
)

func InternalServerErrorMessage() todoapi.ErrorMessage {
	return todoapi.ErrorMessage{Error: todoapi.DetailedError{
		Code:    InternalServerErr,
		Message: "Internal server error",
		Status:  http.StatusInternalServerError,
	}}
}

func JSONDecodeErrorMessage() todoapi.ErrorMessage {
	return todoapi.ErrorMessage{Error: todoapi.DetailedError{
		Code:    JSONDecodeErr,
		Message: "Can't decode JSON body",
		Status:  http.StatusBadRequest,
	}}
}

func OAPIValidatorErrorMessage(message string, code int) todoapi.ErrorMessage {
	switch code {
	case http.StatusBadRequest:
		return todoapi.ErrorMessage{Error: todoapi.DetailedError{
			Code:    ValidationErr,
			Message: message,
			Status:  code,
		}}
	case http.StatusForbidden:
		return todoapi.ErrorMessage{Error: todoapi.DetailedError{
			Code:    ForbiddenErr,
			Message: message,
			Status:  code,
		}}
	}

	return InternalServerErrorMessage()
}

func TooManyParameters(message string) todoapi.ErrorMessage {
	return todoapi.ErrorMessage{Error: todoapi.DetailedError{
		Code:    ParamsErr,
		Message: message,
		Status:  http.StatusBadRequest,
	}}
}

func RequiredHeaderError(message string) todoapi.ErrorMessage {
	return todoapi.ErrorMessage{Error: todoapi.DetailedError{
		Code:    RequiredHeaderErr,
		Message: message,
		Status:  http.StatusBadRequest,
	}}
}

func RequiredParamError(message string) todoapi.ErrorMessage {
	return todoapi.ErrorMessage{Error: todoapi.DetailedError{
		Code:    RequiredParamErr,
		Message: message,
		Status:  http.StatusBadRequest,
	}}
}
