package main

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	db, err := sql.Open("sqlite3", "./crdledger.db")
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	if err := createTables(db); err != nil {
		log.Fatalf("failed to create tables: %v", err)
	}

	log.Println("database ready: crdledger.db")
}

func createTables(db *sql.DB) error {
	usersTable := `
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT NOT NULL UNIQUE,
		password_hash TEXT NOT NULL,
		display_name TEXT NOT NULL,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	);`

	transactionsTable := `
	CREATE TABLE IF NOT EXISTS transactions (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		seller_id INTEGER NOT NULL REFERENCES users(id),
		buyer_id INTEGER NOT NULL REFERENCES users(id),
		amount REAL NOT NULL,
		description TEXT NOT NULL,
		status TEXT NOT NULL DEFAULT 'pending',
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		paid_at DATETIME
	);`

	if _, err := db.Exec(usersTable); err != nil {
		return err
	}
	if _, err := db.Exec(transactionsTable); err != nil {
		return err
	}
	return nil
}