package models

import "time"

type Order struct {
	Number     string    `db:"number"`
	UserID     int64     `db:"user_id"`
	Status     string    `db:"status"`
	Accrual    int64     `db:"accrual"`
	UploadedAt time.Time `db:"uploaded_at"`
}
