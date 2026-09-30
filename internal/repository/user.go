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
		`INSERT INTO users (username, email, password_hash, display_name, account_type, subscription_status, subscription_plan) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		user.Username, user.Email, user.PasswordHash, user.DisplayName, user.AccountType, user.SubscriptionStatus, user.SubscriptionPlan,
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
		`SELECT id, username, email, password_hash, display_name, photo_path, account_type, subscription_status, subscription_plan, subscription_ends_at, created_at FROM users WHERE username = ?`,
		username,
	)

	err := row.Scan(&u.ID, &u.Username, &u.Email, &u.PasswordHash, &u.DisplayName, &photoPath, &u.AccountType, &u.SubscriptionStatus, &u.SubscriptionPlan, &subscriptionEndsAt, &u.CreatedAt)
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
		`SELECT id, username, email, password_hash, display_name, photo_path, account_type, subscription_status, subscription_plan, subscription_ends_at, created_at FROM users WHERE id = ?`,
		id,
	)

	err := row.Scan(&u.ID, &u.Username, &u.Email, &u.PasswordHash, &u.DisplayName, &photoPath, &u.AccountType, &u.SubscriptionStatus, &u.SubscriptionPlan, &subscriptionEndsAt, &u.CreatedAt)
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

func (r *UserRepository) UpdateEmail(userID int64, email string) error {
	_, err := r.db.Exec(`UPDATE users SET email = ? WHERE id = ?`, email, userID)
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

func (r *UserRepository) CreateSubscriptionPayment(userID int64, reference, plan string) error {
	_, err := r.db.Exec(`INSERT INTO subscription_payments (user_id, reference, plan, status) VALUES (?, ?, ?, 'pending')`, userID, reference, plan); return err
}
func (r *UserRepository) SubscriptionPaymentByReference(reference string) (int64, string, error) {
	var userID int64
	var plan string
	err := r.db.QueryRow(`SELECT user_id, plan FROM subscription_payments WHERE reference=?`, reference).Scan(&userID, &plan)
	return userID, plan, err
}

func (r *UserRepository) PendingSubscriptionPayment(userID int64, reference string) (string,error) {
	var plan string; err := r.db.QueryRow(`SELECT plan FROM subscription_payments WHERE user_id=? AND reference=? AND status='pending'`,userID,reference).Scan(&plan); return plan,err
}
func (r *UserRepository) ActivateSellerSubscription(userID int64, reference, plan string) error {
	tx, err := r.db.Begin(); if err != nil { return err }; defer tx.Rollback()
	var status string; if err := tx.QueryRow(`SELECT status FROM subscription_payments WHERE user_id=? AND reference=?`,userID,reference).Scan(&status); err != nil { return err }
	if status == "completed" { return tx.Commit() }
	modifier := "+1 month"; if plan == "yearly" { modifier = "+1 year" }
	if _,err=tx.Exec(`UPDATE users SET account_type='seller', subscription_status='active', subscription_plan=?, subscription_ends_at=datetime(CASE WHEN subscription_ends_at > CURRENT_TIMESTAMP THEN subscription_ends_at ELSE CURRENT_TIMESTAMP END, ?) WHERE id=?`,plan,modifier,userID); err != nil{return err}
	if _,err=tx.Exec(`UPDATE subscription_payments SET status='completed', completed_at=CURRENT_TIMESTAMP WHERE user_id=? AND reference=?`,userID,reference);err!=nil{return err}
	return tx.Commit()
}
