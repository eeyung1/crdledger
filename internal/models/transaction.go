package models

import "time"

type Transaction struct {
	ID          int64
	SellerID    int64
	BuyerID     int64
	Amount      float64
	Description string
	Status      string
	CreatedAt   time.Time
	PaidAt      *time.Time
}
