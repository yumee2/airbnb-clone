package service

import "errors"

var (
	ErrProfileNotFound = errors.New("profile with provided ID was not found")
)
