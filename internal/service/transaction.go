package service

import (
	"errors"
	"fmt"

	"crdledger/internal/repository"
)

type TransactionService struct {
	transactionRepo *repository.TransactionRepository
	userRepo        *repository.UserRepository
}

func NewTransactionService(transactionRepo *repository.TransactionRepository, userRepo *repository.UserRepository) *TransactionService {
	return &TransactionService{
		transactionRepo: transactionRepo,
		userRepo:        userRepo,
	}
}

func (s *TransactionService) CreateTransaction(sellerUsername, buyerUsername string, amount int64, description string) error {
	// Get seller user
	seller, err := s.userRepo.GetUserByUsername(sellerUsername)
	if err != nil {
		return fmt.Errorf("failed to get seller: %w", err)
	}
	if seller == nil {
		return errors.New("seller not found")
	}

	// Get buyer user
	buyer, err := s.userRepo.GetUserByUsername(buyerUsername)
	if err != nil {
		return fmt.Errorf("failed to get buyer: %w", err)
	}
	if buyer == nil {
		return errors.New("buyer not found")
	}

	// Prevent user from transacting with themselves
	if seller.ID == buyer.ID {
		return errors.New("cannot create transaction with yourself")
	}

	// Create transaction
	_, err = s.transactionRepo.CreateTransaction(seller.ID, buyer.ID, amount, description)
	if err != nil {
		return fmt.Errorf("failed to create transaction: %w", err)
	}

	return nil
}

func (s *TransactionService) GetTransactionsByUser(userID int64) ([]repository.Transaction, error) {
	return s.transactionRepo.GetTransactionsByUser(userID)
}

func (s *TransactionService) GetPendingTransactionsBySeller(sellerID int64) ([]repository.Transaction, error) {
	return s.transactionRepo.GetPendingTransactionsBySeller(sellerID)
}

func (s *TransactionService) GetPendingTransactionsByBuyer(buyerID int64) ([]repository.Transaction, error) {
	return s.transactionRepo.GetPendingTransactionsByBuyer(buyerID)
}

func (s *TransactionService) MarkTransactionAsPaid(transactionID int64) error {
	// First get the transaction to verify it exists
	transaction, err := s.transactionRepo.GetTransactionByID(transactionID)
	if err != nil {
		return fmt.Errorf("failed to get transaction: %w", err)
	}
	if transaction == nil {
		return errors.New("transaction not found")
	}

	// Mark as paid
	return s.transactionRepo.MarkTransactionAsPaid(transactionID)
}

func (s *TransactionService) GetTransactionByID(transactionID int64) (*repository.Transaction, error) {
	return s.transactionRepo.GetTransactionByID(transactionID)
}