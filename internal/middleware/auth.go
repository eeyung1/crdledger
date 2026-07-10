package middleware

import (
	"context"
	"net/http"
)

type contextKey string

const userIDKey contextKey = "userID"

// RequireAuth checks for a valid session and, if present, attaches the user
// ID to the request context before calling next. If there's no valid
// session, it redirects to /login instead of calling next.
func (s *SessionStore) RequireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := s.CurrentUserID(r)
		if !ok {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		ctx := context.WithValue(r.Context(), userIDKey, userID)
		next(w, r.WithContext(ctx))
	}
}

// UserIDFromContext retrieves the user ID attached by RequireAuth.
func UserIDFromContext(r *http.Request) (int64, bool) {
	userID, ok := r.Context().Value(userIDKey).(int64)
	return userID, ok
}
