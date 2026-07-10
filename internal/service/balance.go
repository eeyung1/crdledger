package service

import (
	"time"
	"strings"
	"crdledger/internal/models"
	"crdledger/internal/repository"
)

// TransactionView is a role-aware presentation of a transaction from the
// perspective of the currently logged-in user.
type TransactionView struct {
	ID              int64
	CounterpartName string
	CounterpartUsername string
	Role            string // "You are owed" or "You owe"
	Amount          float64
	Description     string
	Status          string
	IsSeller        bool
	PaidAt          *time.Time
}

type Balance struct {
	TotalReceivable float64 // owed to this user as seller
	TotalOwed       float64 // this user owes as buyer
	Transactions    []TransactionView
}

type BalanceService struct {
	transactions *repository.TransactionRepository
	users        *repository.UserRepository
}

func NewBalanceService(transactions *repository.TransactionRepository, users *repository.UserRepository) *BalanceService {
	return &BalanceService{transactions: transactions, users: users}
}

func (s *BalanceService) GetBalance(userID int64) (*Balance, error) {
	txs, err := s.transactions.GetByUser(userID)
	if err != nil {
		return nil, err
	}

	balance := &Balance{}

	for _, t := range txs {
		view, err := s.toView(t, userID)
		if err != nil {
			return nil, err
		}

		if t.Status == "pending" {
			if t.SellerID == userID {
				balance.TotalReceivable += t.Amount
			} else {
				balance.TotalOwed += t.Amount
			}
		}

		balance.Transactions = append(balance.Transactions, view)
	}

	return balance, nil
}

func (s *BalanceService) toView(t models.Transaction, userID int64) (TransactionView, error) {
	isSeller := t.SellerID == userID

	var counterpartID int64
	var role string
	if isSeller {
		counterpartID = t.BuyerID
		role = "Creditor (you're owed money)"
	} else {
		counterpartID = t.SellerID
		role = "Debtor (you owe money)"
	}

	counterpart, err := s.users.GetByID(counterpartID)
	if err != nil {
		return TransactionView{}, err
	}

	return TransactionView{
		ID:              t.ID,
		CounterpartName: counterpart.DisplayName,
		CounterpartUsername: counterpart.Username,
		Role:            role,
		Amount:          t.Amount,
		Description:     t.Description,
		Status:          t.Status,
		IsSeller:        isSeller,
		PaidAt:          t.PaidAt,
	}, nil
}

func FilterTransactions(views []TransactionView, query string) []TransactionView {
	if query == "" {
		return views
	}

	query = strings.ToLower(query)
	var filtered []TransactionView
	for _, v := range views {
		if strings.Contains(strings.ToLower(v.CounterpartName), query) ||
			strings.Contains(strings.ToLower(v.CounterpartUsername), query) {
			filtered = append(filtered, v)
		}
	}
	return filtered
}

func FilterByRole(views []TransactionView, isSeller bool) []TransactionView {
	var filtered []TransactionView
	for _, v := range views {
		if v.IsSeller == isSeller {
			filtered = append(filtered, v)
		}
	}
	return filtered
}
