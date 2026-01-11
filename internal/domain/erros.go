package domain

import "errors"

var (
	// ErrNotFound is returned when a requested resource is not found
	ErrNotFound = errors.New("resource not found")

	// ErrUnauthorized is returned when authentication fails
	ErrUnauthorized = errors.New("unauthorized")

	// ErrForbidden is returned when the user doesn't have permission
	ErrForbidden = errors.New("forbidden")

	// ErrConflict is returned when there's a unique constraint violation
	ErrConflict = errors.New("resource already exists")

	// ErrInvalidInput is returned when input validation fails
	ErrInvalidInput = errors.New("invalid input")
)

