package repository

import (
	"database/sql"
	"path/filepath"
	"testing"

	"crdledger/internal/models"

	_ "github.com/tursodatabase/libsql-client-go/libsql"
)

func newSyncTestDB(t *testing.T) *sql.DB {
	t.Helper()

	dbPath := filepath.Join(t.TempDir(), "sync-test.db")
	db, err := sql.Open("libsql", "file:"+dbPath)
	if err != nil {
		t.Fatalf("open test database: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	statements := []string{
		`CREATE TABLE transactions (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			seller_id INTEGER NOT NULL,
			buyer_id INTEGER NOT NULL,
			amount REAL NOT NULL,
			description TEXT NOT NULL,
			status TEXT NOT NULL DEFAULT 'pending',
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			paid_at DATETIME,
			photo_path TEXT,
			amount_paid REAL NOT NULL DEFAULT 0,
			confirmation_status TEXT NOT NULL DEFAULT 'confirmed',
			created_by INTEGER,
			client_operation_id TEXT
		)`,
		`CREATE UNIQUE INDEX idx_transactions_client_operation_id
			ON transactions(client_operation_id)
			WHERE client_operation_id IS NOT NULL`,
	}
	for _, statement := range statements {
		if _, err := db.Exec(statement); err != nil {
			t.Fatalf("create test schema: %v", err)
		}
	}

	return db
}

func TestCreateIdempotentReturnsCanonicalTransactionOnRetry(t *testing.T) {
	db := newSyncTestDB(t)
	repo := NewTransactionRepository(db)

	firstInput := &models.Transaction{
		SellerID:    1,
		BuyerID:     2,
		Amount:      2500,
		Description: "Lunch credit",
		Status:      "pending",
		CreatedByID: 1,
	}

	first, created, err := repo.CreateIdempotent(firstInput, "offline-op-123")
	if err != nil {
		t.Fatalf("first CreateIdempotent: %v", err)
	}
	if !created {
		t.Fatal("first CreateIdempotent created = false, want true")
	}
	if first.ID == 0 {
		t.Fatal("first transaction ID = 0, want persisted ID")
	}

	retryInput := &models.Transaction{
		SellerID:    1,
		BuyerID:     2,
		Amount:      9999,
		Description: "This retry payload must not create another debt",
		Status:      "pending",
		CreatedByID: 1,
	}

	retry, created, err := repo.CreateIdempotent(retryInput, "offline-op-123")
	if err != nil {
		t.Fatalf("retry CreateIdempotent: %v", err)
	}
	if created {
		t.Fatal("retry CreateIdempotent created = true, want false")
	}
	if retry.ID != first.ID {
		t.Fatalf("retry transaction ID = %d, want canonical ID %d", retry.ID, first.ID)
	}
	if retry.Amount != first.Amount || retry.Description != first.Description {
		t.Fatalf("retry returned mutated transaction: amount=%v description=%q; want amount=%v description=%q", retry.Amount, retry.Description, first.Amount, first.Description)
	}

	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM transactions WHERE client_operation_id = ?`, "offline-op-123").Scan(&count); err != nil {
		t.Fatalf("count operation rows: %v", err)
	}
	if count != 1 {
		t.Fatalf("rows for operation ID = %d, want 1", count)
	}
}

func TestCreateIdempotentRejectsMissingOperationID(t *testing.T) {
	db := newSyncTestDB(t)
	repo := NewTransactionRepository(db)

	_, created, err := repo.CreateIdempotent(&models.Transaction{}, "")
	if err == nil {
		t.Fatal("CreateIdempotent with empty operation ID returned nil error")
	}
	if created {
		t.Fatal("CreateIdempotent with empty operation ID created a transaction")
	}
}
