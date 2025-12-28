package models

import "errors"

var (
	ErrValidation         = errors.New("validation error")
	ErrInternal           = errors.New("internal error")
	ErrConflict           = errors.New("conflict error")
	ErrTokenNotValid      = errors.New("token validation error")
	ErrTokenNotFound      = errors.New("token not found error")
	ErrURLIsDeleted       = errors.New("url is deleted error")
	ErrOperationForbidden = errors.New("operation forbidden error")
	ErrOrderAlreadyExists = errors.New("order already exists error")
)
