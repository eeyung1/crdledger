package service

import (
	"crypto/rand"
	"encoding/hex"
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

// saveFile validates and writes an uploaded image to static/uploads,
// returning its web-servable path. Shared by profile avatars and
// transaction receipts — the validation rules (size, type) are identical;
// only what happens to the resulting path differs by caller.
func saveFile(storedName string, size int64, filename string, src io.Reader) (string, error) {
	if size > maxPhotoSize {
		return "", ErrPhotoTooLarge
	}

	ext := filepath.Ext(filename)
	if !allowedExtensions[ext] {
		return "", ErrInvalidPhotoType
	}

	destPath := filepath.Join("static", "uploads", storedName+ext)
	dest, err := os.Create(destPath)
	if err != nil {
		return "", err
	}
	defer dest.Close()

	if _, err := io.Copy(dest, src); err != nil {
		return "", err
	}

	return "/static/uploads/" + storedName + ext, nil
}

func (s *PhotoService) SavePhoto(userID int64, filename string, size int64, src io.Reader) (string, error) {
	webPath, err := saveFile(fmt.Sprintf("%d", userID), size, filename, src)
	if err != nil {
		return "", err
	}

	if err := s.users.UpdatePhotoPath(userID, webPath); err != nil {
		return "", err
	}

	return webPath, nil
}

// SaveReceipt stores an optional evidence photo attached to a transaction
// at record time. Unlike avatars, this doesn't update any repository row
// itself — the caller attaches the returned path to the transaction being
// created.
func (s *PhotoService) SaveReceipt(sellerID int64, filename string, size int64, src io.Reader) (string, error) {
	storedName := fmt.Sprintf("receipt-%d-%s", sellerID, randomSuffix())
	return saveFile(storedName, size, filename, src)
}

func randomSuffix() string {
	b := make([]byte, 6)
	if _, err := rand.Read(b); err != nil {
		return "0"
	}
	return hex.EncodeToString(b)
}
