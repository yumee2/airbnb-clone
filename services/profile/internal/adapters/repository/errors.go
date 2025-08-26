package repository

import "errors"

var (
	ErrPhoneNumberExist = errors.New("provided phone number is already exists")
	ErrProfileNotFound  = errors.New("profile with provided ID was not found")
)
