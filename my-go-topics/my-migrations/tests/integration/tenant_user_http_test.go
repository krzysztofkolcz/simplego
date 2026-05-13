package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHTTP_CreateTenant(t *testing.T) {
	ts := newTestServer()
	defer ts.Close()

	body := bytes.NewBufferString(`{"schema_name":"test_tenant_create"}`)
	req, _ := http.NewRequest(http.MethodPost, todoURL(ts.URL, "/tenants"), body)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	var result map[string]any
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&result))
	assert.NotEmpty(t, result["id"])
}

func TestHTTP_CreateTenant_Conflict(t *testing.T) {
	ts := newTestServer()
	defer ts.Close()

	for i, expected := range []int{http.StatusCreated, http.StatusConflict} {
		body := bytes.NewBufferString(`{"schema_name":"conflict_tenant"}`)
		req, _ := http.NewRequest(http.MethodPost, todoURL(ts.URL, "/tenants"), body)
		req.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		resp.Body.Close()
		assert.Equal(t, expected, resp.StatusCode, "attempt %d", i+1)
	}
}

func TestHTTP_GetTenant_OK(t *testing.T) {
	ts := newTestServer()
	defer ts.Close()

	// create
	body := bytes.NewBufferString(`{"schema_name":"test_tenant_get"}`)
	req, _ := http.NewRequest(http.MethodPost, todoURL(ts.URL, "/tenants"), body)
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	var created map[string]any
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&created))
	id := created["id"].(string)

	// get
	req, _ = http.NewRequest(http.MethodGet, todoURL(ts.URL, "/tenants/"+id), nil)
	resp, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	var got map[string]any
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&got))
	assert.Equal(t, id, got["id"])
	assert.Equal(t, "test_tenant_get", got["schema_name"])
}

func TestHTTP_GetTenant_NotFound(t *testing.T) {
	ts := newTestServer()
	defer ts.Close()

	req, _ := http.NewRequest(http.MethodGet, todoURL(ts.URL, "/tenants/00000000-0000-0000-0000-000000000001"), nil)
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestHTTP_CreateUser(t *testing.T) {
	ts := newTestServer()
	defer ts.Close()

	body := bytes.NewBufferString(`{"email":"newuser@example.com"}`)
	req, _ := http.NewRequest(http.MethodPost, todoURL(ts.URL, "/users"), body)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	var result map[string]any
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&result))
	assert.NotEmpty(t, result["id"])
}

func TestHTTP_CreateUser_Conflict(t *testing.T) {
	ts := newTestServer()
	defer ts.Close()

	for i, expected := range []int{http.StatusCreated, http.StatusConflict} {
		body := bytes.NewBufferString(`{"email":"conflict@example.com"}`)
		req, _ := http.NewRequest(http.MethodPost, todoURL(ts.URL, "/users"), body)
		req.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		resp.Body.Close()
		assert.Equal(t, expected, resp.StatusCode, fmt.Sprintf("attempt %d", i+1))
	}
}

func TestHTTP_GetUser_OK(t *testing.T) {
	ts := newTestServer()
	defer ts.Close()

	// create
	body := bytes.NewBufferString(`{"email":"getuser@example.com"}`)
	req, _ := http.NewRequest(http.MethodPost, todoURL(ts.URL, "/users"), body)
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	var created map[string]any
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&created))
	id := created["id"].(string)

	// get
	req, _ = http.NewRequest(http.MethodGet, todoURL(ts.URL, "/users/"+id), nil)
	resp, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	var got map[string]any
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&got))
	assert.Equal(t, id, got["id"])
	assert.Equal(t, "getuser@example.com", got["email"])
}

func TestHTTP_GetUser_NotFound(t *testing.T) {
	ts := newTestServer()
	defer ts.Close()

	req, _ := http.NewRequest(http.MethodGet, todoURL(ts.URL, "/users/00000000-0000-0000-0000-000000000001"), nil)
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}
