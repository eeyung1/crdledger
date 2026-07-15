package models

import "time"

type Transaction struct {
	ID                 int64
	SellerID           int64
	BuyerID            int64
	Amount             float64
	Description        string
	Status             string
	CreatedAt          time.Time
	PaidAt             *time.Time
	PhotoPath          *string // optional receipt/evidence photo, nil if none was attached
	AmountPaid         float64 // running total paid so far; equals Amount once Status is "paid"
	ConfirmationStatus string  // "pending", "confirmed", or "rejected" — whether the buyer has accepted this record
	CreatedByID        int64   // who submitted this entry — SellerID or BuyerID, whichever reported it. Confirmation is required from whichever side this ISN'T.
}

// Confirmation status values. Kept as constants so callers never have to
// guess or hand-type these strings.
const (
	ConfirmationPending   = "pending"
	ConfirmationConfirmed = "confirmed"
	ConfirmationRejected  = "rejected"
)
