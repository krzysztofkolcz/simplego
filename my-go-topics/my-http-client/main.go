package myhttpclient

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"time"
)

const (
	statusFinished = "finished"
	statusProgress = "in progress"
	statusError    = "error"

	maxBodySize = 10 << 20 // 10MB
)

var ErrQueryFailed = errors.New("query failed")

// HTTPClient is a minimal interface to allow injecting custom http.Client for tests.
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// ClientOption allows configuring the Client created by NewClient.
type ClientOption func(*Client)

// WithHTTPClient allows injecting a custom http.Client (useful in tests).
func WithHTTPClient(h HTTPClient) ClientOption {
	return func(c *Client) { c.httpClient = h }
}

// WithTLSConfig allows providing a custom TLS config; if provided it will override
// the TLS config built from certificate files. It will try to set TLS config on
// the underlying Transport when possible.
func WithTLSConfig(tlsCfg *tls.Config) ClientOption {
	return func(c *Client) {
		// If httpClient is the default *http.Client with *http.Transport, set its TLS config.
		if hc, ok := c.httpClient.(*http.Client); ok {
			if tr, ok2 := hc.Transport.(*http.Transport); ok2 {
				tr.TLSClientConfig = tlsCfg
				return
			}
			// if Transport is nil or not *http.Transport, replace with a new one carrying tlsCfg
			hc.Transport = &http.Transport{TLSClientConfig: tlsCfg}
			return
		}
		// otherwise, nothing we can do for generic HTTPClient
	}
}

// WithBackoff configures the initial and maximum backoff used while polling query status.
func WithBackoff(initial, max time.Duration) ClientOption {
	return func(c *Client) {
		if initial > 0 {
			c.backoffBase = initial
		}
		if max > 0 {
			c.backoffMax = max
		}
	}
}

// WithHTTPTimeout configures the http.Client Timeout value (applies only if
// the default http.Client was created by NewClient).
func WithHTTPTimeout(timeout time.Duration) ClientOption {
	return func(c *Client) {
		if hc, ok := c.httpClient.(*http.Client); ok {
			hc.Timeout = timeout
		}
	}
}

type Client struct {
	baseURL    string
	httpClient HTTPClient

	// backoff settings used during polling
	backoffBase time.Duration
	backoffMax  time.Duration

	rnd *rand.Rand
}

// NewClient creates a new ALS client. It still accepts client certificate files
// for mTLS by default but also supports functional options to override behavior.
func NewClient(baseURL, clientCertFile, clientKeyFile string, opts ...ClientOption) (*Client, error) {
	cert, err := tls.LoadX509KeyPair(clientCertFile, clientKeyFile)
	if err != nil {
		return nil, fmt.Errorf("load client cert/key: %w", err)
	}

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		MinVersion:   tls.VersionTLS12,
	}

	tr := &http.Transport{
		TLSClientConfig:     tlsConfig,
		IdleConnTimeout:     30 * time.Second,
		TLSHandshakeTimeout: 10 * time.Second,
		MaxIdleConnsPerHost: 10,
	}

	httpClient := &http.Client{
		Transport: tr,
		//Timeout:   30 * time.Second,
	}

	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))

	c := &Client{
		baseURL:     baseURL,
		httpClient:  httpClient,
		backoffBase: time.Second,
		backoffMax:  10 * time.Second,
		rnd:         rnd,
	}

	for _, o := range opts {
		o(c)
	}

	return c, nil
}

func (c *Client) ExecuteQuery(ctx context.Context, req RetrievalRequest) ([]ALSResponseItem, error) {
	queryID, err := c.createQuery(ctx, req)
	if err != nil {
		return nil, err
	}

	if err := c.waitForQueryFinished(ctx, queryID); err != nil {
		return nil, err
	}

	return c.fetchRetrievalQueryResults(ctx, queryID)
}

func (c *Client) createQuery(ctx context.Context, retrievalRequest RetrievalRequest) (string, error) {
	body, err := json.Marshal(retrievalRequest)
	if err != nil {
		return "", fmt.Errorf("marshal retrieval request: %w", err)
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		c.baseURL,
		bytes.NewReader(body),
	)
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("create query request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read create query response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("create query returned status %d: %s", resp.StatusCode, string(respBody))
	}

	var r QueryResponse
	if err := json.Unmarshal(respBody, &r); err != nil {
		return "", fmt.Errorf("unmarshal create-query response: %w; raw: %s", err, string(respBody))
	}

	if r.ID == "" {
		return "", fmt.Errorf("create-query: empty id in response; raw: %s", string(respBody))
	}

	return r.ID, nil
}

func (c *Client) waitForQueryFinished(ctx context.Context, id string) error {
	statusURL := fmt.Sprintf("%s/%s/status", c.baseURL, id)

	backoff := c.backoffBase
	maxBackoff := c.backoffMax

	for {
		status, err := c.getQueryStatus(ctx, statusURL)
		if err != nil {
			return err
		}

		// what about statusPorgress?
		switch status {
		case statusFinished:
			return nil
		case statusError:
			return ErrQueryFailed
		}

		// add jitter up to 50% of backoff
		jitter := time.Duration(c.rnd.Int63n(int64(backoff/2) + 1))
		d := backoff + jitter

		timer := time.NewTimer(d)
		select {
		case <-ctx.Done():
			if !timer.Stop() {
				<-timer.C
			}
			return ctx.Err()
		case <-timer.C:
		}
		timer.Stop()

		if backoff < maxBackoff {
			backoff *= 2
			if backoff > maxBackoff {
				backoff = maxBackoff
			}
		}
	}
}

func (c *Client) getQueryStatus(ctx context.Context, url string) (string, error) {
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		url,
		nil,
	)
	if err != nil {
		return "", err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("get status request failed: %w", err)
	}
	defer resp.Body.Close()

	// readBodyLimited
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read status response: %w", err)
	}

	// checkStatus
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("get status returned %d: %s", resp.StatusCode, string(body))
	}

	var r StatusResponse
	if err := json.Unmarshal(body, &r); err != nil {
		return "", fmt.Errorf("unmarshal status response: %w; raw: %s", err, string(body))
	}

	return r.Status, nil
}

func (c *Client) fetchRetrievalQueryResults(ctx context.Context, id string) ([]ALSResponseItem, error) {
	url := fmt.Sprintf("%s/%s/results", c.baseURL, id)

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		url,
		nil,
	)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do fetch results: %w", err)
	}
	defer resp.Body.Close()

	//if resp.StatusCode < 200 || resp.StatusCode >= 300 {
	//	body, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	//	return nil, fmt.Errorf("fetch results returned %d: %s", resp.StatusCode, string(body))
	//}
	if err := checkStatus(resp, nil); err != nil {
		body, _ := readBodyLimited(resp.Body)
		return nil, fmt.Errorf("%w: %s", err, string(body))
	}

	var items []ALSResponseItem
	if err := json.NewDecoder(
		io.LimitReader(resp.Body, maxBodySize),
	).Decode(&items); err != nil {
		return nil, fmt.Errorf("decode results: %w", err)
	}
	return items, nil
}

func readBodyLimited(r io.Reader) ([]byte, error) {

	body, err := io.ReadAll(
		io.LimitReader(r, maxBodySize),
	)
	if err != nil {
		return nil, fmt.Errorf("read body: %w", err)
	}

	return body, nil
}

func checkStatus(resp *http.Response, body []byte) error {

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return nil
	}

	if body != nil {
		return fmt.Errorf(
			"http %d: %s",
			resp.StatusCode,
			string(body),
		)
	}

	return fmt.Errorf(
		"http %d",
		resp.StatusCode,
	)
}
