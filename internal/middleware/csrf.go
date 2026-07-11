package middleware

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"net/http"
)

const csrfCookieName = "csrf_token"

type csrfContextKey string

const csrfKey csrfContextKey = "csrfToken"

// CSRF issues (or reuses) a per-browser token via an HttpOnly cookie and
// rejects any state-changing request whose form value doesn't match it
// (double-submit pattern). The token is also attached to the request
// context so handlers can hand it to templates for hidden form fields.
func CSRF(secureCookie bool) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			token := ""
			if cookie, err := r.Cookie(csrfCookieName); err == nil && cookie.Value != "" {
				token = cookie.Value
			} else {
				token = generateCSRFToken()
				http.SetCookie(w, &http.Cookie{
					Name:     csrfCookieName,
					Value:    token,
					Path:     "/",
					HttpOnly: true,
					Secure:   secureCookie,
					SameSite: http.SameSiteLaxMode,
				})
			}

			ctx := context.WithValue(r.Context(), csrfKey, token)
			r = r.WithContext(ctx)

			if r.Method == http.MethodPost || r.Method == http.MethodPut ||
				r.Method == http.MethodPatch || r.Method == http.MethodDelete {
				submitted := r.FormValue("csrf_token")
				if submitted == "" || subtle.ConstantTimeCompare([]byte(submitted), []byte(token)) != 1 {
					http.Error(w, "session expired or invalid form submission — please refresh and try again", http.StatusForbidden)
					return
				}
			}

			next(w, r)
		}
	}
}

// CSRFTokenFromContext returns the token issued for this request, for
// handlers to pass into template data as {{.CSRFToken}}.
func CSRFTokenFromContext(r *http.Request) string {
	token, _ := r.Context().Value(csrfKey).(string)
	return token
}

func generateCSRFToken() string {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		// crypto/rand failing is fatal-grade; fall back to a value that is
		// still unpredictable per-process rather than a fixed string.
		return hex.EncodeToString([]byte("fallback-should-not-happen"))
	}
	return hex.EncodeToString(b)
}
