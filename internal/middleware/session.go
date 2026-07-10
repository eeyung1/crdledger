package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"sync"
)

const SessionCookieName = "session_id"

type SessionStore struct {
	mu       sync.Mutex
	sessions map[string]int64 // session token -> user ID
}

func NewSessionStore() *SessionStore {
	return &SessionStore{
		sessions: make(map[string]int64),
	}
}

func (s *SessionStore) Create(userID int64) (string, error) {
	token, err := generateToken()
	if err != nil {
		return "", err
	}

	s.mu.Lock()
	s.sessions[token] = userID
	s.mu.Unlock()

	return token, nil
}

func (s *SessionStore) UserID(token string) (int64, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	userID, ok := s.sessions[token]
	return userID, ok
}

func (s *SessionStore) Destroy(token string) {
	s.mu.Lock()
	delete(s.sessions, token)
	s.mu.Unlock()
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
