package models

import "time"

type Withdrawal struct {
	ID          int64     `db:"id"`
	UserID      int64     `db:"user_id"`
	OrderNumber string    `db:"order_number"`
	Sum         int64     `db:"sum"`
	ProcessedAt time.Time `db:"processed_at"`
}
