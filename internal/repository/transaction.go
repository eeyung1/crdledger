package repository

import (
	"database/sql"
	"errors"
	"time"

	"crdledger/internal/models"
)

var ErrTransactionNotFound = errors.New("transaction not found")

type TransactionRepository struct {
	db *sql.DB
}

func NewTransactionRepository(db *sql.DB) *TransactionRepository {
	return &TransactionRepository{db: db}
}

func (r *TransactionRepository) Create(t *models.Transaction) error {
	t.ConfirmationStatus = models.ConfirmationPending
	result, err := r.db.Exec(
		`INSERT INTO transactions (seller_id, buyer_id, amount, description, status, photo_path, confirmation_status) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		t.SellerID, t.BuyerID, t.Amount, t.Description, t.Status, t.PhotoPath, t.ConfirmationStatus,
	)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	t.ID = id
	return nil
}

func (r *TransactionRepository) GetByID(id int64) (*models.Transaction, error) {
	var t models.Transaction
	row := r.db.QueryRow(
		`SELECT id, seller_id, buyer_id, amount, description, status, created_at, paid_at, photo_path, amount_paid, confirmation_status FROM transactions WHERE id = ?`,
		id,
	)

	err := row.Scan(&t.ID, &t.SellerID, &t.BuyerID, &t.Amount, &t.Description, &t.Status, &t.CreatedAt, &t.PaidAt, &t.PhotoPath, &t.AmountPaid, &t.ConfirmationStatus)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrTransactionNotFound
		}
		return nil, err
	}
	return &t, nil
}

func (r *TransactionRepository) GetByUser(userID int64) ([]models.Transaction, error) {
	rows, err := r.db.Query(
		`SELECT id, seller_id, buyer_id, amount, description, status, created_at, paid_at, photo_path, amount_paid, confirmation_status
		 FROM transactions
		 WHERE seller_id = ? OR buyer_id = ?
		 ORDER BY created_at DESC`,
		userID, userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var transactions []models.Transaction
	for rows.Next() {
		var t models.Transaction
		if err := rows.Scan(&t.ID, &t.SellerID, &t.BuyerID, &t.Amount, &t.Description, &t.Status, &t.CreatedAt, &t.PaidAt, &t.PhotoPath, &t.AmountPaid, &t.ConfirmationStatus); err != nil {
			return nil, err
		}
		transactions = append(transactions, t)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return transactions, nil
}

// GetPendingForBuyer returns transactions awaiting this user's confirmation,
// oldest first, so the longest-outstanding requests surface first.
func (r *TransactionRepository) GetPendingForBuyer(buyerID int64) ([]models.Transaction, error) {
	rows, err := r.db.Query(
		`SELECT id, seller_id, buyer_id, amount, description, status, created_at, paid_at, photo_path, amount_paid, confirmation_status
		 FROM transactions
		 WHERE buyer_id = ? AND confirmation_status = ?
		 ORDER BY created_at ASC`,
		buyerID, models.ConfirmationPending,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var transactions []models.Transaction
	for rows.Next() {
		var t models.Transaction
		if err := rows.Scan(&t.ID, &t.SellerID, &t.BuyerID, &t.Amount, &t.Description, &t.Status, &t.CreatedAt, &t.PaidAt, &t.PhotoPath, &t.AmountPaid, &t.ConfirmationStatus); err != nil {
			return nil, err
		}
		transactions = append(transactions, t)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return transactions, nil
}

// UpdateConfirmationStatus sets a transaction's confirmation_status —
// used when a buyer accepts or rejects a pending record.
func (r *TransactionRepository) UpdateConfirmationStatus(id int64, status string) error {
	_, err := r.db.Exec(
		`UPDATE transactions SET confirmation_status = ? WHERE id = ?`,
		status, id,
	)
	return err
}

// RecordPayment updates the running amount_paid total and, when the
// transaction is now fully settled, sets status to "paid" and stamps
// paid_at. For a partial payment, status stays "pending" and paidAt is
// nil — the row remains open for further payments.
func (r *TransactionRepository) RecordPayment(id int64, newAmountPaid float64, status string, paidAt *time.Time) error {
	_, err := r.db.Exec(
		`UPDATE transactions SET amount_paid = ?, status = ?, paid_at = ? WHERE id = ?`,
		newAmountPaid, status, paidAt, id,
	)
	return err
}
