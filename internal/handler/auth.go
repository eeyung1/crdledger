package handler

import (
    "fmt"
    "net/http"
)

type AuthHandler struct {
    // We'll add service later
}

func NewAuthHandler() *AuthHandler {
    return &AuthHandler{}
}

func (h *AuthHandler) RegisterPage(w http.ResponseWriter, r *http.Request) {
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
        button:hover { background: #45a049; }
        .error { color: red; }
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
    <p style="margin-top: 20px;">Already have an account? <a href="/login">Login</a></p>
</body>
</html>
`
    w.Header().Set("Content-Type", "text/html")
    fmt.Fprint(w, html)
}