package admin

import "errors"

var (
	ErrNotFound      = errors.New("admin not found")
	ErrEmailConflict = errors.New("email already registered")
)
