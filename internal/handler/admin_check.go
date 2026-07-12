package handler

import (
	"net/http"

	"crdledger/internal/middleware"
	"crdledger/internal/repository"
)

// AdminChecker resolves whether the logged-in user on a request is an
// admin. Used by page handlers only to decide whether to show the admin
// nav link — AdminHandler re-checks independently before actually allowing
// a reset, so this is a display convenience, not the security boundary.
type AdminChecker struct {
	users  *repository.UserRepository
	admins map[string]bool
}

func NewAdminChecker(users *repository.UserRepository, admins map[string]bool) *AdminChecker {
	return &AdminChecker{users: users, admins: admins}
}

func (c *AdminChecker) IsAdmin(r *http.Request) bool {
	userID, ok := middleware.UserIDFromContext(r)
	if !ok {
		return false
	}
	user, err := c.users.GetByID(userID)
	if err != nil {
		return false
	}
	return c.admins[user.Username]
}
