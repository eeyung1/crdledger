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
		`INSERT INTO users (username, password_hash, display_name, account_type, subscription_status, subscription_plan) VALUES (?, ?, ?, ?, ?, ?)`,
		user.Username, user.PasswordHash, user.DisplayName, user.AccountType, user.SubscriptionStatus, user.SubscriptionPlan,
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
	var subscriptionEndsAt sql.NullTime
	row := r.db.QueryRow(
		`SELECT id, username, password_hash, display_name, photo_path, account_type, subscription_status, subscription_plan, subscription_ends_at, created_at FROM users WHERE username = ?`,
		username,
	)

	err := row.Scan(&u.ID, &u.Username, &u.PasswordHash, &u.DisplayName, &photoPath, &u.AccountType, &u.SubscriptionStatus, &u.SubscriptionPlan, &subscriptionEndsAt, &u.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	u.PhotoPath = photoPath.String
	if subscriptionEndsAt.Valid { u.SubscriptionEndsAt = &subscriptionEndsAt.Time }
	return &u, nil
}

func (r *UserRepository) GetByID(id int64) (*models.User, error) {
	var u models.User
	var photoPath sql.NullString
	var subscriptionEndsAt sql.NullTime
	row := r.db.QueryRow(
		`SELECT id, username, password_hash, display_name, photo_path, account_type, subscription_status, subscription_plan, subscription_ends_at, created_at FROM users WHERE id = ?`,
		id,
	)

	err := row.Scan(&u.ID, &u.Username, &u.PasswordHash, &u.DisplayName, &photoPath, &u.AccountType, &u.SubscriptionStatus, &u.SubscriptionPlan, &subscriptionEndsAt, &u.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	u.PhotoPath = photoPath.String
	if subscriptionEndsAt.Valid { u.SubscriptionEndsAt = &subscriptionEndsAt.Time }
	return &u, nil
}

func (r *UserRepository) UpdatePhotoPath(userID int64, photoPath string) error {
	_, err := r.db.Exec(`UPDATE users SET photo_path = ? WHERE id = ?`, photoPath, userID)
	return err
}

func (r *UserRepository) UpdateDisplayName(userID int64, displayName string) error {
	_, err := r.db.Exec(`UPDATE users SET display_name = ? WHERE id = ?`, displayName, userID)
	return err
}

func (r *UserRepository) UpdatePasswordHash(userID int64, newHash string) error {
	_, err := r.db.Exec(`UPDATE users SET password_hash = ? WHERE id = ?`, newHash, userID)
	return err
}

func (r *UserRepository) HasActiveSellerSubscription(userID int64) (bool, error) {
	var active int
	err := r.db.QueryRow(`SELECT CASE WHEN account_type = 'seller' AND subscription_status = 'active' AND (subscription_ends_at IS NULL OR subscription_ends_at > CURRENT_TIMESTAMP) THEN 1 ELSE 0 END FROM users WHERE id = ?`, userID).Scan(&active)
	return active == 1, err
}
