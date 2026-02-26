package write

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/krzysztofkolcz/my-http-server/internal/api/myhttpserver"
	"github.com/krzysztofkolcz/my-http-server/internal/log"
	"github.com/krzysztofkolcz/my-http-server/utils"
)

// ErrorResponse writes an error response to the client and logs the error
func ErrorResponse(ctx context.Context, w http.ResponseWriter, errorResponse myhttpserver.ErrorMessage) {
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
