package client

import "github.com/attentiontech/walstream-go/types"

// MessageLevel indicates the severity of an API response message.
type MessageLevel string

const (
	MessageLevelInfo    MessageLevel = "info"
	MessageLevelWarning MessageLevel = "warning"
)

// Message is a structured message returned by the API.
type Message struct {
	Level MessageLevel `json:"level"`
	Text  string       `json:"text"`
}

// Response is the envelope for mutating API responses.
type Response struct {
	Messages []Message `json:"messages,omitempty"`
}

// DestroyResponse is returned by the destroy endpoint.
type DestroyResponse struct {
	Response
	Status string `json:"status"`
}

// Change represents a single detected field difference between two pipeline specs.
type Change struct {
	Field string `json:"field"`
}

// ApplyResponse is returned by the apply endpoint.
type ApplyResponse struct {
	Response
	types.PipelineState
	Changes []Change `json:"changes"`
	Created bool     `json:"-"`
}
