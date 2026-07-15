package service

import (
	"errors"

	"crdledger/internal/models"
)

var ErrNotCounterparty = errors.New("only the other party on this transaction can confirm or reject it")
var ErrAlreadyResponded = errors.New("this transaction has already been confirmed or rejected")

// counterpartyID returns whichever side of the transaction did NOT create
// it — that's the only party allowed to confirm or reject, regardless of
// whether it was the seller recording a debt or a buyer self-reporting an
// order.
func counterpartyID(t *models.Transaction) int64 {
	if t.CreatedByID == t.SellerID {
		return t.BuyerID
	}
	return t.SellerID
}

// Confirm marks a pending transaction as confirmed, making it count toward
// both parties' balances. Only the participant who didn't create the entry
// may confirm it.
func (s *TransactionService) Confirm(transactionID, requestingUserID int64) error {
	t, err := s.transactions.GetByID(transactionID)
	if err != nil {
		return err
	}

	if requestingUserID != counterpartyID(t) {
		return ErrNotCounterparty
	}

	if t.ConfirmationStatus != models.ConfirmationPending {
		return ErrAlreadyResponded
	}

	return s.transactions.UpdateConfirmationStatus(transactionID, models.ConfirmationConfirmed)
}

// Reject marks a pending transaction as rejected. The record is kept, not
// deleted, so there's a dispute history — but a rejected transaction never
// counts toward either party's balance.
func (s *TransactionService) Reject(transactionID, requestingUserID int64) error {
	t, err := s.transactions.GetByID(transactionID)
	if err != nil {
		return err
	}

	if requestingUserID != counterpartyID(t) {
		return ErrNotCounterparty
	}

	if t.ConfirmationStatus != models.ConfirmationPending {
		return ErrAlreadyResponded
	}

	return s.transactions.UpdateConfirmationStatus(transactionID, models.ConfirmationRejected)
}
