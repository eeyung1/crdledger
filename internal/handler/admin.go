package handler

import (
	"errors"
	"fmt"
	"html/template"
	"net/http"

	"crdledger/internal/middleware"
	"crdledger/internal/repository"
	"crdledger/internal/service"
)

type AdminHandler struct {
	auth      *service.AuthService
	checker   *AdminChecker
	templates *template.Template
}

func NewAdminHandler(auth *service.AuthService, checker *AdminChecker, templates *template.Template) *AdminHandler {
	return &AdminHandler{auth: auth, checker: checker, templates: templates}
}

func (h *AdminHandler) isAdmin(r *http.Request) bool {
	return h.checker.IsAdmin(r)
}

func (h *AdminHandler) ResetPasswordPage(w http.ResponseWriter, r *http.Request) {
	if !h.isAdmin(r) {
		http.Error(w, "not authorized", http.StatusForbidden)
		return
	}

	csrfToken := middleware.CSRFTokenFromContext(r)

	if r.Method == http.MethodGet {
		h.templates.ExecuteTemplate(w, "admin_reset_password.html", map[string]any{"CSRFToken": csrfToken})
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	targetUsername := r.FormValue("username")
	newPassword := r.FormValue("new_password")

	err := h.auth.AdminResetPassword(targetUsername, newPassword)
	if err != nil {
		var errMsg string
		switch {
		case errors.Is(err, repository.ErrUserNotFound):
			errMsg = "No user found with that username."
		case errors.Is(err, service.ErrPasswordTooShort):
			errMsg = "Password must be at least 8 characters."
		default:
			errMsg = "Something went wrong. Please try again."
		}
		h.templates.ExecuteTemplate(w, "admin_reset_password.html", map[string]any{
			"Error":     errMsg,
			"CSRFToken": csrfToken,
			"Username":  targetUsername,
		})
		return
	}

	h.templates.ExecuteTemplate(w, "admin_reset_password.html", map[string]any{
		"Success":   fmt.Sprintf("Password for %s has been reset.", targetUsername),
		"CSRFToken": csrfToken,
	})
}
