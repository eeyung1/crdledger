package service

import (
	"errors"

	"crdledger/internal/models"
	"crdledger/internal/repository"
)

var ErrInvalidAmount = errors.New("amount must be positive")
var ErrInvalidDescription = errors.New("description is required")
var ErrBuyerNotFound = errors.New("no user with that username")
var ErrSellerNotFound = errors.New("no user with that username")
var ErrCannotRecordSelf = errors.New("you cannot record a transaction with yourself")

type TransactionService struct {
	transactions *repository.TransactionRepository
	users        *repository.UserRepository
}

func NewTransactionService(transactions *repository.TransactionRepository, users *repository.UserRepository) *TransactionService {
	return &TransactionService{transactions: transactions, users: users}
}

func (s *TransactionService) Record(sellerID int64, buyerUsername string, amount float64, description, receiptPath string) (*models.Transaction, error) {
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
		CreatedByID: sellerID,
	}
	if receiptPath != "" {
		t.PhotoPath = &receiptPath
	}

	if err := s.transactions.Create(t); err != nil {
		return nil, err
	}

	return t, nil
}

// RecordOrder lets a buyer self-report an order they placed — the mirror
// image of Record. Here the logged-in user is the buyer, and they name the
// seller. It lands on the seller's Creditors page needing their response,
// exactly the way a seller-recorded debt lands needing the buyer's
// response — same confirmation rules, same balance treatment, just the
// creator and confirmer roles swapped.
func (s *TransactionService) RecordOrder(buyerID int64, sellerUsername string, amount float64, description, receiptPath string) (*models.Transaction, error) {
	if amount <= 0 {
		return nil, ErrInvalidAmount
	}
	if description == "" {
		return nil, ErrInvalidDescription
	}

	seller, err := s.users.GetByUsername(sellerUsername)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return nil, ErrSellerNotFound
		}
		return nil, err
	}

	if seller.ID == buyerID {
		return nil, ErrCannotRecordSelf
	}

	t := &models.Transaction{
		SellerID:    seller.ID,
		BuyerID:     buyerID,
		Amount:      amount,
		Description: description,
		Status:      "pending",
		CreatedByID: buyerID,
	}
	if receiptPath != "" {
		t.PhotoPath = &receiptPath
	}

	if err := s.transactions.Create(t); err != nil {
		return nil, err
	}

	return t, nil
}
