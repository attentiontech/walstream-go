package walstream

import (
	"encoding/json"
	"fmt"
	"time"
)

// Table identifies a table by schema and name.
type Table struct {
	Schema string `json:"schema"`
	Name   string `json:"name"`
}

// QualifiedName returns the schema-qualified table name (e.g. "public.users").
func (t Table) QualifiedName() string {
	return fmt.Sprintf("%s.%s", t.Schema, t.Name)
}

// Duration wraps time.Duration with JSON support (marshals as a string like "1h", "30m").
type Duration time.Duration

func (d Duration) MarshalJSON() ([]byte, error) {
	return json.Marshal(time.Duration(d).String())
}

func (d *Duration) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	dur, err := time.ParseDuration(s)
	if err != nil {
		return err
	}
	*d = Duration(dur)
	return nil
}

// DesiredStatus is the user-requested state of a pipeline.
type DesiredStatus string

const (
	DesiredStatusRunning DesiredStatus = "running"
	DesiredStatusStopped DesiredStatus = "stopped"
)

// CleanupPolicy controls Kafka topic retention behavior.
type CleanupPolicy string

const (
	CleanupPolicyCompact CleanupPolicy = "compact"
	CleanupPolicyDelete  CleanupPolicy = "delete"
)

// EffectiveStatus is the actual runtime state of a pipeline.
type EffectiveStatus string

const (
	EffectiveStatusRunning    EffectiveStatus = "running"
	EffectiveStatusFailing    EffectiveStatus = "failing"
	EffectiveStatusRestarting EffectiveStatus = "restarting"
	EffectiveStatusStopped    EffectiveStatus = "stopped"
)

// SourceConfig holds per-pipeline source (postgres) settings.
type SourceConfig struct {
	Connection string  `json:"connection"`
	Tables     []Table `json:"tables"`
}

// KafkaTopicInitial holds settings applied only when creating a Kafka topic.
type KafkaTopicInitial struct {
	Partitions    int           `json:"partitions"`
	CleanupPolicy CleanupPolicy `json:"cleanup_policy"`
	Retention     *Duration     `json:"retention,omitempty"`
}

// KafkaTopicOverrideInitial holds per-topic creation-time overrides.
type KafkaTopicOverrideInitial struct {
	Partitions    *int           `json:"partitions,omitempty"`
	CleanupPolicy *CleanupPolicy `json:"cleanup_policy,omitempty"`
	Retention     *Duration      `json:"retention,omitempty"`
}

// KafkaTopicOverride allows overriding default Kafka settings for a specific topic.
type KafkaTopicOverride struct {
	Table     Table                      `json:"table"`
	Initial   *KafkaTopicOverrideInitial `json:"initial,omitempty"`
	KeyColumn *string                    `json:"key_column,omitempty"`
}

// KafkaDestinationConfig holds Kafka-specific destination settings.
type KafkaDestinationConfig struct {
	TopicPrefix string               `json:"topic_prefix,omitempty"`
	Initial     KafkaTopicInitial    `json:"initial"`
	Topics      []KafkaTopicOverride `json:"topics,omitempty"`
}

// DestinationConfig holds per-pipeline destination settings.
type DestinationConfig struct {
	Connection string                 `json:"connection"`
	Kafka      KafkaDestinationConfig `json:"kafka"`
}

// PipelineSpec is the persistent definition of a pipeline.
type PipelineSpec struct {
	Name          string            `json:"name"`
	Source        SourceConfig      `json:"source"`
	Destination   DestinationConfig `json:"destination"`
	DesiredStatus DesiredStatus     `json:"desired_status"`
}

// CounterSnapshot is the JSON-serializable view of a counter.
type CounterSnapshot struct {
	Total  int64 `json:"total"`
	PerMin int64 `json:"per_minute"`
	PerHr  int64 `json:"per_hour"`
}

// PipelineStats holds the runtime counters for a pipeline.
type PipelineStats struct {
	Changes    CounterSnapshot `json:"changes"`
	Keepalives CounterSnapshot `json:"keepalives"`
}

// PipelineState is a PipelineSpec combined with its runtime status.
type PipelineState struct {
	PipelineSpec
	Status    EffectiveStatus `json:"status"`
	LastError *string         `json:"last_error"`
	Stats     *PipelineStats  `json:"stats,omitempty"`
}

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
	PipelineState
	Changes []Change `json:"changes"`
}
