package streaming

import (
	"context"

	stypes "github.com/attentiontech/walstream-go/streaming/types"
)

// ChangeHandlerChain fans out each event to multiple handlers sequentially.
// It stops and returns the first error encountered.
type ChangeHandlerChain []ChangeHandler

// NewChangeHandlerChain creates a chain of handlers called one after the other.
func NewChangeHandlerChain(handlers ...ChangeHandler) ChangeHandlerChain {
	return ChangeHandlerChain(handlers)
}

func (c ChangeHandlerChain) HandleChange(ctx context.Context, event stypes.ChangeEvent) error {
	for _, h := range c {
		if err := h.HandleChange(ctx, event); err != nil {
			return err
		}
	}
	return nil
}
