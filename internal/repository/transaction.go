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
	result, err := r.db.Exec(
		`INSERT INTO transactions (seller_id, buyer_id, amount, description, status) VALUES (?, ?, ?, ?, ?)`,
		t.SellerID, t.BuyerID, t.Amount, t.Description, t.Status,
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
		`SELECT id, seller_id, buyer_id, amount, description, status, created_at, paid_at FROM transactions WHERE id = ?`,
		id,
	)

	err := row.Scan(&t.ID, &t.SellerID, &t.BuyerID, &t.Amount, &t.Description, &t.Status, &t.CreatedAt, &t.PaidAt)
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
		`SELECT id, seller_id, buyer_id, amount, description, status, created_at, paid_at
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
		if err := rows.Scan(&t.ID, &t.SellerID, &t.BuyerID, &t.Amount, &t.Description, &t.Status, &t.CreatedAt, &t.PaidAt); err != nil {
			return nil, err
		}
		transactions = append(transactions, t)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return transactions, nil
}

func (r *TransactionRepository) MarkPaid(id int64, paidAt time.Time) error {
	_, err := r.db.Exec(
		`UPDATE transactions SET status = 'paid', paid_at = ? WHERE id = ?`,
		paidAt, id,
	)
	return err
}
