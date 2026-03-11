package types

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

type ChangeEventTable struct {
	Schema string `json:"schema"`
	Table  string `json:"table"`
}

func (t ChangeEventTable) QualifiedName() string {
	return fmt.Sprintf("%s.%s", t.Schema, t.Table)
}

type ChangeEvent struct {
	Table     ChangeEventTable    `json:"table"`
	Operation Operation           `json:"operation"`
	Columns   []ChangeEventColumn `json:"columns"`
}

type ChangeEventColumn struct {
	Name     string `json:"name"`
	DataType string `json:"type"`
	IsKey    bool   `json:"is_key,omitempty"`
	// Value holds the parsed column data. The concrete type depends on the PostgreSQL type:
	//   nil              - NULL
	//   int64            - int2, int4, int8
	//   float64          - float4, float8
	//   bool             - bool
	//   Numeric          - numeric (marshals as JSON number)
	//   json.RawMessage  - json, jsonb. "null" at SQL column level is nil, not "null", to preserve JSON semantics.
	//   time.Time        - timestamp, timestamptz
	//   string           - everything else (text, varchar, uuid, ...)
	Value any `json:"value"`
}

// UnmarshalJSON deserializes a ChangeEventColumn from JSON, restoring the
// correct Go type for Value based on DataType. This ensures round-trip
// consistency: the types produced by parseValue on the WAL capture side are
// recovered after a JSON serialization cycle (e.g. through Kafka).
func (c *ChangeEventColumn) UnmarshalJSON(data []byte) error {
	var raw struct {
		Name     string          `json:"name"`
		DataType string          `json:"type"`
		IsKey    bool            `json:"is_key,omitempty"`
		Value    json.RawMessage `json:"value"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	c.Name = raw.Name
	c.DataType = raw.DataType
	c.IsKey = raw.IsKey

	if len(raw.Value) == 0 || string(raw.Value) == "null" {
		c.Value = nil
		return nil
	}

	switch c.DataType {
	case "int2", "int4", "int8":
		var v int64
		if err := json.Unmarshal(raw.Value, &v); err != nil {
			return fmt.Errorf("failed to unmarshal %s value: %w", c.DataType, err)
		}
		c.Value = v
	case "float4", "float8":
		var v float64
		if err := json.Unmarshal(raw.Value, &v); err != nil {
			return fmt.Errorf("failed to unmarshal %s value: %w", c.DataType, err)
		}
		c.Value = v
	case "bool":
		var v bool
		if err := json.Unmarshal(raw.Value, &v); err != nil {
			return fmt.Errorf("failed to unmarshal bool value: %w", err)
		}
		c.Value = v
	case "numeric":
		c.Value = Numeric(strings.TrimSpace(string(raw.Value)))
	case "json", "jsonb":
		c.Value = raw.Value
	case "timestamp", "timestamptz":
		var s string
		if err := json.Unmarshal(raw.Value, &s); err != nil {
			return fmt.Errorf("failed to unmarshal %s value: %w", c.DataType, err)
		}
		t, err := time.Parse(time.RFC3339Nano, s)
		if err != nil {
			c.Value = s
		} else {
			c.Value = t
		}
	default:
		var s string
		if err := json.Unmarshal(raw.Value, &s); err != nil {
			return fmt.Errorf("failed to unmarshal string value: %w", err)
		}
		c.Value = s
	}
	return nil
}
