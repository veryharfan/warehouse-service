package domain

import "errors"

var (
	ErrNotFound        = errors.New("not found")
	ErrBadRequest      = errors.New("bad request")
	ErrInvalidRequest  = errors.New("invalid request")
	ErrValidation      = errors.New("validation error")
	ErrUnauthorized    = errors.New("unauthorized")
	ErrVersionMismatch = errors.New("version mismatch")
	ErrInternal        = errors.New("internal server error")
)
