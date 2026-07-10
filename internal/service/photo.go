package service

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"crdledger/internal/repository"
)

var ErrPhotoTooLarge = errors.New("photo must be smaller than 2MB")
var ErrInvalidPhotoType = errors.New("photo must be a JPG, PNG, or GIF")

const maxPhotoSize = 2 * 1024 * 1024 // 2MB

var allowedExtensions = map[string]bool{
	".jpg":  true,
	".jpeg": true,
	".png":  true,
	".gif":  true,
}

type PhotoService struct {
	users *repository.UserRepository
}

func NewPhotoService(users *repository.UserRepository) *PhotoService {
	return &PhotoService{users: users}
}

func (s *PhotoService) SavePhoto(userID int64, filename string, size int64, src io.Reader) (string, error) {
	if size > maxPhotoSize {
		return "", ErrPhotoTooLarge
	}

	ext := filepath.Ext(filename)
	if !allowedExtensions[ext] {
		return "", ErrInvalidPhotoType
	}

	storedName := fmt.Sprintf("%d%s", userID, ext)
	destPath := filepath.Join("static", "uploads", storedName)

	dest, err := os.Create(destPath)
	if err != nil {
		return "", err
	}
	defer dest.Close()

	if _, err := io.Copy(dest, src); err != nil {
		return "", err
	}

	webPath := "/static/uploads/" + storedName

	if err := s.users.UpdatePhotoPath(userID, webPath); err != nil {
		return "", err
	}

	return webPath, nil
}
