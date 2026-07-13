package service

import (
	"errors"
	"time"

	"crdledger/internal/models"
	"crdledger/internal/repository"
)

var ErrNotSeller = errors.New("only the seller can mark this transaction as paid")
var ErrAlreadyPaid = errors.New("transaction is already marked as paid")
var ErrPaidDateInFuture = errors.New("payment date cannot be in the future")
var ErrPaidDateBeforeCreated = errors.New("payment date cannot be before the transaction was created")
var ErrInvalidPaymentAmount = errors.New("payment amount must be positive")
var ErrPaymentExceedsBalance = errors.New("payment amount is more than the remaining balance")
var ErrNotConfirmed = errors.New("the buyer must confirm this transaction before it can be marked as paid")

// epsilon absorbs float rounding noise so a payment of, say, exactly the
// remaining balance doesn't get rejected for being 0.0000000001 over.
const epsilon = 0.005

func (s *TransactionService) MarkPaid(transactionID, requestingUserID int64, paymentAmount float64, paidAt time.Time) error {
	t, err := s.transactions.GetByID(transactionID)
	if err != nil {
		return err
	}

	if t.SellerID != requestingUserID {
		return ErrNotSeller
	}

	if t.ConfirmationStatus != models.ConfirmationConfirmed {
		return ErrNotConfirmed
	}

	if t.Status == "paid" {
		return ErrAlreadyPaid
	}

	if paymentAmount <= 0 {
		return ErrInvalidPaymentAmount
	}

	remaining := t.Amount - t.AmountPaid
	if paymentAmount > remaining+epsilon {
		return ErrPaymentExceedsBalance
	}

	today := time.Now().Truncate(24 * time.Hour)
	createdDate := t.CreatedAt.Truncate(24 * time.Hour)

	if paidAt.After(today) {
		return ErrPaidDateInFuture
	}
	if paidAt.Before(createdDate) {
		return ErrPaidDateBeforeCreated
	}

	newAmountPaid := t.AmountPaid + paymentAmount

	// paidAt tracks the date of the most recent payment, whether this
	// settles the transaction or is only a partial payment — so the UI
	// always has a date to show next to the status, not just once "paid".
	recordedAt := paidAt

	if newAmountPaid >= t.Amount-epsilon {
		// Fully settled — clamp to the exact amount so float drift never
		// leaves a transaction stuck a fraction of a cent short of "paid".
		return s.transactions.RecordPayment(transactionID, t.Amount, "paid", &recordedAt)
	}

	return s.transactions.RecordPayment(transactionID, newAmountPaid, "pending", &recordedAt)
}

// ErrTransactionNotFound is re-exported here for handlers that only import
// the service package.
var ErrTransactionNotFound = repository.ErrTransactionNotFound
