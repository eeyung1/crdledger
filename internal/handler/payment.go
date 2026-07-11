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

	markErr := h.transactions.MarkPaid(id, userID, paidAt)

	if isHTMXRequest(r) {
		h.renderRowFragment(w, r, id, userID, markErr)
		return
	}

	if markErr != nil {
		httpErrorForMarkPaid(w, markErr)
		return
	}

	back := r.Referer()
	if back == "" {
		back = "/transactions"
	}
	http.Redirect(w, r, back, http.StatusSeeOther)
}

// renderRowFragment re-renders a single tx_row after a mark-paid attempt,
// whether it succeeded or not, so HTMX can swap it in place without a
// full page reload. On failure the row still reflects its real (unchanged)
// state, with an inline error message.
func (h *TransactionHandler) renderRowFragment(w http.ResponseWriter, r *http.Request, transactionID, userID int64, markErr error) {
	view, err := h.balances.GetTransactionView(transactionID, userID)
	if err != nil {
		http.Error(w, "transaction not found", http.StatusNotFound)
		return
	}

	errMsg := ""
	if markErr != nil {
		switch {
		case errors.Is(markErr, service.ErrNotSeller):
			errMsg = "Only the seller can mark this as paid."
		case errors.Is(markErr, service.ErrAlreadyPaid):
			errMsg = "Already marked as paid."
		case errors.Is(markErr, service.ErrPaidDateInFuture):
			errMsg = "Payment date can't be in the future."
		case errors.Is(markErr, service.ErrPaidDateBeforeCreated):
			errMsg = "Payment date can't be before the transaction was created."
		default:
			errMsg = "Something went wrong — please try again."
		}
	}

	h.templates.ExecuteTemplate(w, "tx_row", TxRowData{
		Tx:        view,
		CSRFToken: middleware.CSRFTokenFromContext(r),
		Error:     errMsg,
	})
}

func httpErrorForMarkPaid(w http.ResponseWriter, err error) {
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
}
