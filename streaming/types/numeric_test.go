package types

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNumeric_String(t *testing.T) {
	assert.Equal(t, "123.45", Numeric("123.45").String())
	assert.Equal(t, "", Numeric("").String())
}

func TestNumeric_Float64(t *testing.T) {
	f, err := Numeric("99.5").Float64()
	require.NoError(t, err)
	assert.Equal(t, 99.5, f)

	_, err = Numeric("notanumber").Float64()
	assert.Error(t, err)
}

func TestNumeric_MarshalJSON(t *testing.T) {
	tests := []struct {
		n    Numeric
		want string
	}{
		{Numeric("123.45"), "123.45"},
		{Numeric("0"), "0"},
		{Numeric(""), "null"},
	}
	for _, tt := range tests {
		data, err := json.Marshal(tt.n)
		require.NoError(t, err)
		assert.Equal(t, tt.want, string(data))
	}
}

func TestNumeric_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		json string
		want Numeric
	}{
		{"123.45", Numeric("123.45")},
		{"null", Numeric("")},
		{" 99 ", Numeric("99")},
	}
	for _, tt := range tests {
		var n Numeric
		require.NoError(t, json.Unmarshal([]byte(tt.json), &n))
		assert.Equal(t, tt.want, n)
	}
}

func TestNumeric_RoundTrip(t *testing.T) {
	original := Numeric("12345.6789")
	data, err := json.Marshal(original)
	require.NoError(t, err)

	var decoded Numeric
	require.NoError(t, json.Unmarshal(data, &decoded))
	assert.Equal(t, original, decoded)
}
