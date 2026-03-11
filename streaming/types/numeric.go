package types

import (
	"strconv"
	"strings"
)

// Numeric preserves the exact decimal representation from PostgreSQL.
// Use String() for the raw value or Float64() when a floating-point approximation is acceptable.
// Marshals as a JSON number to preserve numeric semantics.
type Numeric string

func (n Numeric) String() string {
	return string(n)
}

func (n Numeric) Float64() (float64, error) {
	return strconv.ParseFloat(string(n), 64)
}

func (n Numeric) MarshalJSON() ([]byte, error) {
	if n == "" {
		return []byte("null"), nil
	}
	return []byte(n), nil
}

func (n *Numeric) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		*n = ""
		return nil
	}
	*n = Numeric(strings.TrimSpace(string(data)))
	return nil
}
