package models

import "errors"

var (
	ErrNotEnoughMoney = errors.New("not enough money")
)

type Account struct {
	UserID    int64 `db:"user_id"`
	Current   int64 `db:"current"`
	Withdrawn int64 `db:"withdrawn"`
}
