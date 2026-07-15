package main

import (
	"context"
	"database/sql"
	"errors"
	"html/template"
	"log"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"crdledger/internal/config"
	"crdledger/internal/handler"
	"crdledger/internal/middleware"
	"crdledger/internal/repository"
	"crdledger/internal/service"

	_ "github.com/tursodatabase/libsql-client-go/libsql"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	dsn := cfg.TursoURL + "?authToken=" + url.QueryEscape(cfg.TursoAuthToken)
	db, err := sql.Open("libsql", dsn)
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		log.Fatalf("failed to enable foreign keys: %v", err)
	}

	if err := createTables(db); err != nil {
		log.Fatalf("failed to create tables: %v", err)
	}

	slog.Info("database ready", "url", cfg.TursoURL)

	templates, err := template.ParseGlob("templates/*.html")
	if err != nil {
		log.Fatalf("failed to parse templates: %v", err)
	}

	userRepo := repository.NewUserRepository(db)
	transactionRepo := repository.NewTransactionRepository(db)

	authService := service.NewAuthService(userRepo)
	transactionService := service.NewTransactionService(transactionRepo, userRepo)
	balanceService := service.NewBalanceService(transactionRepo, userRepo)
	photoService := service.NewPhotoService(userRepo)

	sessions := middleware.NewSessionStore()
	sessions.SecureCookies = cfg.SecureCookies

	adminChecker := handler.NewAdminChecker(userRepo, cfg.AdminUsernames)

	authHandler := handler.NewAuthHandler(authService, sessions, templates)
	dashboardHandler := handler.NewDashboardHandler(userRepo, balanceService, adminChecker, templates)
	transactionHandler := handler.NewTransactionHandler(transactionService, balanceService, photoService, adminChecker, templates)
	orderHandler := handler.NewOrderHandler(transactionService, photoService, adminChecker, templates)
	transactionsMenuHandler := handler.NewTransactionsMenuHandler(adminChecker, templates)
	transactionsListHandler := handler.NewTransactionsListHandler(balanceService, adminChecker, templates)
	photoHandler := handler.NewPhotoHandler(photoService, templates)
	profileHandler := handler.NewProfileHandler(userRepo, authService, adminChecker, templates)
	exportHandler := handler.NewExportHandler(balanceService)
	adminHandler := handler.NewAdminHandler(authService, adminChecker, templates)

	csrf := middleware.CSRF(cfg.SecureCookies)
	authLimiter := middleware.NewRateLimiter(10, time.Minute)

	mux := http.NewServeMux()

	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	mux.HandleFunc("/service-worker.js", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/javascript")
		w.Header().Set("Service-Worker-Allowed", "/")
		w.Header().Set("Cache-Control", "no-cache")
		http.ServeFile(w, r, "static/service-worker.js")
	})
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		if err := db.Ping(); err != nil {
			http.Error(w, "db unavailable", http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			w.WriteHeader(http.StatusNotFound)
			templates.ExecuteTemplate(w, "not_found.html", nil)
			return
		}
		if _, ok := sessions.CurrentUserID(r); ok {
			http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
			return
		}
		http.Redirect(w, r, "/login", http.StatusSeeOther)
	})

	mux.HandleFunc("/register", csrf(authLimiter.Limit(authHandler.RegisterPage)))
	mux.HandleFunc("/login", csrf(authLimiter.Limit(authHandler.LoginPage)))
	mux.HandleFunc("/logout", csrf(authHandler.Logout))
	mux.HandleFunc("/dashboard", csrf(sessions.RequireAuth(dashboardHandler.Dashboard)))
	mux.HandleFunc("/transactions", csrf(sessions.RequireAuth(transactionsMenuHandler.Menu)))
	mux.HandleFunc("/transactions/creditors", csrf(sessions.RequireAuth(transactionsListHandler.Creditors)))
	mux.HandleFunc("/transactions/debtors", csrf(sessions.RequireAuth(transactionsListHandler.Debtors)))
	mux.HandleFunc("/transactions/new", csrf(sessions.RequireAuth(transactionHandler.RecordPage)))
	mux.HandleFunc("/orders/new", csrf(sessions.RequireAuth(orderHandler.NewOrderPage)))
	mux.HandleFunc("/transactions/mark-paid", csrf(sessions.RequireAuth(transactionHandler.MarkPaid)))
	mux.HandleFunc("/transactions/confirm", csrf(sessions.RequireAuth(transactionHandler.Confirm)))
	mux.HandleFunc("/transactions/reject", csrf(sessions.RequireAuth(transactionHandler.Reject)))
	mux.HandleFunc("/profile/edit", csrf(sessions.RequireAuth(profileHandler.EditProfilePage)))
	mux.HandleFunc("/profile/display-name", csrf(sessions.RequireAuth(profileHandler.UpdateDisplayName)))
	mux.HandleFunc("/photo/upload", csrf(sessions.RequireAuth(photoHandler.Upload)))
	mux.HandleFunc("/transactions/export.csv", csrf(sessions.RequireAuth(exportHandler.TransactionsCSV)))
	mux.HandleFunc("/admin/reset-password", csrf(sessions.RequireAuth(adminHandler.ResetPasswordPage)))

	var root http.Handler = mux
	root = middleware.SecurityHeaders(root)
	if cfg.SecureCookies {
		root = middleware.HSTS(root)
	}
	root = middleware.RequestLog(root)
	root = middleware.Recover(root)

	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           root,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	go func() {
		slog.Info("listening", "port", cfg.Port, "environment", cfg.Environment)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("server failed: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	slog.Info("shutting down")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("graceful shutdown failed", "error", err)
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
		paid_at DATETIME,
		photo_path TEXT
	);`

	indexes := []string{
		`CREATE INDEX IF NOT EXISTS idx_transactions_seller ON transactions(seller_id);`,
		`CREATE INDEX IF NOT EXISTS idx_transactions_buyer ON transactions(buyer_id);`,
	}

	if _, err := db.Exec(usersTable); err != nil {
		return err
	}
	if _, err := db.Exec(transactionsTable); err != nil {
		return err
	}

	// Defensive migration for databases created before receipts existed.
	// SQLite has no "ADD COLUMN IF NOT EXISTS", so we attempt it and
	// ignore the specific "duplicate column" failure.
	if _, err := db.Exec(`ALTER TABLE transactions ADD COLUMN photo_path TEXT`); err != nil {
		if !strings.Contains(err.Error(), "duplicate column") {
			return err
		}
	}

	if _, err := db.Exec(`ALTER TABLE transactions ADD COLUMN amount_paid REAL NOT NULL DEFAULT 0`); err != nil {
		if !strings.Contains(err.Error(), "duplicate column") {
			return err
		}
	}

	// confirmation_status tracks whether the buyer has accepted, rejected, or
	// not yet responded to a recorded transaction — separate from the
	// existing "status" column, which tracks payment state (pending/paid).
	// Existing rows default to "confirmed" so transactions recorded before
	// this feature existed aren't retroactively put in dispute.
	if _, err := db.Exec(`ALTER TABLE transactions ADD COLUMN confirmation_status TEXT NOT NULL DEFAULT 'confirmed'`); err != nil {
		if !strings.Contains(err.Error(), "duplicate column") {
			return err
		}
	}

	// created_by tracks which side of the transaction actually submitted the
	// entry — the seller recording a debt someone owes them, or a buyer
	// self-reporting an order they placed. It's what "who must confirm
	// this" is based on: whoever did NOT create it. Existing rows backfill
	// to seller_id, since seller-recorded entries are all there was before
	// buyer-initiated orders existed.
	if _, err := db.Exec(`ALTER TABLE transactions ADD COLUMN created_by INTEGER`); err != nil {
		if !strings.Contains(err.Error(), "duplicate column") {
			return err
		}
	}
	if _, err := db.Exec(`UPDATE transactions SET created_by = seller_id WHERE created_by IS NULL`); err != nil {
		return err
	}

	for _, idx := range indexes {
		if _, err := db.Exec(idx); err != nil {
			return err
		}
	}
	return nil
}
