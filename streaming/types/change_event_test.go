package types

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestChangeEventTable_QualifiedName(t *testing.T) {
	tests := []struct {
		schema string
		table  string
		want   string
	}{
		{"public", "users", "public.users"},
		{"myschema", "orders", "myschema.orders"},
		{"", "", "."},
	}
	for _, tt := range tests {
		tbl := ChangeEventTable{Schema: tt.schema, Table: tt.table}
		assert.Equal(t, tt.want, tbl.QualifiedName())
	}
}

func TestChangeEventColumn_UnmarshalJSON(t *testing.T) {
	ts := time.Date(2025, 6, 15, 10, 30, 0, 0, time.UTC)

	tests := []struct {
		name     string
		json     string
		wantName string
		wantType string
		wantKey  bool
		check    func(t *testing.T, v any)
	}{
		{
			name:     "null value",
			json:     `{"name":"id","type":"int4","value":null}`,
			wantName: "id",
			wantType: "int4",
			check:    func(t *testing.T, v any) { assert.Nil(t, v) },
		},
		{
			name:     "missing value",
			json:     `{"name":"id","type":"int4"}`,
			wantName: "id",
			wantType: "int4",
			check:    func(t *testing.T, v any) { assert.Nil(t, v) },
		},
		{
			name:     "int2",
			json:     `{"name":"age","type":"int2","value":42}`,
			wantName: "age",
			wantType: "int2",
			check:    func(t *testing.T, v any) { assert.Equal(t, int64(42), v) },
		},
		{
			name:     "int4",
			json:     `{"name":"count","type":"int4","value":1000}`,
			wantName: "count",
			wantType: "int4",
			check:    func(t *testing.T, v any) { assert.Equal(t, int64(1000), v) },
		},
		{
			name:     "int8",
			json:     `{"name":"big","type":"int8","value":9999999999}`,
			wantName: "big",
			wantType: "int8",
			check:    func(t *testing.T, v any) { assert.Equal(t, int64(9999999999), v) },
		},
		{
			name:     "float4",
			json:     `{"name":"score","type":"float4","value":3.14}`,
			wantName: "score",
			wantType: "float4",
			check:    func(t *testing.T, v any) { assert.InDelta(t, 3.14, v, 0.001) },
		},
		{
			name:     "float8",
			json:     `{"name":"precise","type":"float8","value":2.718281828}`,
			wantName: "precise",
			wantType: "float8",
			check:    func(t *testing.T, v any) { assert.Equal(t, 2.718281828, v) },
		},
		{
			name:     "bool true",
			json:     `{"name":"active","type":"bool","value":true}`,
			wantName: "active",
			wantType: "bool",
			check:    func(t *testing.T, v any) { assert.Equal(t, true, v) },
		},
		{
			name:     "bool false",
			json:     `{"name":"deleted","type":"bool","value":false}`,
			wantName: "deleted",
			wantType: "bool",
			check:    func(t *testing.T, v any) { assert.Equal(t, false, v) },
		},
		{
			name:     "numeric",
			json:     `{"name":"price","type":"numeric","value":99.95}`,
			wantName: "price",
			wantType: "numeric",
			check:    func(t *testing.T, v any) { assert.Equal(t, Numeric("99.95"), v) },
		},
		{
			name:     "json",
			json:     `{"name":"meta","type":"json","value":{"key":"val"}}`,
			wantName: "meta",
			wantType: "json",
			check: func(t *testing.T, v any) {
				raw, ok := v.(json.RawMessage)
				require.True(t, ok, "expected json.RawMessage, got %T", v)
				assert.JSONEq(t, `{"key":"val"}`, string(raw))
			},
		},
		{
			name:     "jsonb array",
			json:     `{"name":"tags","type":"jsonb","value":[1,2,3]}`,
			wantName: "tags",
			wantType: "jsonb",
			check: func(t *testing.T, v any) {
				raw, ok := v.(json.RawMessage)
				require.True(t, ok, "expected json.RawMessage, got %T", v)
				assert.JSONEq(t, `[1,2,3]`, string(raw))
			},
		},
		{
			name:     "timestamp",
			json:     `{"name":"created","type":"timestamp","value":"2025-06-15T10:30:00Z"}`,
			wantName: "created",
			wantType: "timestamp",
			check: func(t *testing.T, v any) {
				got, ok := v.(time.Time)
				require.True(t, ok, "expected time.Time, got %T", v)
				assert.True(t, ts.Equal(got))
			},
		},
		{
			name:     "timestamptz",
			json:     `{"name":"updated","type":"timestamptz","value":"2025-06-15T10:30:00Z"}`,
			wantName: "updated",
			wantType: "timestamptz",
			check: func(t *testing.T, v any) {
				got, ok := v.(time.Time)
				require.True(t, ok, "expected time.Time, got %T", v)
				assert.True(t, ts.Equal(got))
			},
		},
		{
			name:     "timestamp unparseable falls back to string",
			json:     `{"name":"created","type":"timestamp","value":"not-a-date"}`,
			wantName: "created",
			wantType: "timestamp",
			check:    func(t *testing.T, v any) { assert.Equal(t, "not-a-date", v) },
		},
		{
			name:     "text (default)",
			json:     `{"name":"name","type":"text","value":"hello"}`,
			wantName: "name",
			wantType: "text",
			check:    func(t *testing.T, v any) { assert.Equal(t, "hello", v) },
		},
		{
			name:     "varchar (default)",
			json:     `{"name":"email","type":"varchar","value":"a@b.com"}`,
			wantName: "email",
			wantType: "varchar",
			check:    func(t *testing.T, v any) { assert.Equal(t, "a@b.com", v) },
		},
		{
			name:     "uuid (default)",
			json:     `{"name":"id","type":"uuid","value":"550e8400-e29b-41d4-a716-446655440000"}`,
			wantName: "id",
			wantType: "uuid",
			check:    func(t *testing.T, v any) { assert.Equal(t, "550e8400-e29b-41d4-a716-446655440000", v) },
		},
		{
			name:     "is_key flag",
			json:     `{"name":"id","type":"int4","is_key":true,"value":1}`,
			wantName: "id",
			wantType: "int4",
			wantKey:  true,
			check:    func(t *testing.T, v any) { assert.Equal(t, int64(1), v) },
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var col ChangeEventColumn
			require.NoError(t, json.Unmarshal([]byte(tt.json), &col))
			assert.Equal(t, tt.wantName, col.Name)
			assert.Equal(t, tt.wantType, col.DataType)
			assert.Equal(t, tt.wantKey, col.IsKey)
			tt.check(t, col.Value)
		})
	}
}

