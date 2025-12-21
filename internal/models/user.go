package models

type User struct {
	ID           int64  `db:"user_id"`
	Login        string `db:"login"`
	PasswordHash string `db:"password_hash"`
	CreatedAt    int64  `db:"created_at"`
}
