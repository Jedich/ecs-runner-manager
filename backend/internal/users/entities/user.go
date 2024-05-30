package entities

import "time"

type User struct {
	UserID    int64
	Username  string
	Email     string
	Password  string
	IsActive  bool
	CreatedAt time.Time
	UpdatedAt time.Time
}
