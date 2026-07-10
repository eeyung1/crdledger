package repository

import (
	"database/sql"
	"errors"

	"crdledger/internal/models"
)

var ErrUserNotFound = errors.New("user not found")

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(user *models.User) error {
	result, err := r.db.Exec(
		`INSERT INTO users (username, password_hash, display_name) VALUES (?, ?, ?)`,
		user.Username, user.PasswordHash, user.DisplayName,
	)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	user.ID = id
	return nil
}

func (r *UserRepository) GetByUsername(username string) (*models.User, error) {
	var u models.User
	var photoPath sql.NullString
	row := r.db.QueryRow(
		`SELECT id, username, password_hash, display_name, photo_path, created_at FROM users WHERE username = ?`,
		username,
	)

	err := row.Scan(&u.ID, &u.Username, &u.PasswordHash, &u.DisplayName, &photoPath, &u.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	u.PhotoPath = photoPath.String
	return &u, nil
}

func (r *UserRepository) GetByID(id int64) (*models.User, error) {
	var u models.User
	var photoPath sql.NullString
	row := r.db.QueryRow(
		`SELECT id, username, password_hash, display_name, photo_path, created_at FROM users WHERE id = ?`,
		id,
	)

	err := row.Scan(&u.ID, &u.Username, &u.PasswordHash, &u.DisplayName, &photoPath, &u.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	u.PhotoPath = photoPath.String
	return &u, nil
}

func (r *UserRepository) UpdatePhotoPath(userID int64, photoPath string) error {
	_, err := r.db.Exec(`UPDATE users SET photo_path = ? WHERE id = ?`, photoPath, userID)
	return err
}
