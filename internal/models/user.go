package models

import "time"

type User struct {
	ID           int64
	Username     string
	PasswordHash string
	DisplayName  string
	PhotoPath    string
	CreatedAt    time.Time
}
