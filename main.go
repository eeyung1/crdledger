package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"crdledger/internal/handler"
	"crdledger/internal/middleware"
	"crdledger/internal/repository"
	"crdledger/internal/service"
)

func main() {
	// Database setup
	dbPath := os.Getenv("CRDLEDGER_DB_PATH")
	if dbPath == "" {
		dbPath = "./crdledger.db"
	}

	db, err := repository.InitDB(dbPath)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	if err := repository.CreateTables(db); err != nil {
		log.Fatalf("Failed to create tables: %v", err)
	}

	// Initialize repositories
	userRepo := repository.NewUserRepository(db)
	transactionRepo := repository.NewTransactionRepository(db)

	// Initialize services
	authService := service.NewAuthService(userRepo)
	transactionService := service.NewTransactionService(transactionRepo, userRepo)

	// Initialize session manager
	sessionMgr := middleware.NewSessionManager(userRepo)

	// Initialize handlers
	authHandler := handler.NewAuthHandler(authService, sessionMgr)
	dashboardHandler := handler.NewDashboardHandler(authService, transactionService, sessionMgr)

	// Routes
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "CRDLEDGER is running!")
	})
	http.HandleFunc("/register", authHandler.RegisterPage)
	http.HandleFunc("/login", authHandler.LoginPage)
	http.HandleFunc("/logout", authHandler.Logout)

	// Protected routes
	http.HandleFunc("/dashboard", sessionMgr.RequireAuth(dashboardHandler.Dashboard))
	http.HandleFunc("/transaction/new", sessionMgr.RequireAuth(dashboardHandler.NewTransaction))
	http.HandleFunc("/transaction/", sessionMgr.RequireAuth(dashboardHandler.MarkAsPaid))

	fmt.Println("✓ CRDLEDGER database initialized")
	fmt.Println("✓ Database:", dbPath)
	fmt.Println("✓ Server starting on http://localhost:8080")
	fmt.Println("✓ Visit http://localhost:8080/register to create an account")
	fmt.Println("✓ Visit http://localhost:8080/login to login")

	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}