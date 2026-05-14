package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	appquery "github.com/krzysztofkolcz/mymigrations/internal/application/query"
	"github.com/krzysztofkolcz/mymigrations/internal/domain"
	"github.com/krzysztofkolcz/mymigrations/internal/http/router"
)

func userSrv(s *Server) *httptest.Server      { return httptest.NewServer(router.New(s)) }
func userPath(ts *httptest.Server, p string) string { return ts.URL + "/v1" + p }

func TestCreateUser_Returns201(t *testing.T) {
	id := uuid.New()
	ts := userSrv(&Server{createUser: &fakeCreateUser{id: id}})
	defer ts.Close()

	body := bytes.NewBufferString(`{"email":"alice@example.com"}`)
	req, _ := http.NewRequest(http.MethodPost, userPath(ts, "/users"), body)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	var got map[string]any
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&got))
	assert.Equal(t, id.String(), got["id"])
}

func TestCreateUser_InvalidEmail_Returns400(t *testing.T) {
	ts := userSrv(&Server{createUser: &fakeCreateUser{err: domain.ErrInvalidEmail}})
	defer ts.Close()

	body := bytes.NewBufferString(`{"email":"not-an-email"}`)
	req, _ := http.NewRequest(http.MethodPost, userPath(ts, "/users"), body)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestCreateUser_Conflict_Returns409(t *testing.T) {
	ts := userSrv(&Server{createUser: &fakeCreateUser{err: domain.ErrConflict}})
	defer ts.Close()

	body := bytes.NewBufferString(`{"email":"alice@example.com"}`)
	req, _ := http.NewRequest(http.MethodPost, userPath(ts, "/users"), body)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusConflict, resp.StatusCode)
}

func TestGetUser_Returns200(t *testing.T) {
	id := uuid.New()
	result := &appquery.GetUserResult{ID: id, Email: "alice@example.com"}
	ts := userSrv(&Server{getUser: &fakeGetUser{result: result}})
	defer ts.Close()

	req, _ := http.NewRequest(http.MethodGet, userPath(ts, "/users/"+id.String()), nil)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	var got map[string]any
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&got))
	assert.Equal(t, id.String(), got["id"])
	assert.Equal(t, "alice@example.com", got["email"])
}

func TestGetUser_NotFound_Returns404(t *testing.T) {
	ts := userSrv(&Server{getUser: &fakeGetUser{err: domain.ErrNotFound}})
	defer ts.Close()

	req, _ := http.NewRequest(http.MethodGet, userPath(ts, "/users/"+uuid.New().String()), nil)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}
