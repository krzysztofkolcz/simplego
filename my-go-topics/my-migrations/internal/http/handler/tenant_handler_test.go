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

func tenantSrv(s *Server) *httptest.Server      { return httptest.NewServer(router.New(s)) }
func tenantPath(ts *httptest.Server, p string) string { return ts.URL + "/v1" + p }

func TestCreateTenant_Returns201(t *testing.T) {
	id := uuid.New()
	ts := tenantSrv(&Server{createTenant: &fakeCreateTenant{id: id}})
	defer ts.Close()

	body := bytes.NewBufferString(`{"schema_name":"acme"}`)
	req, _ := http.NewRequest(http.MethodPost, tenantPath(ts, "/tenants"), body)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	var got map[string]any
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&got))
	assert.Equal(t, id.String(), got["id"])
}

func TestCreateTenant_Conflict_Returns409(t *testing.T) {
	ts := tenantSrv(&Server{createTenant: &fakeCreateTenant{err: domain.ErrConflict}})
	defer ts.Close()

	body := bytes.NewBufferString(`{"schema_name":"acme"}`)
	req, _ := http.NewRequest(http.MethodPost, tenantPath(ts, "/tenants"), body)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusConflict, resp.StatusCode)
}

func TestCreateTenant_InternalError_Returns500(t *testing.T) {
	ts := tenantSrv(&Server{createTenant: &fakeCreateTenant{err: errors.New("db down")}})
	defer ts.Close()

	body := bytes.NewBufferString(`{"schema_name":"acme"}`)
	req, _ := http.NewRequest(http.MethodPost, tenantPath(ts, "/tenants"), body)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
}

func TestGetTenant_Returns200(t *testing.T) {
	id := uuid.New()
	result := &appquery.GetTenantResult{
		ID:               id,
		SchemaName:       "acme",
		MigrationStatus:  "done",
		MigrationError:   "",
		MigrationUpdated: time.Now(),
		CreatedAt:        time.Now(),
	}
	ts := tenantSrv(&Server{getTenant: &fakeGetTenant{result: result}})
	defer ts.Close()

	req, _ := http.NewRequest(http.MethodGet, tenantPath(ts, "/tenants/"+id.String()), nil)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	var got map[string]any
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&got))
	assert.Equal(t, id.String(), got["id"])
	assert.Equal(t, "acme", got["schema_name"])
}

func TestGetTenant_NotFound_Returns404(t *testing.T) {
	ts := tenantSrv(&Server{getTenant: &fakeGetTenant{err: domain.ErrNotFound}})
	defer ts.Close()

	req, _ := http.NewRequest(http.MethodGet, tenantPath(ts, "/tenants/"+uuid.New().String()), nil)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}