func TestChangeEventColumn_UnmarshalJSON_Errors(t *testing.T) {
	tests := []struct {
		name string
		json string
	}{
		{"invalid json", `{broken`},
		{"int4 with string value", `{"name":"x","type":"int4","value":"notint"}`},
		{"float8 with string value", `{"name":"x","type":"float8","value":"notfloat"}`},
		{"bool with string value", `{"name":"x","type":"bool","value":"notbool"}`},
		{"timestamp with number value", `{"name":"x","type":"timestamp","value":12345}`},
		{"text with number value", `{"name":"x","type":"text","value":12345}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var col ChangeEventColumn
			assert.Error(t, json.Unmarshal([]byte(tt.json), &col))
		})
	}
}

func TestChangeEvent_RoundTrip(t *testing.T) {
	ts := time.Date(2025, 1, 2, 3, 4, 5, 0, time.UTC)
	original := ChangeEvent{
		Table:     ChangeEventTable{Schema: "public", Table: "users"},
		Operation: OperationInsert,
		Columns: []ChangeEventColumn{
			{Name: "id", DataType: "int4", IsKey: true, Value: int64(1)},
			{Name: "name", DataType: "text", Value: "Alice"},
			{Name: "score", DataType: "float8", Value: 9.5},
			{Name: "active", DataType: "bool", Value: true},
			{Name: "price", DataType: "numeric", Value: Numeric("123.45")},
			{Name: "meta", DataType: "jsonb", Value: json.RawMessage(`{"k":"v"}`)},
			{Name: "created", DataType: "timestamptz", Value: ts},
			{Name: "deleted_at", DataType: "timestamptz", Value: nil},
		},
	}

	data, err := json.Marshal(original)
	require.NoError(t, err)

	var decoded ChangeEvent
	require.NoError(t, json.Unmarshal(data, &decoded))

	assert.Equal(t, original.Table.QualifiedName(), decoded.Table.QualifiedName())
	assert.Equal(t, original.Operation, decoded.Operation)
	require.Len(t, decoded.Columns, len(original.Columns))

	for i, col := range decoded.Columns {
		orig := original.Columns[i]
		assert.Equal(t, orig.Name, col.Name)
		assert.Equal(t, orig.DataType, col.DataType)
		assert.Equal(t, orig.IsKey, col.IsKey)

		switch expected := orig.Value.(type) {
		case nil:
			assert.Nil(t, col.Value)
		case json.RawMessage:
			raw, ok := col.Value.(json.RawMessage)
			require.True(t, ok, "column %q: expected json.RawMessage, got %T", col.Name, col.Value)
			assert.JSONEq(t, string(expected), string(raw))
		case time.Time:
			got, ok := col.Value.(time.Time)
			require.True(t, ok, "column %q: expected time.Time, got %T", col.Name, col.Value)
			assert.True(t, expected.Equal(got), "column %q: time = %v, want %v", col.Name, got, expected)
		default:
			assert.Equal(t, orig.Value, col.Value)
		}
	}
}
