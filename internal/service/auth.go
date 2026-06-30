package service

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"

	"crdledger/internal/repository"
)

// GetUserByID returns a user by their ID
func (s *AuthService) GetUserByID(userID int64) (*repository.User, error) {
	return s.userRepo.GetUserByID(userID)
}

// GetUserByUsername returns a user by their username
func (s *AuthService) GetUserByUsername(username string) (*repository.User, error) {
	return s.userRepo.GetUserByUsername(username)
}

type AuthService struct {
	userRepo *repository.UserRepository
}

func NewAuthService(userRepo *repository.UserRepository) *AuthService {
	return &AuthService{userRepo: userRepo}
}

func (s *AuthService) Register(username, password, displayName string) error {
	// Check if user exists
	existing, err := s.userRepo.GetUserByUsername(username)
	if err != nil {
		return fmt.Errorf("failed to check user existence: %w", err)
	}
	if existing != nil {
		return fmt.Errorf("username already taken")
	}

	// Hash password (simple for now - we'll improve later)
	passwordHash := hashPassword(password)

	// Create user
	_, err = s.userRepo.CreateUser(username, passwordHash, displayName)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

func (s *AuthService) Login(username, password string) (*repository.User, error) {
	user, err := s.userRepo.GetUserByUsername(username)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return nil, fmt.Errorf("invalid username or password")
	}

	// Verify password
	if !verifyPassword(user.PasswordHash, password) {
		return nil, fmt.Errorf("invalid username or password")
	}

	return user, nil
}

// Simple password hashing (for MVP - use bcrypt in production)
func hashPassword(password string) string {
	// Simple hash for now - we'll add bcrypt later
	return password
}

func verifyPassword(hashed, plain string) bool {
	return hashed == plain
}

func generateSessionID() string {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		return ""
	}
	return hex.EncodeToString(b)
}