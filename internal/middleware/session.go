package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"sync"
	"time"
)

const SessionCookieName = "session_id"

// SessionTTL controls both how long a session cookie lives in the browser
// and how long the server honors it.
const SessionTTL = 30 * 24 * time.Hour

type sessionEntry struct {
	userID    int64
	expiresAt time.Time
}

type SessionStore struct {
	mu       sync.Mutex
	sessions map[string]sessionEntry
	// SecureCookies should be true whenever the app is served over HTTPS.
	SecureCookies bool
}

func NewSessionStore() *SessionStore {
	s := &SessionStore{
		sessions: make(map[string]sessionEntry),
	}
	go s.expireLoop()
	return s
}

func (s *SessionStore) expireLoop() {
	ticker := time.NewTicker(1 * time.Hour)
	for range ticker.C {
		s.mu.Lock()
		now := time.Now()
		for token, entry := range s.sessions {
			if now.After(entry.expiresAt) {
				delete(s.sessions, token)
			}
		}
		s.mu.Unlock()
	}
}

func (s *SessionStore) Create(userID int64) (string, error) {
	token, err := generateToken()
	if err != nil {
		return "", err
	}

	s.mu.Lock()
	s.sessions[token] = sessionEntry{userID: userID, expiresAt: time.Now().Add(SessionTTL)}
	s.mu.Unlock()

	return token, nil
}

func (s *SessionStore) UserID(token string) (int64, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	entry, ok := s.sessions[token]
	if !ok || time.Now().After(entry.expiresAt) {
		return 0, false
	}
	return entry.userID, true
}

func (s *SessionStore) Destroy(token string) {
	s.mu.Lock()
	delete(s.sessions, token)
	s.mu.Unlock()
}

// SetCookie writes the session cookie with the flags appropriate for the
// current environment (Secure only when served over HTTPS).
func (s *SessionStore) SetCookie(w http.ResponseWriter, token string) {
	http.SetCookie(w, &http.Cookie{
		Name:     SessionCookieName,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   s.SecureCookies,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   int(SessionTTL.Seconds()),
	})
}

// ClearCookie logs the browser out of the session cookie.
func (s *SessionStore) ClearCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     SessionCookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   s.SecureCookies,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
	})
}

func generateToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// CurrentUserID reads the session cookie from the request and returns the
// associated user ID, if any.
func (s *SessionStore) CurrentUserID(r *http.Request) (int64, bool) {
	cookie, err := r.Cookie(SessionCookieName)
	if err != nil {
		return 0, false
	}
	return s.UserID(cookie.Value)
}
