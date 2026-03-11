package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOperationConstants(t *testing.T) {
	assert.Equal(t, Operation("INSERT"), OperationInsert)
	assert.Equal(t, Operation("UPDATE"), OperationUpdate)
	assert.Equal(t, Operation("DELETE"), OperationDelete)
}
