package handler

import (
	"html/template"
	"net/http"

	"crdledger/internal/middleware"
	"crdledger/internal/service"
)

type TransactionsListHandler struct {
	balances  *service.BalanceService
	templates *template.Template
}

func NewTransactionsListHandler(balances *service.BalanceService, templates *template.Template) *TransactionsListHandler {
	return &TransactionsListHandler{balances: balances, templates: templates}
}

func (h *TransactionsListHandler) render(w http.ResponseWriter, r *http.Request, isSeller bool, title string) {
	userID, ok := middleware.UserIDFromContext(r)
	if !ok {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	balance, err := h.balances.GetBalance(userID)
	if err != nil {
		http.Error(w, "failed to load balance", http.StatusInternalServerError)
		return
	}

	roleFiltered := service.FilterByRole(balance.Transactions, isSeller)

	query := r.URL.Query().Get("q")
	transactions := service.FilterTransactions(roleFiltered, query)

	h.templates.ExecuteTemplate(w, "transactions_list.html", map[string]any{
		"Title":        title,
		"BasePath":     r.URL.Path,
		"Transactions": transactions,
		"Query":        query,
	})
}

func (h *TransactionsListHandler) Creditors(w http.ResponseWriter, r *http.Request) {
	h.render(w, r, true, "Creditors (you're owed money)")
}

func (h *TransactionsListHandler) Debtors(w http.ResponseWriter, r *http.Request) {
	h.render(w, r, false, "Debtors (you owe money)")
}
