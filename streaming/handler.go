package streaming

import (
	"context"

	stypes "github.com/attentiontech/walstream-go/streaming/types"
)

// ChangeHandler handles WAL change events.
type ChangeHandler interface {
	HandleChange(ctx context.Context, event stypes.ChangeEvent) error
}

// ChangeHandlerFunc is an adapter to allow the use of ordinary functions as ChangeHandlers.
type ChangeHandlerFunc func(ctx context.Context, event stypes.ChangeEvent) error

func (f ChangeHandlerFunc) HandleChange(ctx context.Context, event stypes.ChangeEvent) error {
	return f(ctx, event)
}
