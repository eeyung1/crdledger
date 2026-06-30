package handler

import (
	"fmt"
	"net/http"

	"crdledger/internal/middleware"
	"crdledger/internal/service"
)

type AuthHandler struct {
	authService *service.AuthService
	sessionMgr  *middleware.SessionManager
}

func NewAuthHandler(authService *service.AuthService, sessionMgr *middleware.SessionManager) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		sessionMgr:  sessionMgr,
	}
}

func (h *AuthHandler) RegisterPage(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		h.showRegisterForm(w, r)
		return
	}

	if r.Method == "POST" {
		h.handleRegister(w, r)
		return
	}

	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
}

func (h *AuthHandler) showRegisterForm(w http.ResponseWriter, r *http.Request) {
	html := `
<!DOCTYPE html>
<html>
<head>
    <title>CRDLEDGER - Register</title>
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <style>
        body { font-family: Arial, sans-serif; max-width: 400px; margin: 50px auto; padding: 20px; }
        input { width: 100%; padding: 10px; margin: 5px 0 15px 0; border: 1px solid #ddd; border-radius: 4px; }
        button { width: 100%; padding: 10px; background: #4CAF50; color: white; border: none; border-radius: 4px; cursor: pointer; }
        button:hover { background: #45a045a049; }
        .error { color: red; }
        .success { color: green; }
    </style>
</head>
<body>
    <h1>Register for CRDLEDGER</h1>
    <form method="POST" action="/register">
        <label>Username:</label>
        <input type="text" name="username" required>

        <label>Password:</label>
        <input type="password" name="password" required>

        <label>Display Name:</label>
        <input type="text" name="display_name" placeholder="How others will see you" required>

        <button type="submit">Register</button>
    </form>
    <p style="margin-top:  margin-top:;">Already have an account? <a href="/login">Login</a></p>
</body>
</html>
	`
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprint(w, html)
}

func (h *AuthHandler) handleRegister(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	username := r.FormValue("username")
	password := r.FormValue("password")
	displayName := r.FormValue("display_name")

	if err := h.authService.Register(username, password, displayName); err != nil {
		errorHTML := fmt.Sprintf(`
		<!DOCTYPE html>
		<html>
		<head>
			<title>CRDLEDGER - Register</title>
			<meta name="viewport" content="width=device-width, initial-scale=1">
			<style>
				body { font-family: Arial, sans-serif; max-width: 400px; margin: 50px auto; padding: 20px; }
				input { width: 100%; padding: 10px; margin: 5px 0 15px 0; border: 1px solid #ddd; border-radius: 4px; }
				button { width: 100%; padding: 10px; background: #4CAF50; color: white; border: none; border-radius: 4px; cursor: pointer; }
				button:hover { background: #45a049; }
				.error { color: red; }
				.success { color: green; }
			</style>
		</head>
		<body>
			<h1>Register for CRDLEDGER</h1>
			<form method="POST" action="/register">
				<label>Username:</label>
				<input type="text" name="username" required value="%s">

				<label>Password:</label>
				<input type="password" name="password" required>

				<label>Display Name:</label>
				<input type="text" name="display_name" placeholder="How others will see you" required value="%s">

				<button type="submit">Register</button>
			</form>
			<p class="error">%s</p>
			<p style="margin-top: 20px;">Already have an account? <a href="/login">Login</a></p>
		</body>
		</html>
		`, username, displayName, err)
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, errorHTML)
		return
	}

	successHTML := `
	<!DOCTYPE html>
	<html>
	<head>
		<title>CRDLEDGER - Register</title>
		<meta name="viewport" content="width=device-width, initial-scale=1">
		<style>
			body { font-family: Arial, sans-serif; max-width: 400px; margin: 50px auto; padding: 20px; }
			.success { color: green; }
			a { color: #4CAF50; text-decoration: none; }
			a:hover { text-decoration: underline; }
		</style>
	</head>
	<body>
		<h1>Registration Successful!</h1>
		<p class="success">Your account has been created successfully.</p>
		<p><a href="/login">Click here to login</a></p>
	</body>
	</html>
	`
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprint(w, successHTML)
}

func (h *AuthHandler) LoginPage(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		h.showLoginForm(w, r)
		return
	}

	if r.Method == "POST" {
		h.handleLogin(w, r)
		return
	}

	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
}

func (h *AuthHandler) showLoginForm(w http.ResponseWriter, r *http.Request) {
	html := `
<!DOCTYPE html>
<html>
<head>
    <title>CRDLEDGER - Login</title>
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <style>
        body { font-family: Arial, sans-serif; max-width: 400px; margin: 50px auto; padding: 20px; }
        input { width: 100%; padding: 10px; margin: 5px 0 15px 0; border: 1px solid #ddd; border-radius: 4px; }
        button { width: 100%; padding: 10px; background: #2196F3; color: white; border: none; border-radius: 4px; cursor: pointer; }
        button:hover { background: #0b7dda; }
        .error { color: red; }
    </style>
</head>
<body>
    <h1>Login to CRDLEDGER</h1>
    <form method="POST" action="/login">
        <label>Username:</label>
        <input type="text" name="username" required>

        <label>Password:</label>
        <input type="password" name="password" required>

        <button type="submit">Login</button>
    </form>
    <p style="margin-top: 20px;">Don't have an account? <a href="/register">Register</a></p>
</body>
</html>
	`
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprint(w, html)
}

func (h *AuthHandler) handleLogin(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	username := r.FormValue("username")
	password := r.FormValue("password")

	user, err := h.authService.Login(username, password)
	if err != nil {
		errorHTML := fmt.Sprintf(`
		<!DOCTYPE html>
		<html>
		<head>
			<title>CRDLEDGER - Login</title>
			<meta name="viewport" content="width=device-width, initial-scale=1">
			<style>
				body { font-family: Arial, sans-serif; max-width: 400px; margin: 50px auto; padding: 20px; }
				input { width: 100%; padding: 10px; margin: 5px 0 15px 0; border: 1px solid #ddd; border-radius: 4px; }
				button { width: 100%; padding: 10px; background: #2196F3; color: white; border: none; border-radius: 4px; cursor: pointer; }
				button:hover { background: #0b7dda; }
				.error { color: red; }
			</style>
		</head>
		<body>
			<h1>Login to CRDLEDGER</h1>
			<form method="POST" action="/login">
				<label>Username:</label>
				<input type="text" name="username" required value="%s">

				<label>Password:</label>
				<input type="password" name="password" required>

				<button type="submit">Login</button>
			</form>
			<p class="error">%s</p>
			<p style="margin-top: 20px;">Don't have an account? <a href="/register">Register</a></p>
		</body>
		</html>
		`, username, err)
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, errorHTML)
		return
	}

	// Create session
	sessionID := h.sessionMgr.CreateSession(user)
	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    sessionID,
		Path:     "/",
		MaxAge:   3600, // 1 hour
		HttpOnly: true,
	})

	http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session_id")
	if err == nil {
		h.sessionMgr.DeleteSession(cookie.Value)
	}
	http.SetCookie(w, &http.Cookie{
		Name:   "session_id",
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	})
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}