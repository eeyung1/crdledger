package handler

import (
	"errors"
	"html/template"
	"net/http"
	"strconv"

	"crdledger/internal/middleware"
	"crdledger/internal/service"
)

type OrderHandler struct {
	transactions *service.TransactionService
	photos       *service.PhotoService
	admin        *AdminChecker
	templates    *template.Template
}

func NewOrderHandler(transactions *service.TransactionService, photos *service.PhotoService, admin *AdminChecker, templates *template.Template) *OrderHandler {
	return &OrderHandler{transactions: transactions, photos: photos, admin: admin, templates: templates}
}

// OrderFormData backs the "order_form" partial — the mirror image of
// RecordFormData. Here the logged-in user is the buyer, so the field that
// changes is which username they're providing: the seller's.
type OrderFormData struct {
	Error          string
	Success        bool
	CSRFToken      string
	PhotoPath      string
	SellerUsername string
	Amount         string
	Description    string
	IsAdmin        bool
}

// NewOrderPage handles "Add Orders" — a buyer self-reporting something
// they bought, which then needs the seller's Accept/Reject on their
// Creditors page, exactly the way a seller-recorded debt needs the
// buyer's response today.
func (h *OrderHandler) NewOrderPage(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		h.templates.ExecuteTemplate(w, "add_order.html", OrderFormData{
			CSRFToken: middleware.CSRFTokenFromContext(r),
			IsAdmin:   h.admin.IsAdmin(r),
		})
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	buyerID, ok := middleware.UserIDFromContext(r)
	if !ok {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	sellerUsername := r.FormValue("seller_username")
	description := r.FormValue("description")
	amountStr := r.FormValue("amount")
	csrfToken := middleware.CSRFTokenFromContext(r)

	respondError := func(msg string) {
		data := OrderFormData{
			Error:          msg,
			CSRFToken:      csrfToken,
			SellerUsername: sellerUsername,
			Amount:         amountStr,
			Description:    description,
			IsAdmin:        h.admin.IsAdmin(r),
		}
		if isHTMXRequest(r) {
			h.templates.ExecuteTemplate(w, "order_form", data)
			return
		}
		h.templates.ExecuteTemplate(w, "add_order.html", data)
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
		path, saveErr := h.photos.SaveReceipt(buyerID, header.Filename, header.Size, file)
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

	_, err := h.transactions.RecordOrder(buyerID, sellerUsername, amount, description, receiptPath)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidAmount):
			respondError("Amount must be positive.")
		case errors.Is(err, service.ErrInvalidDescription):
			respondError("Please enter a description.")
		case errors.Is(err, service.ErrSellerNotFound):
			respondError("We couldn't find that username — check the spelling and try again.")
		case errors.Is(err, service.ErrCannotRecordSelf):
			respondError("You cannot record a transaction with yourself.")
		default:
			respondError("Something went wrong. Please try again.")
		}
		return
	}

	if isHTMXRequest(r) {
		h.templates.ExecuteTemplate(w, "order_form", OrderFormData{
			Success:   true,
			CSRFToken: csrfToken,
		})
		return
	}

	http.Redirect(w, r, "/dashboard?ordered=1", http.StatusSeeOther)
}
