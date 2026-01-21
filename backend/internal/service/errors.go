package service

import "errors"

var (
	// ErrUnauthorized は認証に失敗した場合に返される
	ErrUnauthorized = errors.New("unauthorized")

	// ErrForbidden はユーザーに権限がない場合に返される
	ErrForbidden = errors.New("forbidden")
)
