package repository

import "errors"

var (
	ErrAptNotFound = errors.New("apartments with provided ID was not found")
)
