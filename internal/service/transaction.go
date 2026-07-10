package service

import (
	"errors"

	"crdledger/internal/models"
	"crdledger/internal/repository"
)

var ErrInvalidAmount = errors.New("amount must be positive")
var ErrInvalidDescription = errors.New("description is required")
var ErrBuyerNotFound = errors.New("no user with that username")
var ErrCannotRecordSelf = errors.New("you cannot record a transaction with yourself")

type TransactionService struct {
	transactions *repository.TransactionRepository
	users        *repository.UserRepository
}

func NewTransactionService(transactions *repository.TransactionRepository, users *repository.UserRepository) *TransactionService {
	return &TransactionService{transactions: transactions, users: users}
}

func (s *TransactionService) Record(sellerID int64, buyerUsername string, amount float64, description string) (*models.Transaction, error) {
	if amount <= 0 {
		return nil, ErrInvalidAmount
	}
	if description == "" {
		return nil, ErrInvalidDescription
	}

	buyer, err := s.users.GetByUsername(buyerUsername)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return nil, ErrBuyerNotFound
		}
		return nil, err
	}

	if buyer.ID == sellerID {
		return nil, ErrCannotRecordSelf
	}

	t := &models.Transaction{
		SellerID:    sellerID,
		BuyerID:     buyer.ID,
		Amount:      amount,
		Description: description,
		Status:      "pending",
	}

	if err := s.transactions.Create(t); err != nil {
		return nil, err
	}

	return t, nil
}
