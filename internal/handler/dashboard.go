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

	query := r.URL.Query().Get("q")
	transactions := service.FilterTransactions(balance.Transactions, query)

	h.templates.ExecuteTemplate(w, "dashboard.html", map[string]any{
		"DisplayName":     user.DisplayName,
		"TotalReceivable": balance.TotalReceivable,
		"TotalOwed":       balance.TotalOwed,
		"Transactions":    transactions,
		"Query":           query,
	})
}
