package service

import (
	"errors"

	"crdledger/internal/models"
)

var ErrNotBuyer = errors.New("only the buyer can confirm or reject this transaction")
var ErrAlreadyResponded = errors.New("this transaction has already been confirmed or rejected")

// Confirm marks a pending transaction as confirmed by the buyer, making it
// count toward both parties' balances. Only the buyer named on the
// transaction may confirm it.
func (s *TransactionService) Confirm(transactionID, requestingUserID int64) error {
	t, err := s.transactions.GetByID(transactionID)
	if err != nil {
		return err
	}

	if t.BuyerID != requestingUserID {
		return ErrNotBuyer
	}

	if t.ConfirmationStatus != models.ConfirmationPending {
		return ErrAlreadyResponded
	}

	return s.transactions.UpdateConfirmationStatus(transactionID, models.ConfirmationConfirmed)
}

// Reject marks a pending transaction as rejected by the buyer. The record
// is kept, not deleted, so there's a dispute history — but a rejected
// transaction never counts toward either party's balance.
func (s *TransactionService) Reject(transactionID, requestingUserID int64) error {
	t, err := s.transactions.GetByID(transactionID)
	if err != nil {
		return err
	}

	if t.BuyerID != requestingUserID {
		return ErrNotBuyer
	}

	if t.ConfirmationStatus != models.ConfirmationPending {
		return ErrAlreadyResponded
	}

	return s.transactions.UpdateConfirmationStatus(transactionID, models.ConfirmationRejected)
}
