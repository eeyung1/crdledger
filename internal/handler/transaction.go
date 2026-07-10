package handler

import (
	"errors"
	"html/template"
	"net/http"
	"strconv"

	"crdledger/internal/middleware"
	"crdledger/internal/service"
)

type TransactionHandler struct {
	transactions *service.TransactionService
	templates    *template.Template
}

func NewTransactionHandler(transactions *service.TransactionService, templates *template.Template) *TransactionHandler {
	return &TransactionHandler{transactions: transactions, templates: templates}
}

func (h *TransactionHandler) RecordPage(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		h.templates.ExecuteTemplate(w, "record.html", nil)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	sellerID, ok := middleware.UserIDFromContext(r)
	if !ok {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	buyerUsername := r.FormValue("buyer_username")
	description := r.FormValue("description")
	amountStr := r.FormValue("amount")

	amount, parseErr := strconv.ParseFloat(amountStr, 64)
	if parseErr != nil {
		h.templates.ExecuteTemplate(w, "record.html", map[string]string{"Error": "Amount must be a valid number."})
		return
	}

	_, err := h.transactions.Record(sellerID, buyerUsername, amount, description)
	if err != nil {
		var errMsg string
		switch {
		case errors.Is(err, service.ErrInvalidAmount):
			errMsg = "Amount must be positive."
		case errors.Is(err, service.ErrInvalidDescription):
			errMsg = "Please enter a description."
		case errors.Is(err, service.ErrBuyerNotFound):
			errMsg = "No user found with that username."
		case errors.Is(err, service.ErrCannotRecordSelf):
			errMsg = "You cannot record a transaction with yourself."
		default:
			errMsg = "Something went wrong. Please try again."
		}
		h.templates.ExecuteTemplate(w, "record.html", map[string]string{"Error": errMsg})
		return
	}

	http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
}
