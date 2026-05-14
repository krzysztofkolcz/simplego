package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	appquery "github.com/krzysztofkolcz/mymigrations/internal/application/query"
	"github.com/krzysztofkolcz/mymigrations/internal/domain"
	"github.com/krzysztofkolcz/mymigrations/internal/http/router"
)

const tenantHeader = "00000000-0000-0000-0000-000000000001"

func todoSrv(s *Server) *httptest.Server   { return httptest.NewServer(router.New(s)) }
func todoPath(ts *httptest.Server, p string) string { return ts.URL + "/v1" + p }

func TestCreateTodo_Returns201(t *testing.T) {
	id := uuid.New()
	ts := todoSrv(&Server{createTodo: &fakeCreateTodo{id: id}})
	defer ts.Close()

	body := bytes.NewBufferString(`{"title":"buy milk"}`)
	req, _ := http.NewRequest(http.MethodPost, todoPath(ts, "/todos"), body)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Tenant-ID", tenantHeader)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	var got map[string]any
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&got))
	assert.Equal(t, id.String(), got["id"])
}

func TestCreateTodo_InvalidTitle_Returns400(t *testing.T) {
	ts := todoSrv(&Server{createTodo: &fakeCreateTodo{err: domain.ErrInvalidTitle}})
	defer ts.Close()

	body := bytes.NewBufferString(`{"title":""}`)
	req, _ := http.NewRequest(http.MethodPost, todoPath(ts, "/todos"), body)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Tenant-ID", tenantHeader)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestCreateTodo_InternalError_Returns500(t *testing.T) {
	ts := todoSrv(&Server{createTodo: &fakeCreateTodo{err: errors.New("db down")}})
	defer ts.Close()

	body := bytes.NewBufferString(`{"title":"buy milk"}`)
	req, _ := http.NewRequest(http.MethodPost, todoPath(ts, "/todos"), body)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Tenant-ID", tenantHeader)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
}

func TestGetTodo_Returns200(t *testing.T) {
	id := uuid.New()
	result := &appquery.GetTodoResult{ID: id, Title: "buy milk", Completed: false, CreatedAt: time.Now()}
	ts := todoSrv(&Server{getTodo: &fakeGetTodo{result: result}})
	defer ts.Close()

	req, _ := http.NewRequest(http.MethodGet, todoPath(ts, "/todos/"+id.String()), nil)
	req.Header.Set("X-Tenant-ID", tenantHeader)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	var got map[string]any
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&got))
	assert.Equal(t, id.String(), got["id"])
	assert.Equal(t, "buy milk", got["title"])
}

func TestGetTodo_NotFound_Returns404(t *testing.T) {
	ts := todoSrv(&Server{getTodo: &fakeGetTodo{err: domain.ErrNotFound}})
	defer ts.Close()

	req, _ := http.NewRequest(http.MethodGet, todoPath(ts, "/todos/"+uuid.New().String()), nil)
	req.Header.Set("X-Tenant-ID", tenantHeader)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestGetTodo_InternalError_Returns500(t *testing.T) {
	ts := todoSrv(&Server{getTodo: &fakeGetTodo{err: errors.New("db down")}})
	defer ts.Close()

	req, _ := http.NewRequest(http.MethodGet, todoPath(ts, "/todos/"+uuid.New().String()), nil)
	req.Header.Set("X-Tenant-ID", tenantHeader)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
}

func TestCompleteTodo_Returns204(t *testing.T) {
	ts := todoSrv(&Server{completeTodo: &fakeCompleteTodo{}})
	defer ts.Close()

	req, _ := http.NewRequest(http.MethodPut, todoPath(ts, "/todos/"+uuid.New().String()+"/complete"), nil)
	req.Header.Set("X-Tenant-ID", tenantHeader)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
}

func TestCompleteTodo_NotFound_Returns404(t *testing.T) {
	ts := todoSrv(&Server{completeTodo: &fakeCompleteTodo{err: domain.ErrNotFound}})
	defer ts.Close()

	req, _ := http.NewRequest(http.MethodPut, todoPath(ts, "/todos/"+uuid.New().String()+"/complete"), nil)
	req.Header.Set("X-Tenant-ID", tenantHeader)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestCompleteTodo_AlreadyCompleted_Returns409(t *testing.T) {
	ts := todoSrv(&Server{completeTodo: &fakeCompleteTodo{err: domain.ErrAlreadyCompleted}})
	defer ts.Close()

	req, _ := http.NewRequest(http.MethodPut, todoPath(ts, "/todos/"+uuid.New().String()+"/complete"), nil)
	req.Header.Set("X-Tenant-ID", tenantHeader)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusConflict, resp.StatusCode)
}

func TestCompleteTodo_InternalError_Returns500(t *testing.T) {
	ts := todoSrv(&Server{completeTodo: &fakeCompleteTodo{err: errors.New("db down")}})
	defer ts.Close()

	req, _ := http.NewRequest(http.MethodPut, todoPath(ts, "/todos/"+uuid.New().String()+"/complete"), nil)
	req.Header.Set("X-Tenant-ID", tenantHeader)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
}

func TestDeleteTodo_Returns204(t *testing.T) {
	ts := todoSrv(&Server{deleteTodo: &fakeDeleteTodo{}})
	defer ts.Close()

	req, _ := http.NewRequest(http.MethodDelete, todoPath(ts, "/todos/"+uuid.New().String()), nil)
	req.Header.Set("X-Tenant-ID", tenantHeader)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
}

func TestDeleteTodo_NotFound_Returns404(t *testing.T) {
	ts := todoSrv(&Server{deleteTodo: &fakeDeleteTodo{err: domain.ErrNotFound}})
	defer ts.Close()

	req, _ := http.NewRequest(http.MethodDelete, todoPath(ts, "/todos/"+uuid.New().String()), nil)
	req.Header.Set("X-Tenant-ID", tenantHeader)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestListTodos_Returns200WithItems(t *testing.T) {
	id := uuid.New()
	items := []appquery.TodoResult{{ID: id, Title: "buy milk", Completed: false, CreatedAt: time.Now()}}
	ts := todoSrv(&Server{listTodos: &fakeListTodos{items: items}})
	defer ts.Close()

	req, _ := http.NewRequest(http.MethodGet, todoPath(ts, "/todos"), nil)
	req.Header.Set("X-Tenant-ID", tenantHeader)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	var got []map[string]any
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&got))
	require.Len(t, got, 1)
	assert.Equal(t, "buy milk", got[0]["title"])
}

func TestListTodos_InternalError_Returns500(t *testing.T) {
	ts := todoSrv(&Server{listTodos: &fakeListTodos{err: errors.New("db down")}})
	defer ts.Close()

	req, _ := http.NewRequest(http.MethodGet, todoPath(ts, "/todos"), nil)
	req.Header.Set("X-Tenant-ID", tenantHeader)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
}
