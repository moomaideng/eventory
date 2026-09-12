package apptest

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"time"
)

const requestTimeout = 10 * time.Second

// Client drives the HTTP API in integration tests.
type Client struct {
	http    *http.Client
	baseURL string
}

// NewClient returns a fresh HTTP client bound to baseURL.
func NewClient(baseURL string) *Client {
	return &Client{
		http:    &http.Client{Timeout: requestTimeout},
		baseURL: baseURL,
	}
}

// BaseURL returns the root URL of the running test server.
func (c *Client) BaseURL() string {
	return c.baseURL
}

// Response captures the HTTP response details for scenario assertions.
type Response struct {
	StatusCode int
	Header     http.Header
	Body       []byte
}

// JSON decodes the response body into target struct v.
func (r *Response) JSON(v any) error {
	return json.Unmarshal(r.Body, v)
}

// Do performs an HTTP request against the API server.
func (c *Client) Do(method, path string, body any, headers map[string]string) (*Response, error) {
	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		bodyReader = bytes.NewReader(data)
	}

	req, err := http.NewRequest(method, c.baseURL+path, bodyReader)
	if err != nil {
		return nil, err
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return &Response{
		StatusCode: resp.StatusCode,
		Header:     resp.Header,
		Body:       respBody,
	}, nil
}
