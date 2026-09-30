package handler

import (
	"html/template"
	"net/http"

	"crdledger/internal/middleware"
	"crdledger/internal/repository"
	"time"
)

type TransactionsMenuHandler struct {
	users *repository.UserRepository
	admin     *AdminChecker
	templates *template.Template
}

func NewTransactionsMenuHandler(users *repository.UserRepository, admin *AdminChecker, templates *template.Template) *TransactionsMenuHandler {
	return &TransactionsMenuHandler{users: users, admin: admin, templates: templates}
}

func (h *TransactionsMenuHandler) Menu(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r)
	if !ok {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	user, err := h.users.GetByID(userID)
	if err != nil { http.Error(w, "failed to load user", http.StatusInternalServerError); return }
	sellerActive := user.AccountType == "seller" && user.SubscriptionStatus == "active" && (user.SubscriptionEndsAt == nil || user.SubscriptionEndsAt.After(time.Now()))
	h.templates.ExecuteTemplate(w, "transactions_menu.html", map[string]any{
		"UserID": userID,
		"PhotoPath": user.PhotoPath,
		"IsSeller": user.AccountType == "seller",
		"SellerActive": sellerActive,
		"CSRFToken": middleware.CSRFTokenFromContext(r),
		"IsAdmin":   h.admin.IsAdmin(r),
	})
}
