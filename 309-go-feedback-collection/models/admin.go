package models

import "time"

type Admin struct {
	ID        int64     `db:"id"`
	Username  string    `db:"username"`
	Password  string    `db:"password"`
	CreatedAt time.Time `db:"created_at"`
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token string `json:"token"`
}
