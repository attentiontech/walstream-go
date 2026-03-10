package client

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/attentiontech/walstream-go/types"
)

// ErrUnauthorized is returned when the server rejects a request due to
// missing or invalid authentication.
var ErrUnauthorized = errors.New("unauthorized")

// DefaultServerURL is the default walstream server URL.
const DefaultServerURL = "http://localhost:9795"

// Requester creates and executes HTTP requests pre-configured with
// base URL and authentication.
type Requester interface {
	NewRequest(ctx context.Context, method, path string, body io.Reader) (*http.Request, error)
	Do(req *http.Request) (*http.Response, error)
}

// Client is a Go client for the walstream API.
type Client struct {
	baseURL     string
	httpClient  *http.Client
	bearerToken string

	Pipelines *PipelineService
}

// Option configures a Client.
type Option func(*Client)

// WithServerURL sets the server URL.
func WithServerURL(url string) Option {
	return func(c *Client) {
		c.baseURL = url
	}
}

// WithHTTPClient sets a custom HTTP client.
func WithHTTPClient(hc *http.Client) Option {
	return func(c *Client) {
		c.httpClient = hc
	}
}

// WithBearerToken sets the bearer token for authentication.
func WithBearerToken(token string) Option {
	return func(c *Client) {
		c.bearerToken = token
	}
}

// New creates an API client with the given options.
// Falls back to WALSTREAM_URL and WALSTREAM_TOKEN env vars for defaults.
func New(opts ...Option) *Client {
	baseURL := DefaultServerURL
	if v := os.Getenv("WALSTREAM_URL"); v != "" {
		baseURL = v
	}
	c := &Client{
		baseURL:     baseURL,
		httpClient:  &http.Client{},
		bearerToken: os.Getenv("WALSTREAM_TOKEN"),
	}
	for _, opt := range opts {
		opt(c)
	}
	c.Pipelines = NewPipelineService(c)
	return c
}

// NewRequest creates an HTTP request with the base URL prepended and
// authentication headers set.
func (c *Client) NewRequest(ctx context.Context, method, path string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, body)
	if err != nil {
		return nil, err
	}
	if c.bearerToken != "" {
		req.Header.Set("Authorization", "Bearer "+c.bearerToken)
	}
	return req, nil
}

// Do executes an HTTP request.
func (c *Client) Do(req *http.Request) (*http.Response, error) {
	return c.httpClient.Do(req)
}

func checkError(resp *http.Response) error {
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return nil
	}

	body, _ := io.ReadAll(resp.Body)

	var apiErr struct {
		Error string `json:"error"`
	}
	if resp.StatusCode == http.StatusUnauthorized {
		return fmt.Errorf("unauthorized: set a token via --token flag or WALSTREAM_TOKEN environment variable: %w", ErrUnauthorized)
	}

	if json.Unmarshal(body, &apiErr) == nil && apiErr.Error != "" {
		switch resp.StatusCode {
		case http.StatusNotFound:
			return fmt.Errorf("%s: %w", apiErr.Error, types.ErrPipelineNotFound)
		case http.StatusBadRequest:
			return fmt.Errorf("%s: %w", apiErr.Error, types.ErrValidation)
		default:
			return fmt.Errorf("server error: %s", apiErr.Error)
		}
	}

	return fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(body))
}
