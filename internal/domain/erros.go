package domain

import "errors"

// ビジネスルール違反のエラー定義
// データベースやHTTPに依存しない、純粋なビジネスロジックのエラーのみをここに定義する

var (
	// ErrInvalidInput は入力値がビジネスルールに違反している場合に返される
	ErrInvalidInput = errors.New("invalid input")

	// 以下は具体的なビジネスルール違反の例（必要に応じて追加）
	// ErrTodoNameRequired = errors.New("todo name is required")
	// ErrTodoNameTooLong = errors.New("todo name exceeds maximum length")
	// ErrPasswordTooShort = errors.New("password must be at least 8 characters")
	// ErrInvalidEmailFormat = errors.New("invalid email format")
)
