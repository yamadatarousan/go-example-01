package repository

import "errors"

var (
	// ErrNotFound はリソースが見つからなかった場合に返される
	ErrNotFound = errors.New("resource not found")

	// ErrConflict はユニーク制約違反など、リソースの競合が発生した場合に返される
	ErrConflict = errors.New("resource already exists")
)
