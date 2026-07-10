package service

import (
	"errors"

	"crdledger/internal/models"
	"crdledger/internal/repository"

	"golang.org/x/crypto/bcrypt"
)

var ErrUsernameTaken = errors.New("username already taken")
var ErrInvalidInput = errors.New("username, password, and display name are required")

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
