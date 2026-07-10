package handler

import (
	"errors"
	"html/template"
	"net/http"

	"crdledger/internal/middleware"
	"crdledger/internal/service"
)

type AuthHandler struct {
	auth      *service.AuthService
	sessions  *middleware.SessionStore
	templates *template.Template
}

func NewAuthHandler(auth *service.AuthService, sessions *middleware.SessionStore, templates *template.Template) *AuthHandler {
	return &AuthHandler{auth: auth, sessions: sessions, templates: templates}
}

func (h *AuthHandler) RegisterPage(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		h.templates.ExecuteTemplate(w, "register.html", nil)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	username := r.FormValue("username")
	password := r.FormValue("password")
	displayName := r.FormValue("display_name")

	user, err := h.auth.Register(username, password, displayName)
	if err != nil {
		var errMsg string
		switch {
		case errors.Is(err, service.ErrUsernameTaken):
			errMsg = "That username is already taken."
		case errors.Is(err, service.ErrInvalidInput):
			errMsg = "Please fill in all fields."
		default:
			errMsg = "Something went wrong. Please try again."
		}
		h.templates.ExecuteTemplate(w, "register.html", map[string]string{"Error": errMsg})
		return
	}

	token, err := h.sessions.Create(user.ID)
	if err != nil {
		http.Error(w, "failed to create session", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     middleware.SessionCookieName,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
	})

	http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
}
