package handler

import (
	"errors"
	"net/http"

	"crdledger/internal/middleware"
	"crdledger/internal/service"
)

func (h *AuthHandler) LoginPage(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		h.templates.ExecuteTemplate(w, "login.html", map[string]any{
			"CSRFToken": middleware.CSRFTokenFromContext(r),
		})
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	username := r.FormValue("username")
	password := r.FormValue("password")

	user, err := h.auth.Login(username, password)
	if err != nil {
		var errMsg string
		if errors.Is(err, service.ErrInvalidCredentials) {
			errMsg = "Invalid username or password."
		} else {
			errMsg = "Something went wrong. Please try again."
		}
		h.templates.ExecuteTemplate(w, "login.html", map[string]any{
			"Error":     errMsg,
			"CSRFToken": middleware.CSRFTokenFromContext(r),
		})
		return
	}

	token, err := h.sessions.Create(user.ID)
	if err != nil {
		http.Error(w, "failed to create session", http.StatusInternalServerError)
		return
	}

	h.sessions.SetCookie(w, token)
	http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	cookie, err := r.Cookie(middleware.SessionCookieName)
	if err == nil {
		h.sessions.Destroy(cookie.Value)
	}

	h.sessions.ClearCookie(w)
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}
