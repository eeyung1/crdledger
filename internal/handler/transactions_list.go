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

// searchThreshold matches the state-inventory rule: search only appears
// once there's enough in the list to need it — on a short list it's
// clutter, not a tool.
const searchThreshold = 5

func (h *TransactionsListHandler) render(w http.ResponseWriter, r *http.Request, isSeller bool, title, emptyMessage string) {
	userID, ok := middleware.UserIDFromContext(r)
	if !ok {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	csrfToken := middleware.CSRFTokenFromContext(r)
	basePath := r.URL.Path

	balance, err := h.balances.GetBalance(userID)
	if err != nil {
		data := map[string]any{
			"Title":     title,
			"BasePath":  basePath,
			"LoadError": true,
			"CSRFToken": csrfToken,
		}
		if isHTMXRequest(r) {
			h.templates.ExecuteTemplate(w, "tx_list_fragment", data)
			return
		}
		h.templates.ExecuteTemplate(w, "transactions_list.html", data)
		return
	}

	roleFiltered := service.FilterByRole(balance.Transactions, isSeller)

	query := r.URL.Query().Get("q")
	transactions := service.FilterTransactions(roleFiltered, query)

	data := map[string]any{
		"Title":        title,
		"BasePath":     basePath,
		"Rows":         buildTxRows(transactions, csrfToken),
		"Query":        query,
		"CSRFToken":    csrfToken,
		"ShowSearch":   len(roleFiltered) > searchThreshold,
		"EmptyMessage": emptyMessage,
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
	h.render(w, r, true, "Creditors", "You're not owed anything right now.")
}

func (h *TransactionsListHandler) Debtors(w http.ResponseWriter, r *http.Request) {
	h.render(w, r, false, "Debtors", "You don't owe anyone right now.")
}
