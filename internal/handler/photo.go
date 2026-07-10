package handler

import (
	"errors"
	"html/template"
	"net/http"

	"crdledger/internal/middleware"
	"crdledger/internal/service"
)

type PhotoHandler struct {
	photos    *service.PhotoService
	templates *template.Template
}

func NewPhotoHandler(photos *service.PhotoService, templates *template.Template) *PhotoHandler {
	return &PhotoHandler{photos: photos, templates: templates}
}

func (h *PhotoHandler) Upload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, ok := middleware.UserIDFromContext(r)
	if !ok {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	file, header, err := r.FormFile("photo")
	if err != nil {
		http.Redirect(w, r, "/dashboard?photo_error=Please+choose+a+file", http.StatusSeeOther)
		return
	}
	defer file.Close()

	_, err = h.photos.SavePhoto(userID, header.Filename, header.Size, file)
	if err != nil {
		var msg string
		switch {
		case errors.Is(err, service.ErrPhotoTooLarge):
			msg = "Photo+must+be+smaller+than+2MB"
		case errors.Is(err, service.ErrInvalidPhotoType):
			msg = "Photo+must+be+a+JPG,+PNG,+or+GIF"
		default:
			msg = "Something+went+wrong"
		}
		http.Redirect(w, r, "/profile/edit?photo_error="+msg, http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, "/profile/edit", http.StatusSeeOther)
}
