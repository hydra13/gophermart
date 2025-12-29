package models

import (
	"errors"
	"time"
)

const (
	OrderStatusNew        = "NEW"
	OrderStatusProcessing = "PROCESSING"
	OrderStatusInvalid    = "INVALID"
	OrderStatusProcessed  = "PROCESSED"
)

var (
	ErrOrderNotFound      = errors.New("order not found")
	ErrOrderAlreadyExists = errors.New("order already exists error")
)

type Order struct {
	Number     string    `db:"number"`
	UserID     int64     `db:"user_id"`
	Status     string    `db:"status"`
	Accrual    int64     `db:"accrual"`
	UploadedAt time.Time `db:"uploaded_at"`
}
