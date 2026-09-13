package service

import "errors"

var (
	ErrInvalidInput         = errors.New("invalid input")
	ErrEmailTaken           = errors.New("email already taken")
	ErrUnauthorized         = errors.New("unauthorized")
	ErrForbidden            = errors.New("forbidden")
	ErrTokenInvalid         = errors.New("token invalid or expired")
	ErrEmailAlreadyVerified = errors.New("email already verified")
	ErrEmailNotVerified     = errors.New("email not verified")
)
