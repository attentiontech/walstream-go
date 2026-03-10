package client

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"

	walstream "github.com/attentiontech/walstream-go"
)

// ErrUnauthorized is returned when the server rejects a request due to
// missing or invalid authentication.
var ErrUnauthorized = errors.New("unauthorized")

// DefaultServerURL is the default walstream server URL.
const DefaultServerURL = "http://localhost:9795"

// Client is a Go client for the walstream API.
type Client struct {
	baseURL     string
	httpClient  *http.Client
	bearerToken string
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
	return c
}

func (c *Client) newRequest(ctx context.Context, method, url string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, err
	}
	if c.bearerToken != "" {
		req.Header.Set("Authorization", "Bearer "+c.bearerToken)
	}
	return req, nil
}

// Apply creates or updates a pipeline. Returns the apply response and whether
// it was newly created (true on 201, false on 200).
func (c *Client) Apply(ctx context.Context, spec walstream.PipelineSpec) (*walstream.ApplyResponse, bool, error) {
	body, err := json.Marshal(spec)
	if err != nil {
		return nil, false, fmt.Errorf("failed to marshal spec: %w", err)
	}

	url := fmt.Sprintf("%s/api/v1/pipelines/%s", c.baseURL, spec.Name)
	req, err := c.newRequest(ctx, http.MethodPut, url, bytes.NewReader(body))
	if err != nil {
		return nil, false, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, false, err
	}
	defer resp.Body.Close()

	if err := checkError(resp); err != nil {
		return nil, false, err
	}

	var result walstream.ApplyResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, false, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, resp.StatusCode == http.StatusCreated, nil
}

// Destroy deletes a pipeline by name.
func (c *Client) Destroy(ctx context.Context, name string) (*walstream.DestroyResponse, error) {
	url := fmt.Sprintf("%s/api/v1/pipelines/%s", c.baseURL, name)
	req, err := c.newRequest(ctx, http.MethodDelete, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if err := checkError(resp); err != nil {
		return nil, err
	}

	var result walstream.DestroyResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

// List returns all pipelines with their current state.
func (c *Client) List(ctx context.Context) ([]walstream.PipelineState, error) {
	url := fmt.Sprintf("%s/api/v1/pipelines", c.baseURL)
	req, err := c.newRequest(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if err := checkError(resp); err != nil {
		return nil, err
	}

	var states []walstream.PipelineState
	if err := json.NewDecoder(resp.Body).Decode(&states); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return states, nil
}

// Get returns a single pipeline's state.
func (c *Client) Get(ctx context.Context, name string) (*walstream.PipelineState, error) {
	url := fmt.Sprintf("%s/api/v1/pipelines/%s", c.baseURL, name)
	req, err := c.newRequest(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if err := checkError(resp); err != nil {
		return nil, err
	}

	var state walstream.PipelineState
	if err := json.NewDecoder(resp.Body).Decode(&state); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &state, nil
}

// Healthz returns the effective status of a pipeline.
func (c *Client) Healthz(ctx context.Context, name string) (walstream.EffectiveStatus, error) {
	url := fmt.Sprintf("%s/api/v1/pipelines/%s/healthz", c.baseURL, name)
	req, err := c.newRequest(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	// 503 is expected for non-running pipelines, not an error
	if resp.StatusCode == http.StatusNotFound {
		return "", fmt.Errorf("pipeline %q: %w", name, walstream.ErrPipelineNotFound)
	}

	return walstream.EffectiveStatus(result.Status), nil
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
			return fmt.Errorf("%s: %w", apiErr.Error, walstream.ErrPipelineNotFound)
		case http.StatusBadRequest:
			return fmt.Errorf("%s: %w", apiErr.Error, walstream.ErrValidation)
		default:
			return fmt.Errorf("server error: %s", apiErr.Error)
		}
	}

	return fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(body))
}
