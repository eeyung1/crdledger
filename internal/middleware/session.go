package middleware

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"sync"

	"crdledger/internal/repository"
)

// Session stores user session data
type Session struct {
	UserID   int64
	Username string
}

// SessionManager handles user sessions
type SessionManager struct {
	sessions map[string]*Session
	mu       sync.RWMutex
	userRepo *repository.UserRepository
}

// NewSessionManager creates a new session manager
func NewSessionManager(userRepo *repository.UserRepository) *SessionManager {
	return &SessionManager{
		sessions: make(map[string]*Session),
		userRepo: userRepo,
	}
}

// GenerateSessionID creates a random session ID
func GenerateSessionID() string {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return ""
	}
	return hex.EncodeToString(b)
}

// CreateSession creates a new session for the user
func (m *SessionManager) CreateSession(user *repository.User) string {
	sessionID := GenerateSessionID()
	m.sessions[sessionID] = &Session{
		UserID:   user.ID,
		Username: user.Username,
	}
	return sessionID
}

// GetSession retrieves a session by ID
func (m *SessionManager) GetSession(sessionID string) *Session {
	if session, exists := m.sessions[sessionID]; exists {
		return session
	}
	return nil
}

// DeleteSession removes a session
func (m *SessionManager) DeleteSession(sessionID string) {
	delete(m.sessions, sessionID)
}

// RequireAuth is middleware that requires authentication
func (m *SessionManager) RequireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("session_id")
		if err != nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		session := m.GetSession(cookie.Value)
		if session == nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		// Add user info to context for handlers to use
		ctx := r.Context()
		ctx = context.WithValue(ctx, "userID", session.UserID)
		ctx = context.WithValue(ctx, "username", session.Username)
		next(w, r.WithContext(ctx))
	}
}