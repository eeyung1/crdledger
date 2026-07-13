package handler

import (
	"errors"
	"net/http"
	"strconv"

	"crdledger/internal/middleware"
	"crdledger/internal/service"
)

// Confirm handles a buyer accepting a pending transaction as real.
func (h *TransactionHandler) Confirm(w http.ResponseWriter, r *http.Request) {
	h.respondToTransaction(w, r, h.transactions.Confirm)
}

// Reject handles a buyer disputing a pending transaction. The record is
// kept (marked "rejected"), not deleted, so there's a dispute history.
func (h *TransactionHandler) Reject(w http.ResponseWriter, r *http.Request) {
	h.respondToTransaction(w, r, h.transactions.Reject)
}

// respondToTransaction is the shared plumbing behind Confirm/Reject: parse
// the id, resolve the requesting user, run the given action, then render
// the same way MarkPaid does — a refreshed row fragment for HTMX, or a
// plain redirect for a no-JS form post.
func (h *TransactionHandler) respondToTransaction(w http.ResponseWriter, r *http.Request, action func(transactionID, requestingUserID int64) error) {
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

	actionErr := action(id, userID)

	if isHTMXRequest(r) {
		h.renderRowFragment(w, r, id, userID, actionErr)
		return
	}

	if actionErr != nil {
		httpErrorForConfirmation(w, actionErr)
		return
	}

	back := r.Referer()
	if back == "" {
		back = "/dashboard"
	}
	http.Redirect(w, r, back, http.StatusSeeOther)
}

func httpErrorForConfirmation(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrNotBuyer):
		http.Error(w, "only the buyer can confirm or reject this transaction", http.StatusForbidden)
	case errors.Is(err, service.ErrAlreadyResponded):
		http.Error(w, "this transaction has already been confirmed or rejected", http.StatusBadRequest)
	case errors.Is(err, service.ErrTransactionNotFound):
		http.Error(w, "transaction not found", http.StatusNotFound)
	default:
		http.Error(w, "something went wrong", http.StatusInternalServerError)
	}
}
