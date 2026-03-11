package streaming

import (
	"context"
	"errors"
	"testing"

	stypes "github.com/attentiontech/walstream-go/streaming/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestChangeHandlerFunc(t *testing.T) {
	var called bool
	var gotEvent stypes.ChangeEvent
	fn := ChangeHandlerFunc(func(_ context.Context, e stypes.ChangeEvent) error {
		called = true
		gotEvent = e
		return nil
	})

	event := stypes.ChangeEvent{
		Table:     stypes.ChangeEventTable{Schema: "public", Table: "test"},
		Operation: stypes.OperationUpdate,
	}
	err := fn.HandleChange(context.Background(), event)
	require.NoError(t, err)
	assert.True(t, called, "handler was not called")
	assert.Equal(t, event.Operation, gotEvent.Operation)
}

func TestChangeHandlerFunc_Error(t *testing.T) {
	want := errors.New("handler failed")
	fn := ChangeHandlerFunc(func(_ context.Context, _ stypes.ChangeEvent) error {
		return want
	})
	got := fn.HandleChange(context.Background(), stypes.ChangeEvent{})
	assert.ErrorIs(t, got, want)
}
