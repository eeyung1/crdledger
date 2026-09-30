package models

import "time"

type User struct {
	ID           int64
	Username     string
	PasswordHash string
	DisplayName  string
	PhotoPath    string
	AccountType  string
	SubscriptionStatus string
	SubscriptionPlan string
	SubscriptionEndsAt *time.Time
	CreatedAt    time.Time
}
