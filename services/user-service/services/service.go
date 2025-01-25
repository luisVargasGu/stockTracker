package services

import "errors"

var (
	ErrUserNotFound   = errors.New("user not found")
	ErrUnauthorized   = errors.New("user is unauthorized")
	ErrDuplicateEmail = errors.New("email is a duplicate")
)
