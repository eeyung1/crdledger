package handler

import (
	"html/template"
	"net/http"

	"crdledger/internal/middleware"
	"crdledger/internal/repository"
)

type ProfileHandler struct {
	users     *repository.UserRepository
	templates *template.Template
}

func NewProfileHandler(users *repository.UserRepository, templates *template.Template) *ProfileHandler {
	return &ProfileHandler{users: users, templates: templates}
}

func (h *ProfileHandler) EditProfilePage(w http.ResponseWriter, r *http.Request) {
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

	h.templates.ExecuteTemplate(w, "edit_profile.html", map[string]any{
		"PhotoPath":  user.PhotoPath,
		"PhotoError": r.URL.Query().Get("photo_error"),
		"CSRFToken":  middleware.CSRFTokenFromContext(r),
	})
}
