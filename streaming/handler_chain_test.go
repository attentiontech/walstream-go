package streaming

import (
	"context"
	"errors"
	"testing"

	stypes "github.com/attentiontech/walstream-go/streaming/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestChangeHandlerChain(t *testing.T) {
	var order []int
	makeHandler := func(id int) ChangeHandler {
		return ChangeHandlerFunc(func(_ context.Context, _ stypes.ChangeEvent) error {
			order = append(order, id)
			return nil
		})
	}

	chain := NewChangeHandlerChain(makeHandler(1), makeHandler(2), makeHandler(3))
	err := chain.HandleChange(context.Background(), stypes.ChangeEvent{})
	require.NoError(t, err)
	assert.Equal(t, []int{1, 2, 3}, order)
}

func TestChangeHandlerChain_StopsOnError(t *testing.T) {
	want := errors.New("fail at 2")
	var order []int
	chain := NewChangeHandlerChain(
		ChangeHandlerFunc(func(_ context.Context, _ stypes.ChangeEvent) error {
			order = append(order, 1)
			return nil
		}),
		ChangeHandlerFunc(func(_ context.Context, _ stypes.ChangeEvent) error {
			order = append(order, 2)
			return want
		}),
		ChangeHandlerFunc(func(_ context.Context, _ stypes.ChangeEvent) error {
			order = append(order, 3)
			return nil
		}),
	)
	got := chain.HandleChange(context.Background(), stypes.ChangeEvent{})
	assert.ErrorIs(t, got, want)
	assert.Equal(t, []int{1, 2}, order, "chain should stop at first error")
}

func TestChangeHandlerChain_Empty(t *testing.T) {
	chain := NewChangeHandlerChain()
	err := chain.HandleChange(context.Background(), stypes.ChangeEvent{})
	assert.NoError(t, err)
}
