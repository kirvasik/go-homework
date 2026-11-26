package domain

import "errors"

var (
	ErrNotFound = errors.New("movie not found")

	ErrInvalidRange = errors.New("invalid range")
)
