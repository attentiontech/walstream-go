package types

import "errors"

var (
	// ErrPipelineNotFound is returned when a pipeline name doesn't exist.
	ErrPipelineNotFound = errors.New("pipeline not found")

	// ErrValidation is returned when a pipeline spec fails validation.
	ErrValidation = errors.New("validation error")
)
