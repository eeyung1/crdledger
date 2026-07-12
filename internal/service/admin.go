package service

import (
	"errors"

	"crdledger/internal/repository"

	"golang.org/x/crypto/bcrypt"
)

var ErrPasswordTooShort = errors.New("password must be at least 8 characters")

// AdminResetPassword looks up a user by username and overwrites their
// password hash directly — no email/SMS, no reset token, because resets
// are performed in person by a trusted admin, not self-service.
func (s *AuthService) AdminResetPassword(targetUsername, newPassword string) error {
	if len(newPassword) < 8 {
		return ErrPasswordTooShort
	}

	user, err := s.users.GetByUsername(targetUsername)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return err
		}
		return err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	return s.users.UpdatePasswordHash(user.ID, string(hash))
}
