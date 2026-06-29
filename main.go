package main

import (
    "fmt"
    "log"
    "net/http"
    "os"

    "crdledger/internal/handler"
    "crdledger/internal/repository"
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

    // Initialize handlers
    authHandler := handler.NewAuthHandler()

    // Routes
    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        fmt.Fprintf(w, "CRDLEDGER is running!")
    })
    http.HandleFunc("/register", authHandler.RegisterPage)

    fmt.Println("✓ CRDLEDGER database initialized")
    fmt.Println("✓ Database:", dbPath)
    fmt.Println("✓ Server starting on http://localhost:8080")
    fmt.Println("✓ Visit http://localhost:8080/register to create an account")

    if err := http.ListenAndServe(":8080", nil); err != nil {
        log.Fatalf("Failed to start server: %v", err)
    }
}