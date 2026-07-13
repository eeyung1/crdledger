package handler

import (
	"errors"
	"html/template"
	"net/http"
	"net/url"

	"crdledger/internal/middleware"
	"crdledger/internal/repository"
	"crdledger/internal/service"
)

type ProfileHandler struct {
	users     *repository.UserRepository
	auth      *service.AuthService
	admin     *AdminChecker
	templates *template.Template
}

func NewProfileHandler(users *repository.UserRepository, auth *service.AuthService, admin *AdminChecker, templates *template.Template) *ProfileHandler {
	return &ProfileHandler{users: users, auth: auth, admin: admin, templates: templates}
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
		"PhotoPath":   user.PhotoPath,
		"PhotoError":  r.URL.Query().Get("photo_error"),
		"DisplayName": user.DisplayName,
		"NameError":   r.URL.Query().Get("name_error"),
		"CSRFToken":   middleware.CSRFTokenFromContext(r),
		"IsAdmin":     h.admin.IsAdmin(r),
	})
}

// UpdateDisplayName handles the "edit profile" display-name form. On
// failure it redirects back with an error in the query string, the same
// pattern the photo upload handler uses, so the page never needs a
// separate error-rendering path.
func (h *ProfileHandler) UpdateDisplayName(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, ok := middleware.UserIDFromContext(r)
	if !ok {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	displayName := r.FormValue("display_name")

	if err := h.auth.UpdateDisplayName(userID, displayName); err != nil {
		msg := "Something+went+wrong"
		if errors.Is(err, service.ErrInvalidDisplayName) {
			msg = url.QueryEscape("Display name can't be empty")
		}
		http.Redirect(w, r, "/profile/edit?name_error="+msg, http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, "/profile/edit", http.StatusSeeOther)
}
