package service

import (
	"errors"
	"time"

	"crdledger/internal/repository"
)

var ErrNotSeller = errors.New("only the seller can mark this transaction as paid")
var ErrAlreadyPaid = errors.New("transaction is already marked as paid")
var ErrPaidDateInFuture = errors.New("payment date cannot be in the future")
var ErrPaidDateBeforeCreated = errors.New("payment date cannot be before the transaction was created")

func (s *TransactionService) MarkPaid(transactionID, requestingUserID int64, paidAt time.Time) error {
	t, err := s.transactions.GetByID(transactionID)
	if err != nil {
		if errors.Is(err, repository.ErrTransactionNotFound) {
			return err
		}
		return err
	}

	if t.SellerID != requestingUserID {
		return ErrNotSeller
	}

	if t.Status == "paid" {
		return ErrAlreadyPaid
	}

	today := time.Now().Truncate(24 * time.Hour)
	createdDate := t.CreatedAt.Truncate(24 * time.Hour)

	if paidAt.After(today) {
		return ErrPaidDateInFuture
	}
	if paidAt.Before(createdDate) {
		return ErrPaidDateBeforeCreated
	}

	return s.transactions.MarkPaid(transactionID, paidAt)
}
