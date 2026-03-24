// Package rerrors defines typed errors used across goragkit.
package rerrors

import "fmt"

// Kind classifies an operational error.
type Kind string

const (
	// InvalidInput indicates invalid caller input.
	InvalidInput Kind = "invalid_input"
	// External indicates failures from external systems.
	External Kind = "external"
	// NotFound indicates missing data.
	NotFound Kind = "not_found"
)

// Error is a structured error type for deterministic agent handling.
type Error struct {
	Kind    Kind   `json:"kind"`
	Op      string `json:"op"`
	Message string `json:"message"`
	Err     error  `json:"-"`
}

// Error implements error.
func (e *Error) Error() string {
	if e.Err == nil {
		return fmt.Sprintf("%s: %s", e.Op, e.Message)
	}
	return fmt.Sprintf("%s: %s: %v", e.Op, e.Message, e.Err)
}

// Unwrap returns wrapped error.
func (e *Error) Unwrap() error { return e.Err }

// Wrap returns a structured error.
func Wrap(kind Kind, op, msg string, err error) error {
	return &Error{Kind: kind, Op: op, Message: msg, Err: err}
}
