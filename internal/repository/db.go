package repository

import (
    "database/sql"
    _ "github.com/mattn/go-sqlite3"
)

func InitDB(dbPath string) (*sql.DB, error) {
    db, err := sql.Open("sqlite3", dbPath)
    if err != nil {
        return nil, err
    }

    if err := db.Ping(); err != nil {
        return nil, err
    }

    return db, nil
}

func CreateTables(db *sql.DB) error {
    queries := []string{
        `CREATE TABLE IF NOT EXISTS users (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            username TEXT UNIQUE NOT NULL,
            password_hash TEXT NOT NULL,
            display_name TEXT NOT NULL,
            created_at DATETIME DEFAULT CURRENT_TIMESTAMP
        );`,
        `CREATE TABLE IF NOT EXISTS transactions (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            seller_id INTEGER NOT NULL,
            buyer_id INTEGER NOT NULL,
            amount INTEGER NOT NULL,
            description TEXT NOT NULL,
            status TEXT DEFAULT 'pending',
            created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
            paid_at DATETIME,
            FOREIGN KEY (seller_id) REFERENCES users(id),
            FOREIGN KEY (buyer_id) REFERENCES users(id)
        );`,
        `CREATE INDEX IF NOT EXISTS idx_transactions_seller ON transactions(seller_id);`,
        `CREATE INDEX IF NOT EXISTS idx_transactions_buyer ON transactions(buyer_id);`,
        `CREATE INDEX IF NOT EXISTS idx_transactions_status ON transactions(status);`,
    }

    for _, query := range queries {
        if _, err := db.Exec(query); err != nil {
            return err
        }
    }

    return nil
}