package repository

import "errors"

var (
	ErrEmailNotFound        = errors.New("user with this email not found")
	ErrRefreshTokenNotFound = errors.New("refresh token not found")
	ErrEmailExist           = errors.New("provided email is already exists")
)
