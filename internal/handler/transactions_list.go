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

	csrfToken := middleware.CSRFTokenFromContext(r)
	data := map[string]any{
		"Title":     title,
		"BasePath":  r.URL.Path,
		"Rows":      buildTxRows(transactions, csrfToken),
		"Query":     query,
		"CSRFToken": csrfToken,
	}

	// HTMX-driven search only needs the list fragment re-rendered, not the
	// whole page (sidebar, head, etc). Falls back to a full page for
	// regular navigation / no-JS.
	if isHTMXRequest(r) {
		h.templates.ExecuteTemplate(w, "tx_list_fragment", data)
		return
	}

	h.templates.ExecuteTemplate(w, "transactions_list.html", data)
}

func (h *TransactionsListHandler) Creditors(w http.ResponseWriter, r *http.Request) {
	h.render(w, r, true, "Creditors (you're owed money)")
}

func (h *TransactionsListHandler) Debtors(w http.ResponseWriter, r *http.Request) {
	h.render(w, r, false, "Debtors (you owe money)")
}
