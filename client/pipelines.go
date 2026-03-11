package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/attentiontech/walstream-go/types"
)

// PipelineService handles all pipeline-related API operations.
type PipelineService struct {
	r Requester
}

// NewPipelineService creates a PipelineService backed by the given Requester.
func NewPipelineService(r Requester) *PipelineService {
	return &PipelineService{r: r}
}

// Apply creates or updates a pipeline. The response includes a Created field
// indicating whether the pipeline was newly created (HTTP 201) or updated (HTTP 200).
func (s *PipelineService) Apply(ctx context.Context, spec types.PipelineSpec) (*ApplyResponse, error) {
	body, err := json.Marshal(spec)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal spec: %w", err)
	}

	req, err := s.r.NewRequest(ctx, http.MethodPut, "/api/v1/pipelines/"+spec.Name, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.r.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if err := checkError(resp); err != nil {
		return nil, err
	}

	var result ApplyResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	result.Created = resp.StatusCode == http.StatusCreated
	return &result, nil
}

// Destroy deletes a pipeline by name.
func (s *PipelineService) Destroy(ctx context.Context, name string) (*DestroyResponse, error) {
	req, err := s.r.NewRequest(ctx, http.MethodDelete, "/api/v1/pipelines/"+name, nil)
	if err != nil {
		return nil, err
	}

	resp, err := s.r.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if err := checkError(resp); err != nil {
		return nil, err
	}

	var result DestroyResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

// List returns all pipelines with their current state.
func (s *PipelineService) List(ctx context.Context) ([]types.PipelineState, error) {
	req, err := s.r.NewRequest(ctx, http.MethodGet, "/api/v1/pipelines", nil)
	if err != nil {
		return nil, err
	}

	resp, err := s.r.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if err := checkError(resp); err != nil {
		return nil, err
	}

	var states []types.PipelineState
	if err := json.NewDecoder(resp.Body).Decode(&states); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return states, nil
}

// Get returns a single pipeline's state.
func (s *PipelineService) Get(ctx context.Context, name string) (*types.PipelineState, error) {
	req, err := s.r.NewRequest(ctx, http.MethodGet, "/api/v1/pipelines/"+name, nil)
	if err != nil {
		return nil, err
	}

	resp, err := s.r.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if err := checkError(resp); err != nil {
		return nil, err
	}

	var state types.PipelineState
	if err := json.NewDecoder(resp.Body).Decode(&state); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &state, nil
}

// Healthz returns the effective status of a pipeline.
func (s *PipelineService) Healthz(ctx context.Context, name string) (types.EffectiveStatus, error) {
	req, err := s.r.NewRequest(ctx, http.MethodGet, "/api/v1/pipelines/"+name+"/healthz", nil)
	if err != nil {
		return "", err
	}

	resp, err := s.r.Do(req)
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
		return "", fmt.Errorf("pipeline %q: %w", name, types.ErrPipelineNotFound)
	}

	return types.EffectiveStatus(result.Status), nil
}
