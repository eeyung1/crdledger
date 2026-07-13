package service

import (
	"errors"
	"strings"

	"crdledger/internal/models"
	"crdledger/internal/repository"

	"golang.org/x/crypto/bcrypt"
)

var ErrUsernameTaken = errors.New("username already taken")
var ErrInvalidInput = errors.New("username, password, and display name are required")
var ErrInvalidDisplayName = errors.New("display name is required")

type AuthService struct {
	users *repository.UserRepository
}

func NewAuthService(users *repository.UserRepository) *AuthService {
	return &AuthService{users: users}
}

func (s *AuthService) Register(username, password, displayName string) (*models.User, error) {
	if username == "" || password == "" || displayName == "" {
		return nil, ErrInvalidInput
	}

	_, err := s.users.GetByUsername(username)
	if err == nil {
		return nil, ErrUsernameTaken
	}
	if !errors.Is(err, repository.ErrUserNotFound) {
		return nil, err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &models.User{
		Username:     username,
		PasswordHash: string(hash),
		DisplayName:  displayName,
	}

	if err := s.users.Create(user); err != nil {
		return nil, err
	}

	return user, nil
}

// UpdateDisplayName changes how a user shows up to others across the app
// (transaction rows, dashboard greeting, etc). It's purely cosmetic — no
// uniqueness constraint like username — so the only rule is that it can't
// be blank.
func (s *AuthService) UpdateDisplayName(userID int64, displayName string) error {
	displayName = strings.TrimSpace(displayName)
	if displayName == "" {
		return ErrInvalidDisplayName
	}
	return s.users.UpdateDisplayName(userID, displayName)
}
