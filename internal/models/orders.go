package models

import "time"

type Orders struct {
	ID         int       `json:"id" db:"id"`
	UserId     int       `json:"user_id" db:"user_id"`
	Status     string    `json:"status" db:"status"`
	TotalPrice int       `json:"total_price" db:"total_price"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
}
