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

	data := map[string]any{
		"DisplayName": user.DisplayName,
		"PhotoPath":   user.PhotoPath,
		"PhotoError":  r.URL.Query().Get("photo_error"),
		"CSRFToken":   middleware.CSRFTokenFromContext(r),
	}

	// The dashboard totals are the single most trust-sensitive numbers in
	// the app — on a load failure we show a retry banner, never a
	// fabricated "0" that could be misread as "you're settled up".
	balance, err := h.balances.GetBalance(userID)
	if err != nil {
		data["LoadError"] = true
	} else {
		data["TotalReceivable"] = balance.TotalReceivable
		data["TotalOwed"] = balance.TotalOwed
		data["NetPosition"] = balance.NetPosition
		data["IsEmpty"] = len(balance.Transactions) == 0
		data["TopCreditors"] = balance.TopCreditors
		data["TopDebtors"] = balance.TopDebtors
		data["CreditorBarData"] = NetBarChartData{Rows: balance.TopCreditors, Variant: "positive"}
		data["DebtorBarData"] = NetBarChartData{Rows: balance.TopDebtors, Variant: "attention"}
		data["Sparkline"] = balance.Sparkline
		// A chart earns its place only once there's more than one or two
		// numbers to compare — with a single counterparty the totals
		// above already say everything the bar would.
		data["ShowCreditorChart"] = len(balance.TopCreditors) > 1
		data["ShowDebtorChart"] = len(balance.TopDebtors) > 1
	}

	h.templates.ExecuteTemplate(w, "dashboard.html", data)
}
