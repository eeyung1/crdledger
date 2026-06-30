package repository

import (
	"database/sql"
	"fmt"
)

type User struct {
	ID           int64
	Username     string
	PasswordHash string
	DisplayName  string
	CreatedAt    string
}

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) CreateUser(username, passwordHash, displayName string) (int64, error) {
	query := `INSERT INTO users (username, password_hash, display_name) VALUES (?, ?, ?)`
	result, err := r.db.Exec(query, username, passwordHash, displayName)
	if err != nil {
		return 0, fmt.Errorf("failed to create user: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get user ID: %w", err)
	}

	return id, nil
}

func (r *UserRepository) GetUserByUsername(username string) (*User, error) {
	query := `SELECT id, username, password_hash, display_name, created_at FROM users WHERE username = ?`
	row := r.db.QueryRow(query, username)

	var user User
	err := row.Scan(&user.ID, &user.Username, &user.PasswordHash, &user.DisplayName, &user.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &user, nil
}

func (r *UserRepository) GetUserByID(id int64) (*User, error) {
	query := `SELECT id, username, password_hash, display_name, created_at FROM users WHERE id = ?`
	row := r.db.QueryRow(query, id)

	var user User
	err := row.Scan(&user.ID, &user.Username, &user.PasswordHash, &user.DisplayName, &user.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &user, nil
}