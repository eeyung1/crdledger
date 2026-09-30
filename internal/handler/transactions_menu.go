package handler

import (
	"html/template"
	"net/http"

	"crdledger/internal/middleware"
)

type TransactionsMenuHandler struct {
	admin     *AdminChecker
	templates *template.Template
}

func NewTransactionsMenuHandler(admin *AdminChecker, templates *template.Template) *TransactionsMenuHandler {
	return &TransactionsMenuHandler{admin: admin, templates: templates}
}

func (h *TransactionsMenuHandler) Menu(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r)
	if !ok {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	h.templates.ExecuteTemplate(w, "transactions_menu.html", map[string]any{
		"UserID": userID,
		"CSRFToken": middleware.CSRFTokenFromContext(r),
		"IsAdmin":   h.admin.IsAdmin(r),
	})
}
