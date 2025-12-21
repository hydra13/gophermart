package models

type Account struct {
	UserID    int64 `db:"user_id"`
	Current   int64 `db:"current"`
	Withdrawn int64 `db:"withdrawn"`
}
