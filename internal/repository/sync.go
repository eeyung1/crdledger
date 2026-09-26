package repository

import (
	"database/sql"
	"errors"

	"crdledger/internal/models"
)

// CreateIdempotent records a transaction exactly once for a client-generated
// operation ID. Offline clients may safely retry the same operation after a
// timeout or reconnect without creating duplicate debt.
//
// The database unique index on client_operation_id is the final concurrency
// guard. If another request wins the race, we return the transaction that was
// already created for that operation ID.
func (r *TransactionRepository) CreateIdempotent(t *models.Transaction, operationID string) (*models.Transaction, bool, error) {
	if operationID == "" {
		return nil, false, errors.New("client operation id is required")
	}

	if existing, err := r.GetByClientOperationID(operationID); err == nil {
		return existing, false, nil
	} else if !errors.Is(err, ErrTransactionNotFound) {
		return nil, false, err
	}

	t.ConfirmationStatus = models.ConfirmationPending
	result, err := r.db.Exec(
		`INSERT INTO transactions (seller_id, buyer_id, amount, description, status, photo_path, confirmation_status, created_by, client_operation_id)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		t.SellerID, t.BuyerID, t.Amount, t.Description, t.Status, t.PhotoPath, t.ConfirmationStatus, t.CreatedByID, operationID,
	)
	if err != nil {
		// A concurrent retry can lose the UNIQUE race. In that case the
		// operation now exists, so return the canonical row as success.
		if existing, lookupErr := r.GetByClientOperationID(operationID); lookupErr == nil {
			return existing, false, nil
		}
		return nil, false, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, false, err
	}
	t.ID = id
	return t, true, nil
}

// GetByClientOperationID returns the canonical server transaction created for
// an offline mutation. The operation ID is deliberately scoped to a single
// immutable transaction creation and is never reused by the client.
func (r *TransactionRepository) GetByClientOperationID(operationID string) (*models.Transaction, error) {
	var t models.Transaction
	row := r.db.QueryRow(
		`SELECT id, seller_id, buyer_id, amount, description, status, created_at, paid_at, photo_path, amount_paid, confirmation_status, created_by
		 FROM transactions WHERE client_operation_id = ?`,
		operationID,
	)

	err := row.Scan(&t.ID, &t.SellerID, &t.BuyerID, &t.Amount, &t.Description, &t.Status, &t.CreatedAt, &t.PaidAt, &t.PhotoPath, &t.AmountPaid, &t.ConfirmationStatus, &t.CreatedByID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrTransactionNotFound
		}
		return nil, err
	}
	return &t, nil
}
