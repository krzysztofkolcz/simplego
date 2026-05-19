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

func catalogURL(base, path string) string {
	return fmt.Sprintf("%s/v1%s", base, path)
}

// ── Components ─────────────────────────────────────────────────────────────

func TestHTTP_CreateComponent(t *testing.T) {
	ts := newTestServer()
	defer ts.Close()

	body := bytes.NewBufferString(`{"name":"Capacitor 100nF","code":"CAP-100NF"}`)
	req, _ := http.NewRequest(http.MethodPost, catalogURL(ts.URL, "/components"), body)
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

func TestHTTP_CreateComponent_AllFields(t *testing.T) {
	ts := newTestServer()
	defer ts.Close()

	body := bytes.NewBufferString(`{
		"name":"LED Red 5mm",
		"code":"LED-R5",
		"description":"5mm red LED",
		"manufacturer":"Kingbright",
		"manufacturer_url":"https://kingbright.com",
		"price":0.12,
		"weight_g":1,
		"quantity":500,
		"image_url":"https://example.com/led.jpg"
	}`)
	req, _ := http.NewRequest(http.MethodPost, catalogURL(ts.URL, "/components"), body)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Tenant-ID", TenantId)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusCreated, resp.StatusCode)

	var created map[string]any
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&created))
	id := created["id"].(string)

	// get and verify all fields round-trip
	req, _ = http.NewRequest(http.MethodGet, catalogURL(ts.URL, "/components/"+id), nil)
	req.Header.Set("X-Tenant-ID", TenantId)

	resp, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	var got map[string]any
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&got))
	assert.Equal(t, id, got["id"])
	assert.Equal(t, "LED Red 5mm", got["name"])
	assert.Equal(t, "LED-R5", got["code"])
	assert.Equal(t, "5mm red LED", got["description"])
	assert.Equal(t, "Kingbright", got["manufacturer"])
	assert.InDelta(t, 0.12, got["price"].(float64), 0.001)
	assert.Equal(t, float64(1), got["weight_g"])
	assert.Equal(t, float64(500), got["quantity"])
	assert.NotEmpty(t, got["created_at"])
}

func TestHTTP_GetComponent_NotFound(t *testing.T) {
	ts := newTestServer()
	defer ts.Close()

	req, _ := http.NewRequest(http.MethodGet, catalogURL(ts.URL, "/components/00000000-0000-0000-0000-000000000001"), nil)
	req.Header.Set("X-Tenant-ID", TenantId)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestHTTP_CreateComponent_InvalidName(t *testing.T) {
	ts := newTestServer()
	defer ts.Close()

	body := bytes.NewBufferString(`{"name":"","code":"RES-1K"}`)
	req, _ := http.NewRequest(http.MethodPost, catalogURL(ts.URL, "/components"), body)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Tenant-ID", TenantId)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestHTTP_CreateComponent_InvalidCode(t *testing.T) {
	ts := newTestServer()
	defer ts.Close()

	body := bytes.NewBufferString(`{"name":"Resistor","code":""}`)
	req, _ := http.NewRequest(http.MethodPost, catalogURL(ts.URL, "/components"), body)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Tenant-ID", TenantId)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

// ── Products ───────────────────────────────────────────────────────────────

func TestHTTP_CreateProduct(t *testing.T) {
	ts := newTestServer()
	defer ts.Close()

	body := bytes.NewBufferString(`{"name":"Basic Kit"}`)
	req, _ := http.NewRequest(http.MethodPost, catalogURL(ts.URL, "/products"), body)
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

func TestHTTP_GetProduct_OK(t *testing.T) {
	ts := newTestServer()
	defer ts.Close()

	body := bytes.NewBufferString(`{
		"name":"Sensor Kit",
		"description":"Kit with sensors",
		"price":49.99,
		"tags":["sensor","kit"]
	}`)
	req, _ := http.NewRequest(http.MethodPost, catalogURL(ts.URL, "/products"), body)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Tenant-ID", TenantId)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	var created map[string]any
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&created))
	id := created["id"].(string)

	req, _ = http.NewRequest(http.MethodGet, catalogURL(ts.URL, "/products/"+id), nil)
	req.Header.Set("X-Tenant-ID", TenantId)

	resp, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	var got map[string]any
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&got))
	assert.Equal(t, id, got["id"])
	assert.Equal(t, "Sensor Kit", got["name"])
	assert.Equal(t, "Kit with sensors", got["description"])
	assert.InDelta(t, 49.99, got["price"].(float64), 0.001)
	assert.ElementsMatch(t, []any{"sensor", "kit"}, got["tags"])
	assert.Empty(t, got["recipe"])
	assert.NotEmpty(t, got["created_at"])
}

func TestHTTP_GetProduct_NotFound(t *testing.T) {
	ts := newTestServer()
	defer ts.Close()

	req, _ := http.NewRequest(http.MethodGet, catalogURL(ts.URL, "/products/00000000-0000-0000-0000-000000000001"), nil)
	req.Header.Set("X-Tenant-ID", TenantId)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestHTTP_CreateProduct_InvalidName(t *testing.T) {
	ts := newTestServer()
	defer ts.Close()

	body := bytes.NewBufferString(`{"name":""}`)
	req, _ := http.NewRequest(http.MethodPost, catalogURL(ts.URL, "/products"), body)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Tenant-ID", TenantId)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestHTTP_CreateComponent_DuplicateCode(t *testing.T) {
	ts := newTestServer()
	defer ts.Close()

	code := fmt.Sprintf("UNIQUE-%s", TenantId[:8])
	body := fmt.Sprintf(`{"name":"Component A","code":%q}`, code)

	for i, expectedStatus := range []int{http.StatusCreated, http.StatusInternalServerError} {
		req, _ := http.NewRequest(http.MethodPost, catalogURL(ts.URL, "/components"), bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Tenant-ID", TenantId)

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		resp.Body.Close()
		assert.Equal(t, expectedStatus, resp.StatusCode, "attempt %d", i+1)
	}
}
