package handler

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"crdledger/internal/middleware"
	"crdledger/internal/service"
)

func (h *TransactionHandler) MarkPaid(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, ok := middleware.UserIDFromContext(r)
	if !ok {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	idStr := r.FormValue("transaction_id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid transaction id", http.StatusBadRequest)
		return
	}

	dateStr := r.FormValue("paid_date")
	paidAt, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		http.Error(w, "invalid date", http.StatusBadRequest)
		return
	}

	err = h.transactions.MarkPaid(id, userID, paidAt)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrNotSeller):
			http.Error(w, "only the seller can mark this transaction as paid", http.StatusForbidden)
		case errors.Is(err, service.ErrAlreadyPaid):
			http.Error(w, "transaction is already paid", http.StatusBadRequest)
		case errors.Is(err, service.ErrPaidDateInFuture):
			http.Error(w, "payment date cannot be in the future", http.StatusBadRequest)
		case errors.Is(err, service.ErrPaidDateBeforeCreated):
			http.Error(w, "payment date cannot be before the transaction was created", http.StatusBadRequest)
		default:
			http.Error(w, "something went wrong", http.StatusInternalServerError)
		}
		return
	}

	http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
}
