package domain

import "errors"

var (
	ErrNotImplemented    = errors.New("not implemented")
	ErrInvalidRepoFormat = errors.New("invalid repo string format")
)
