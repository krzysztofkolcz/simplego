package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/krzysztofkolcz/mymigrations/internal/http/handler"
	"github.com/krzysztofkolcz/mymigrations/internal/http/router"
	"github.com/krzysztofkolcz/mymigrations/internal/infrastructure/db"
	commanddb "github.com/krzysztofkolcz/mymigrations/internal/infrastructure/db/sqlc/command"
	querydb "github.com/krzysztofkolcz/mymigrations/internal/infrastructure/db/sqlc/query"
	infraevent "github.com/krzysztofkolcz/mymigrations/internal/infrastructure/event"
	"github.com/krzysztofkolcz/mymigrations/internal/infrastructure/usecase"
)

func newTestServer() *httptest.Server {
	txManager := db.NewTxManager(ConnectionPool)
	commandQ := commanddb.New(ConnectionPool)
	queryQ := querydb.New(ConnectionPool)
	eventPublisher := infraevent.NewLogPublisher()

	srv := handler.NewServer(
		usecase.NewCreateTodoUseCase(txManager, eventPublisher),
		usecase.NewCompleteTodoUseCase(txManager),
		usecase.NewDeleteTodoUseCase(txManager),
		usecase.NewGetTodoUseCase(txManager),
		usecase.NewListTodosUseCase(txManager),
		usecase.NewCreateTenantUseCase(txManager),
		usecase.NewGetTenantUseCase(commandQ, queryQ),
		usecase.NewCreateUserUseCase(txManager),
		usecase.NewGetUserUseCase(commandQ, queryQ),
	)
	return httptest.NewServer(router.New(srv))
}

func todoURL(base, path string) string {
	return fmt.Sprintf("%s/v1%s", base, path)
}

func TestHTTP_CreateTodo(t *testing.T) {
	ts := newTestServer()
	defer ts.Close()

	body := bytes.NewBufferString(`{"title":"buy groceries"}`)
	req, _ := http.NewRequest(http.MethodPost, todoURL(ts.URL, "/todos"), body)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Tenant-ID", TenantId)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var result map[string]any
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&result))
	assert.NotEmpty(t, result["id"])
}

func TestHTTP_GetTodo_OK(t *testing.T) {
	ts := newTestServer()
	defer ts.Close()

	// create
	body := bytes.NewBufferString(`{"title":"read a book"}`)
	req, _ := http.NewRequest(http.MethodPost, todoURL(ts.URL, "/todos"), body)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Tenant-ID", TenantId)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	var created map[string]any
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&created))
	id := created["id"].(string)

	// get
	req, _ = http.NewRequest(http.MethodGet, todoURL(ts.URL, "/todos/"+id), nil)
	req.Header.Set("X-Tenant-ID", TenantId)

	resp, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var got map[string]any
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&got))
	assert.Equal(t, id, got["id"])
	assert.Equal(t, "read a book", got["title"])
	assert.Equal(t, false, got["completed"])
	assert.NotEmpty(t, got["created_at"])
}

func TestHTTP_GetTodo_NotFound(t *testing.T) {
	ts := newTestServer()
	defer ts.Close()

	req, _ := http.NewRequest(http.MethodGet, todoURL(ts.URL, "/todos/00000000-0000-0000-0000-000000000001"), nil)
	req.Header.Set("X-Tenant-ID", TenantId)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestHTTP_ListTodos(t *testing.T) {
	ts := newTestServer()
	defer ts.Close()

	// create two todos
	for _, title := range []string{"http list todo A", "http list todo B"} {
		body := bytes.NewBufferString(fmt.Sprintf(`{"title":%q}`, title))
		req, _ := http.NewRequest(http.MethodPost, todoURL(ts.URL, "/todos"), body)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Tenant-ID", TenantId)
		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		resp.Body.Close()
		require.Equal(t, http.StatusCreated, resp.StatusCode)
	}

	// list
	req, _ := http.NewRequest(http.MethodGet, todoURL(ts.URL, "/todos"), nil)
	req.Header.Set("X-Tenant-ID", TenantId)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var list []map[string]any
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&list))

	titles := make(map[string]bool, len(list))
	for _, item := range list {
		titles[item["title"].(string)] = true
	}
	assert.True(t, titles["http list todo A"])
	assert.True(t, titles["http list todo B"])
}

func TestHTTP_CompleteTodo(t *testing.T) {
	ts := newTestServer()
	defer ts.Close()

	// create
	body := bytes.NewBufferString(`{"title":"write tests"}`)
	req, _ := http.NewRequest(http.MethodPost, todoURL(ts.URL, "/todos"), body)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Tenant-ID", TenantId)
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	var created map[string]any
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&created))
	id := created["id"].(string)

	// complete
	req, _ = http.NewRequest(http.MethodPut, todoURL(ts.URL, "/todos/"+id+"/complete"), nil)
	req.Header.Set("X-Tenant-ID", TenantId)
	resp, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	resp.Body.Close()
	assert.Equal(t, http.StatusNoContent, resp.StatusCode)

	// verify
	req, _ = http.NewRequest(http.MethodGet, todoURL(ts.URL, "/todos/"+id), nil)
	req.Header.Set("X-Tenant-ID", TenantId)
	resp, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	var got map[string]any
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&got))
	assert.Equal(t, true, got["completed"])
}

func TestHTTP_DeleteTodo(t *testing.T) {
	ts := newTestServer()
	defer ts.Close()

	// create
	body := bytes.NewBufferString(`{"title":"to be deleted"}`)
	req, _ := http.NewRequest(http.MethodPost, todoURL(ts.URL, "/todos"), body)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Tenant-ID", TenantId)
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	var created map[string]any
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&created))
	id := created["id"].(string)

	// delete
	req, _ = http.NewRequest(http.MethodDelete, todoURL(ts.URL, "/todos/"+id), nil)
	req.Header.Set("X-Tenant-ID", TenantId)
	resp, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	resp.Body.Close()
	assert.Equal(t, http.StatusNoContent, resp.StatusCode)

	// verify 404
	req, _ = http.NewRequest(http.MethodGet, todoURL(ts.URL, "/todos/"+id), nil)
	req.Header.Set("X-Tenant-ID", TenantId)
	resp, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}
