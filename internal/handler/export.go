package handler

import (
	"encoding/csv"
	"fmt"
	"net/http"

	"crdledger/internal/middleware"
	"crdledger/internal/service"
)

type ExportHandler struct {
	balances *service.BalanceService
}

func NewExportHandler(balances *service.BalanceService) *ExportHandler {
	return &ExportHandler{balances: balances}
}

// TransactionsCSV streams the user's full transaction history — every
// role, every counterparty, every status — as a CSV download. This is the
// honest answer to "what if I need this outside the app": no second
// export product, just the data that's already there, in a format a
// spreadsheet can open.
func (h *ExportHandler) TransactionsCSV(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r)
	if !ok {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	balance, err := h.balances.GetBalance(userID)
	if err != nil {
		http.Error(w, "couldn't load your transactions", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", `attachment; filename="crdledger-transactions.csv"`)

	cw := csv.NewWriter(w)
	cw.Write([]string{"date", "role", "counterparty", "username", "amount", "description", "status", "paid_date", "has_receipt"})

	for _, t := range balance.Transactions {
		role := "you are owed"
		if !t.IsSeller {
			role = "you owe"
		}
		paidDate := ""
		if t.PaidAt != nil {
			paidDate = t.PaidAt.Format("2006-01-02")
		}
		hasReceipt := "no"
		if t.PhotoPath != "" {
			hasReceipt = "yes"
		}
		cw.Write([]string{
			t.CreatedAt.Format("2006-01-02"),
			role,
			t.CounterpartName,
			t.CounterpartUsername,
			fmt.Sprintf("%.2f", t.Amount),
			t.Description,
			t.Status,
			paidDate,
			hasReceipt,
		})
	}

	cw.Flush()
}
