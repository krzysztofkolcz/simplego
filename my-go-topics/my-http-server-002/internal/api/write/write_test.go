package write_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/krzysztofkolcz/my-http-server-002/internal/api/myhttpserver2"
	"github.com/krzysztofkolcz/my-http-server-002/internal/api/write"
	mycontext "github.com/krzysztofkolcz/my-http-server-002/utils"
	"github.com/stretchr/testify/assert"
)

func TestWriteErrorResponse(t *testing.T) {
	t.Run("should write error", func(t *testing.T) {
		ctx := mycontext.InjectRequestID(t.Context())
		w := httptest.NewRecorder()
		errorResponse := myhttpserver2.ErrorMessage{
			Error: myhttpserver2.DetailedError{
				Code:    "TEST_ERROR",
				Message: "This is a test error",
				Status:  http.StatusBadRequest,
			},
		}

		write.ErrorResponse(ctx, w, errorResponse)

		requestID, _ := mycontext.GetRequestID(ctx)

		var errorMessage myhttpserver2.ErrorMessage
		err := json.Unmarshal(w.Body.Bytes(), &errorMessage)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Equal(t, requestID, *errorMessage.Error.RequestID)
	})
}
