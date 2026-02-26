package write_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/krzysztofkolcz/my-http-server/internal/api/myhttpserver"
	"github.com/krzysztofkolcz/my-http-server/internal/api/write"
	"github.com/stretchr/testify/assert"
)

func TestWriteErrorResponse(t *testing.T) {
	t.Run("should write error", func(t *testing.T) {
		ctx := myhttpserver.InjectRequestID(t.Context())
		w := httptest.NewRecorder()
		errorResponse := myhttpserver.ErrorMessage{
			Error: myhttpserver.DetailedError{
				Code:    "TEST_ERROR",
				Message: "This is a test error",
				Status:  http.StatusBadRequest,
			},
		}

		write.ErrorResponse(ctx, w, errorResponse)

		requestID, _ := myhttpserver.GetRequestID(ctx)

		var errorMessage myhttpserver.ErrorMessage
		err := json.Unmarshal(w.Body.Bytes(), &errorMessage)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Equal(t, requestID, *errorMessage.Error.RequestID)
	})
}
