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
	photos       *service.PhotoService
	admin        *AdminChecker
	templates    *template.Template
}

func NewTransactionHandler(transactions *service.TransactionService, balances *service.BalanceService, photos *service.PhotoService, admin *AdminChecker, templates *template.Template) *TransactionHandler {
	return &TransactionHandler{transactions: transactions, balances: balances, photos: photos, admin: admin, templates: templates}
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
	IsAdmin       bool
}

func (h *TransactionHandler) RecordPage(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		h.templates.ExecuteTemplate(w, "record.html", RecordFormData{
			CSRFToken: middleware.CSRFTokenFromContext(r),
			IsAdmin:   h.admin.IsAdmin(r),
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
			IsAdmin:       h.admin.IsAdmin(r),
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

	// The receipt photo is entirely optional — a missing file is not an
	// error, only a genuinely invalid one is.
	receiptPath := ""
	if file, header, err := r.FormFile("receipt"); err == nil {
		defer file.Close()
		path, saveErr := h.photos.SaveReceipt(sellerID, header.Filename, header.Size, file)
		if saveErr != nil {
			switch {
			case errors.Is(saveErr, service.ErrPhotoTooLarge):
				respondError("That receipt photo is too big — please use one under 2MB.")
			case errors.Is(saveErr, service.ErrInvalidPhotoType):
				respondError("Receipt photos must be a JPG, PNG, or GIF.")
			default:
				respondError("Couldn't save that receipt photo. Please try again.")
			}
			return
		}
		receiptPath = path
	}

	_, err := h.transactions.Record(sellerID, buyerUsername, amount, description, receiptPath)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidAmount):
			respondError("Amount must be positive.")
		case errors.Is(err, service.ErrInvalidDescription):
			respondError("Please enter a description.")
		case errors.Is(err, service.ErrBuyerNotFound):
			respondError("We couldn't find that username — check the spelling and try again.")
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
