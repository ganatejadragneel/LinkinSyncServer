package models

import "time"

type Message struct {
	ID        int64     `json:"id"`
	UserEmail string    `json:"user_email"`
	Username  string    `json:"username"`
	Text      string    `json:"text"`
	CreatedAt time.Time `json:"created_at"`
}
