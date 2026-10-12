package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Client is a minimal HTTP client for the Unity Catalog control plane API.
// When a token is configured, requests are authenticated with an
// "Authorization: Bearer <token>" header. Otherwise, when a username and
// password are configured, HTTP Basic authentication is used. When neither is
// configured (e.g. the Unity Catalog server has authentication disabled), no
// authentication is applied. The token takes precedence over basic auth when
// both are supplied.
type Client struct {
	httpClient *http.Client
	host       string
	username   string
	password   string
	token      string
}

// New constructs a new Unity Catalog client. host is the base URL of the
// Unity Catalog control plane (e.g. https://unity.example.com). token, when
// non-empty, enables Bearer authentication and takes precedence over the
// username/password pair. username and password enable HTTP Basic
// authentication when token is empty. When all of token, username and
// password are empty, no authentication is applied (Unity Catalog may be
// configured with authentication disabled).
func New(host, username, password, token string) *Client {
	return &Client{
		httpClient: &http.Client{Timeout: 60 * time.Second},
		host:       strings.TrimRight(host, "/"),
		username:   username,
		password:   password,
		token:      token,
	}
}

// Host returns the base URL the client was configured with.
func (c *Client) Host() string {
	return c.host
}

// setHeaders applies the content-type/accept headers and the configured
// authentication to the request. Token-based Bearer auth takes precedence
// over basic auth. When neither is configured, Unity Catalog may be running
// with authentication disabled.
func (c *Client) setHeaders(req *http.Request, contentType string) {
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("Accept", contentType)
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	} else if c.username != "" || c.password != "" {
		req.SetBasicAuth(c.username, c.password)
	}
}

// do builds and sends an HTTP request with the SCIM2 content type. It is a
// convenience wrapper around doWithContentType for SCIM2 endpoints.
func (c *Client) do(ctx context.Context, method, urlPath string, body interface{}) (*http.Response, error) {
	return c.doWithContentType(ctx, method, urlPath, body, "application/scim+json")
}

// doWithContentType builds and sends an HTTP request with the given method, URL,
// optional JSON body, and explicit content type. Non-SCIM2 Unity Catalog API
// endpoints (e.g. permissions) use "application/json".
func (c *Client) doWithContentType(ctx context.Context, method, urlPath string, body interface{}, contentType string) (*http.Response, error) {
	var reader io.Reader
	if body != nil {
		buf, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		reader = bytes.NewReader(buf)
	}

	req, err := http.NewRequestWithContext(ctx, method, urlPath, reader)
	if err != nil {
		return nil, fmt.Errorf("failed to build request: %w", err)
	}
	c.setHeaders(req, contentType)

	return c.httpClient.Do(req)
}

// readBody reads and closes the response body, returning its bytes.
func readBody(resp *http.Response) ([]byte, error) {
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}

// readResponse reads the response body, then classifies the result against the
// expected status code. It returns:
//   - (body, nil) when the status code matches expected,
//   - (nil, error) for any other status (including 404 Not Found), with the
//     body included in the error.
//
// expected is the success status code for the operation (e.g. http.StatusOK,
// http.StatusCreated). Callers that need to treat 404 as success (e.g. for
// idempotent delete) should check resp.StatusCode before calling readResponse.
func readResponse(resp *http.Response, expected int, op string) ([]byte, error) {
	body, err := readBody(resp)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s response body: %w", op, err)
	}
	if resp.StatusCode == expected {
		return body, nil
	}
	return nil, fmt.Errorf("%s failed: status %d: %s", op, resp.StatusCode, string(body))
}

// decodeOrEmpty decodes body into target. An empty body leaves target at its
// zero value, which is useful for endpoints that may respond with 204 No
// Content.
func decodeOrEmpty(body []byte, target interface{}) error {
	if len(bytes.TrimSpace(body)) == 0 {
		return nil
	}
	if err := json.Unmarshal(body, target); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}
	return nil
}
