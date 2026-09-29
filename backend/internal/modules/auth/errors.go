package auth

import "errors"

var (
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrTokenRevoked       = errors.New("token has been revoked")
	ErrTokenInvalid       = errors.New("token is invalid")
	ErrTokenMissing       = errors.New("authorization token is missing")
)
