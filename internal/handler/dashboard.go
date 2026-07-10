package handler

import (
	"html/template"
	"net/http"

	"crdledger/internal/middleware"
	"crdledger/internal/repository"
	"crdledger/internal/service"
)

type DashboardHandler struct {
	users     *repository.UserRepository
	balances  *service.BalanceService
	templates *template.Template
}

func NewDashboardHandler(users *repository.UserRepository, balances *service.BalanceService, templates *template.Template) *DashboardHandler {
	return &DashboardHandler{users: users, balances: balances, templates: templates}
}

func (h *DashboardHandler) Dashboard(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r)
	if !ok {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	user, err := h.users.GetByID(userID)
	if err != nil {
		http.Error(w, "failed to load user", http.StatusInternalServerError)
		return
	}

	balance, err := h.balances.GetBalance(userID)
	if err != nil {
		http.Error(w, "failed to load balance", http.StatusInternalServerError)
		return
	}

	h.templates.ExecuteTemplate(w, "dashboard.html", map[string]any{
		"DisplayName":     user.DisplayName,
		"PhotoPath":       user.PhotoPath,
		"PhotoError":      r.URL.Query().Get("photo_error"),
		"TotalReceivable": balance.TotalReceivable,
		"TotalOwed":       balance.TotalOwed,
	})
}
