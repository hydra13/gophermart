package repositories

import "errors"

var (
	ErrAccountNotFound = errors.New("account not found")
	ErrOrderExists     = errors.New("order already exists")
	ErrOrderNotFound   = errors.New("order not found")
	ErrOrdersNotFound  = errors.New("orders not found")
	ErrConflict        = errors.New("conflict")
	ErrLoginExists     = errors.New("user with this login already exists")
	ErrUserNotFound    = errors.New("user not found")
)
