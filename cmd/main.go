package main

import (
	"database/sql"
	"html/template"
	"log"
	"net/http"
	"os"

	"crdledger/internal/handler"
	"crdledger/internal/middleware"
	"crdledger/internal/repository"
	"crdledger/internal/service"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	if _, ok := os.LookupEnv("SESSION_SECRET"); !ok {
		log.Fatal("SESSION_SECRET environment variable is required but not set")
	}

	db, err := sql.Open("sqlite3", "./crdledger.db")
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	if err := createTables(db); err != nil {
		log.Fatalf("failed to create tables: %v", err)
	}

	log.Println("database ready: crdledger.db")

	templates, err := template.ParseGlob("templates/*.html")
	if err != nil {
		log.Fatalf("failed to parse templates: %v", err)
	}

	userRepo := repository.NewUserRepository(db)
	transactionRepo := repository.NewTransactionRepository(db)

	authService := service.NewAuthService(userRepo)
	transactionService := service.NewTransactionService(transactionRepo, userRepo)
	balanceService := service.NewBalanceService(transactionRepo, userRepo)

	sessions := middleware.NewSessionStore()

	authHandler := handler.NewAuthHandler(authService, sessions, templates)
	dashboardHandler := handler.NewDashboardHandler(userRepo, balanceService, templates)
	transactionHandler := handler.NewTransactionHandler(transactionService, templates)

	mux := http.NewServeMux()
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		if _, ok := sessions.CurrentUserID(r); ok {
			http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
			return
		}
		http.Redirect(w, r, "/login", http.StatusSeeOther)
	})
	mux.HandleFunc("/register", authHandler.RegisterPage)
	mux.HandleFunc("/login", authHandler.LoginPage)
	mux.HandleFunc("/logout", authHandler.Logout)
	mux.HandleFunc("/dashboard", sessions.RequireAuth(dashboardHandler.Dashboard))
	transactionsMenuHandler := handler.NewTransactionsMenuHandler(templates)
	mux.HandleFunc("/transactions", sessions.RequireAuth(transactionsMenuHandler.Menu))
	transactionsListHandler := handler.NewTransactionsListHandler(balanceService, templates)
	mux.HandleFunc("/transactions/creditors", sessions.RequireAuth(transactionsListHandler.Creditors))
	mux.HandleFunc("/transactions/debtors", sessions.RequireAuth(transactionsListHandler.Debtors))
	mux.HandleFunc("/transactions/new", sessions.RequireAuth(transactionHandler.RecordPage))
	mux.HandleFunc("/transactions/mark-paid", sessions.RequireAuth(transactionHandler.MarkPaid))
	photoService := service.NewPhotoService(userRepo)
	photoHandler := handler.NewPhotoHandler(photoService, templates)
	profileHandler := handler.NewProfileHandler(userRepo, templates)
	mux.HandleFunc("/profile/edit", sessions.RequireAuth(profileHandler.EditProfilePage))
	mux.HandleFunc("/photo/upload", sessions.RequireAuth(photoHandler.Upload))

	log.Println("listening on :8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}

func createTables(db *sql.DB) error {
	usersTable := `
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT NOT NULL UNIQUE,
		password_hash TEXT NOT NULL,
		display_name TEXT NOT NULL,
		photo_path TEXT,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	);`

	transactionsTable := `
	CREATE TABLE IF NOT EXISTS transactions (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		seller_id INTEGER NOT NULL REFERENCES users(id),
		buyer_id INTEGER NOT NULL REFERENCES users(id),
		amount REAL NOT NULL,
		description TEXT NOT NULL,
		status TEXT NOT NULL DEFAULT 'pending',
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		paid_at DATETIME
	);`

	if _, err := db.Exec(usersTable); err != nil {
		return err
	}
	if _, err := db.Exec(transactionsTable); err != nil {
		return err
	}
	return nil
}
