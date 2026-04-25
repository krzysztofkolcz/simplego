package write

import (
	"context"
	"encoding/json"
	"net/http"

	todoapi "github.com/C5383717/my-todo/internal/api/todo"
	"github.com/C5383717/my-todo/internal/log"
	"github.com/C5383717/my-todo/utils"
)

// ErrorResponse writes a structured JSON error response and injects the RequestID.
func ErrorResponse(ctx context.Context, w http.ResponseWriter, errorResponse todoapi.ErrorMessage) {
	requestID, _ := utils.GetRequestID(ctx)

	errorResponse.Error.RequestID = &requestID

	w.WriteHeader(errorResponse.Error.Status)

	enc := json.NewEncoder(w)

	err := enc.Encode(&errorResponse)
	if err != nil {
		log.Error(ctx, "Failed to encode error response", err)
		http.Error(w, "Failed to encode error response", http.StatusInternalServerError)

		return
	}
}
