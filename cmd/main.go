package main

import (
	"DebtEase/internal/api"
	"DebtEase/internal/database"
	"DebtEase/internal/handlers"
	"DebtEase/internal/logger"
	"DebtEase/internal/middleware"
	"DebtEase/internal/worker"
	"context"
	"database/sql"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func main() {
	// Load .env
	if err := godotenv.Load(); err != nil {
		logger.Logger.Fatalw("Error loading .env file", "error", err)
	}

	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		logger.Logger.Fatal("DB_URL not set in environment")
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		logger.Logger.Fatal("JWT_SECRET is required in .env")
	}

	platform := os.Getenv("PLATFORM")

	// Open DB
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		logger.Logger.Fatalw("Failed to open database", "error", err)
	}
	defer db.Close()

	if err = db.Ping(); err != nil {
		logger.Logger.Fatalw("Database ping failed", "error", err)
	}

	dbQueries := database.New(db)

	// Initialize config
	cfg := &api.Config{
		DB:        dbQueries,
		Platform:  platform,
		JWTSecret: jwtSecret,
	}

	// Setup router
	mux := http.NewServeMux()

	// Health
	mux.HandleFunc("GET /api/healthz", func(w http.ResponseWriter, r *http.Request) {
		logger.Logger.Infow("Health check", "path", r.URL.Path, "method", r.Method)
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Static files with logging middleware
	fileServer := http.FileServer(http.Dir("."))
	mux.Handle("/app/", middleware.Logging(middleware.MetricsInc(cfg, http.StripPrefix("/app", fileServer))))

	// Admin
	mux.HandleFunc("GET /admin/metrics", handlers.HandleMetrics(cfg))

	// auth
	mux.HandleFunc("POST /api/login", handlers.HandleLogin(cfg))
	mux.HandleFunc("POST /api/register", handlers.HandleCreateUser(cfg))
	mux.HandleFunc("PUT /api/users", handlers.HandleUpdateUser(cfg))

	// dashboard (authenticated)
	getDashboard := handlers.HandleGetDashboard(cfg)
	mux.Handle("GET /api/dashboard", middleware.Logging(middleware.AuthRequired(cfg, getDashboard)))

	// debts (authenticated)
	{
		// POST /api/debts
		createDebt := handlers.HandleCreateDebt(cfg)
		mux.Handle("POST /api/debts", middleware.Logging(middleware.AuthRequired(cfg, createDebt)))

		// GET /api/debts
		listDebts := handlers.HandleListDebts(cfg)
		mux.Handle("GET /api/debts", middleware.Logging(middleware.AuthRequired(cfg, listDebts)))

		// GET /api/debts/{id}
		getDebt := handlers.HandleGetDebt(cfg)
		mux.Handle("GET /api/debts/{id}", middleware.Logging(middleware.AuthRequired(cfg, getDebt)))

		// PATCH /api/debts/{id}/balance
		updateBalance := handlers.HandleUpdateDebtBalance(cfg)
		mux.Handle("PATCH /api/debts/{id}/balance", middleware.Logging(middleware.AuthRequired(cfg, updateBalance)))

		// PATCH /api/debts/{id}/health
		updateHealth := handlers.HandleUpdateDebtHealth(cfg)
		mux.Handle("PATCH /api/debts/{id}/health", middleware.Logging(middleware.AuthRequired(cfg, updateHealth)))

		// POST /api/debts/{id}/overdue
		markOverdue := handlers.HandleMarkDebtOverdue(cfg)
		mux.Handle("POST /api/debts/{id}/overdue", middleware.Logging(middleware.AuthRequired(cfg, markOverdue)))

		// payments
		recordPayment := handlers.HandleRecordPayment(cfg, db)
		mux.Handle("POST /api/debts/{id}/payments", middleware.Logging(middleware.AuthRequired(cfg, recordPayment)))
		listPayments := handlers.HandleListPayments(cfg)
		mux.Handle("GET /api/debts/{id}/payments", middleware.Logging(middleware.AuthRequired(cfg, listPayments)))

		// // DELETE /api/debts/{id}
		// deleteDebt := handlers.HandleDeleteDebt(cfg)
		// mux.Handle("DELETE /api/debts/{id}", middleware.Logging(middleware.AuthRequired(cfg, deleteDebt)))
	}

	// Start workers
	ctx := context.Background()
	worker.StartDailyAccrual(ctx, dbQueries, 24*time.Hour)              // Daily interest accrual for credit cards
	worker.StartCreditCardStatementWorker(ctx, dbQueries, 24*time.Hour) // Credit card statement generation
	worker.StartLoanEMIWorker(ctx, dbQueries, 24*time.Hour)             // Loan EMI processing

	// Start server
	logger.Logger.Infow("Server starting", "port", 8080, "platform", platform)
	logger.Logger.Fatal(http.ListenAndServe(":8080", mux))
}
