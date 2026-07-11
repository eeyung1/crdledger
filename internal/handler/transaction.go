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
	balances     *service.BalanceService
	templates    *template.Template
}

func NewTransactionHandler(transactions *service.TransactionService, balances *service.BalanceService, templates *template.Template) *TransactionHandler {
	return &TransactionHandler{transactions: transactions, balances: balances, templates: templates}
}

// RecordFormData backs the "record_form" partial — the form fields are
// echoed back on error so a failed submission doesn't lose the user's
// input, and Success flips on after an HTMX-driven submission so the
// person can record another transaction without leaving the page.
type RecordFormData struct {
	Error         string
	Success       bool
	CSRFToken     string
	PhotoPath     string
	BuyerUsername string
	Amount        string
	Description   string
}

func (h *TransactionHandler) RecordPage(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		h.templates.ExecuteTemplate(w, "record.html", RecordFormData{
			CSRFToken: middleware.CSRFTokenFromContext(r),
		})
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
	csrfToken := middleware.CSRFTokenFromContext(r)

	respondError := func(msg string) {
		data := RecordFormData{
			Error:         msg,
			CSRFToken:     csrfToken,
			BuyerUsername: buyerUsername,
			Amount:        amountStr,
			Description:   description,
		}
		if isHTMXRequest(r) {
			h.templates.ExecuteTemplate(w, "record_form", data)
			return
		}
		h.templates.ExecuteTemplate(w, "record.html", data)
	}

	amount, parseErr := strconv.ParseFloat(amountStr, 64)
	if parseErr != nil {
		respondError("Amount must be a valid number.")
		return
	}

	_, err := h.transactions.Record(sellerID, buyerUsername, amount, description)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidAmount):
			respondError("Amount must be positive.")
		case errors.Is(err, service.ErrInvalidDescription):
			respondError("Please enter a description.")
		case errors.Is(err, service.ErrBuyerNotFound):
			respondError("No user found with that username.")
		case errors.Is(err, service.ErrCannotRecordSelf):
			respondError("You cannot record a transaction with yourself.")
		default:
			respondError("Something went wrong. Please try again.")
		}
		return
	}

	if isHTMXRequest(r) {
		h.templates.ExecuteTemplate(w, "record_form", RecordFormData{
			Success:   true,
			CSRFToken: csrfToken,
		})
		return
	}

	http.Redirect(w, r, "/dashboard?recorded=1", http.StatusSeeOther)
}
