package repository

import (
	"database/sql"
	"fmt"
)

type Transaction struct {
	ID          int64
	SellerID    int64
	BuyerID     int64
	Amount      int64 // Storing in cents to avoid floating point issues
	Description string
	Status      string // pending or paid
	CreatedAt   string
	PaidAt      sql.NullString
}

type TransactionRepository struct {
	db *sql.DB
}

func NewTransactionRepository(db *sql.DB) *TransactionRepository {
	return &TransactionRepository{db: db}
}

func (r *TransactionRepository) CreateTransaction(sellerID, buyerID int64, amount int64, description string) (int64, error) {
	query := `INSERT INTO transactions (seller_id, buyer_id, amount, description) VALUES (?, ?, ?, ?)`
	result, err := r.db.Exec(query, sellerID, buyerID, amount, description)
	if err != nil {
		return 0, fmt.Errorf("failed to create transaction: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get transaction ID: %w", err)
	}

	return id, nil
}

func (r *TransactionRepository) GetTransactionsByUser(userID int64) ([]Transaction, error) {
	query := `SELECT id, seller_id, buyer_id, amount, description, status, created_at, paid_at
	          FROM transactions WHERE seller_id = ? OR buyer_id = ? ORDER BY created_at DESC`
	rows, err := r.db.Query(query, userID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get transactions: %w", err)
	}
	defer rows.Close()

	var transactions []Transaction
	for rows.Next() {
		var t Transaction
		var paidAt sql.NullString
		err := rows.Scan(&t.ID, &t.SellerID, &t.BuyerID, &t.Amount, &t.Description, &t.Status, &t.CreatedAt, &paidAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan transaction: %w", err)
		}
		t.PaidAt = paidAt
		transactions = append(transactions, t)
	}

	return transactions, nil
}

func (r *TransactionRepository) GetPendingTransactionsBySeller(sellerID int64) ([]Transaction, error) {
	query := `SELECT id, seller_id, buyer_id, amount, description, status, created_at, paid_at
	          FROM transactions WHERE seller_id = ? AND status = 'pending' ORDER BY created_at DESC`
	rows, err := r.db.Query(query, sellerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get pending transactions: %w", err)
	}
	defer rows.Close()

	var transactions []Transaction
	for rows.Next() {
		var t Transaction
		var paidAt sql.NullString
		err := rows.Scan(&t.ID, &t.SellerID, &t.BuyerID, &t.Amount, &t.Description, &t.Status, &t.CreatedAt, &paidAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan transaction: %w", err)
		}
		t.PaidAt = paidAt
		transactions = append(transactions, t)
	}

	return transactions, nil
}

func (r *TransactionRepository) GetPendingTransactionsByBuyer(buyerID int64) ([]Transaction, error) {
	query := `SELECT id, seller_id, buyer_id, amount, description, status, created_at, paid_at
	          FROM transactions WHERE buyer_id = ? AND status = 'pending' ORDER BY created_at DESC`
	rows, err := r.db.Query(query, buyerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get pending transactions: %w", err)
	}
	defer rows.Close()

	var transactions []Transaction
	for rows.Next() {
		var t Transaction
		var paidAt sql.NullString
		err := rows.Scan(&t.ID, &t.SellerID, &t.BuyerID, &t.Amount, &t.Description, &t.Status, &t.CreatedAt, &paidAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan transaction: %w", err)
		}
		t.PaidAt = paidAt
		transactions = append(transactions, t)
	}

	return transactions, nil
}

func (r *TransactionRepository) MarkTransactionAsPaid(transactionID int64) error {
	query := `UPDATE transactions SET status = 'paid', paid_at = CURRENT_TIMESTAMP WHERE id = ?`
	result, err := r.db.Exec(query, transactionID)
	if err != nil {
		return fmt.Errorf("failed to mark transaction as paid: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("no transaction found with ID %d", transactionID)
	}

	return nil
}

func (r *TransactionRepository) GetTransactionByID(transactionID int64) (*Transaction, error) {
	query := `SELECT id, seller_id, buyer_id, amount, description, status, created_at, paid_at
	          FROM transactions WHERE id = ?`
	row := r.db.QueryRow(query, transactionID)

	var t Transaction
	var paidAt sql.NullString
	err := row.Scan(&t.ID, &t.SellerID, &t.BuyerID, &t.Amount, &t.Description, &t.Status, &t.CreatedAt, &paidAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get transaction: %w", err)
	}
	t.PaidAt = paidAt

	return &t, nil
}