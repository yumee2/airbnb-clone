package service

import "errors"

var (
	ErrAptNotFound   = errors.New("apartment not found")
	ErrInvalidInput  = errors.New("invalid input data")
	ErrInvalidImage  = errors.New("invalid image file")
	ErrImageTooLarge = errors.New("image size too large")
)
