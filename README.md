# DebtEase Backend - Go

A comprehensive financial debt management and analysis platform built with Go. DebtEase helps users track, analyze, and strategize their debts with advanced financial modeling, stress testing, and personalized repayment strategies.

**Language Composition:**
- Go: 77.2% (Core backend logic)
- Python: 11.6% (ML/Data analysis, experiments)
- Jupyter Notebook: 10.9% (Analysis & experimentation)

---

## 📋 Table of Contents

1. [Project Overview](#project-overview)
2. [Architecture](#architecture)
3. [Tech Stack](#tech-stack)
4. [Database Schema](#database-schema)
5. [API Endpoints](#api-endpoints)
6. [Setup & Installation](#setup--installation)
7. [Running the Application](#running-the-application)
8. [Development Guide](#development-guide)
9. [Directory Structure](#directory-structure)

---

## 🎯 Project Overview

DebtEase is a financial wellness platform that enables users to:

- **Track Multiple Debts**: Manage credit cards, personal loans, student loans, mortgages, and other debt types
- **Analyze Financial Health**: View debt health scores, interest calculations, and financial snapshots
- **Strategic Planning**: Generate baseline and strategic repayment plans
- **Stress Testing**: Simulate loan scenarios and financial outcomes
- **Payment Management**: Record and track payment history
- **User Dashboard**: Real-time financial overview with key metrics
- **Admin Controls**: Analytics, user management, feedback collection

### Key Features

| Feature | Description |
|---------|------------|
| **Multi-Debt Support** | Track various debt types with type-specific fields (credit cards, loans, mortgages) |
| **Interest Accrual** | Automatic daily interest calculation for credit cards |
| **Financial Snapshots** | Periodic financial health measurements (income, expenses, debt obligations) |
| **Repayment Strategies** | Baseline and custom strategy calculation engines |
| **Payment Recording** | Track payment history per debt with automatic balance updates |
| **Stress Testing** | Simulate loan scenarios to project financial outcomes |
| **Google OAuth** | Seamless authentication via Google OAuth 2.0 |
| **Admin Dashboard** | Analytics, user search, session tracking, feedback management |
| **Event Tracking** | User interaction logging and financial event recording |

---

## 🏗️ Architecture

### High-Level System Design

```
┌─────────────────────────────────────────────────────────────┐
│                      Client Applications                     │
│              (Web App, Mobile, Third-party)                  │
└────────────────────┬────────────────────────────────────────┘
                     │
                     ▼
┌─────────────────────────────────────────────────────────────┐
│                    API Gateway / Router                      │
│                   (Go net/http Mux)                          │
├─────────────────────────────────────────────────────────────┤
│  ┌─────────────┐   ┌──────────────┐   ┌────────────────┐   │
│  │  Middleware │   │  Handlers    │   │  Services      │   │
│  ├─────────────┤   ├──────────────┤   ├────────────────┤   │
│  │ Logging     │   │ Auth         │   │ Debt Logic     │   │
│  │ JWT Auth    │   │ Debts        │   │ Calculations   │   │
│  │ CORS        │   │ Payments     │   │ Stress Tests   │   │
│  │ Admin Check │   │ Dashboard    │   │ Strategies     │   │
│  └─────────────┘   │ Admin        │   └────────────────┘   │
│                    │ Events       │                         │
│                    │ Snapshots    │                         │
│                    └──────────────┘                         │
├─────────────────────────────────────────────────────────────┤
│                    Domain Models                            │
│        (Business logic, calculations, rules)                │
└────────────────┬────────────────────────────────────────────┘
                 │
                 ▼
┌─────────────────────────────────────────────────────────────┐
│                  Database Layer (sqlc)                      │
│    ┌──────────────────────────────────────────────────┐    │
│    │  Generated Type-Safe Database Queries            │    │
│    └──────────────────────────────────────────────────┘    │
└────────────────┬────────────────────────────────────────────┘
                 │
                 ▼
┌─────────────────────────────────────────────────────────────┐
│              PostgreSQL Database (pgx)                      │
│    ┌──────────────────────────────────────────────────┐    │
│    │ Users │ Debts │ Payments │ Financial Snapshots  │    │
│    │ Events │ Feedback │ Sessions │ Obligations       │    │
│    └──────────────────────────────────────────────────┘    │
└─────────────────────────────────────────────────────────────┘

Optional Components:
┌──────────────────────────┐
│  RabbitMQ (PubSub)       │  ← Notifications (Currently disabled)
└──────────────────────────┘
```

### Request Flow Diagram

```
HTTP Request
     │
     ▼
┌─────────────────────┐
│  Router (Mux)       │  Route to handler
└──────┬──────────────┘
       │
       ▼
┌─────────────────────┐
│  Logging MW         │  Log request details
└──────┬──────────────┘
       │
       ▼
┌─────────────────────┐
│  AuthRequired MW    │  Verify JWT token
│  (if protected)     │  Extract user_id from claims
└──────┬──────────────┘
       │
       ▼
┌─────────────────────┐
│  AdminAuth MW       │  Check email in allowlist
│  (if admin route)   │
└──────┬──────────────┘
       │
       ▼
┌─────────────────────┐
│  Handler            │  Parse request body
│  (e.g., debts.go)   │  Call service/domain logic
└──────┬──────────────┘
       │
       ▼
┌─────────────────────┐
│  Domain/Service     │  Execute business rules
│  (calculations,     │  Perform validations
│   stress tests)     │
└──────┬──────────────┘
       │
       ▼
┌─────────────────────┐
│  Database Queries   │  Fetch/update from PostgreSQL
│  (generated sqlc)   │  Type-safe operations
└──────┬──────────────┘
       │
       ▼
┌─────────────────────┐
│  Response Builder   │  Format response JSON
└──────┬──────────────┘
       │
       ▼
   HTTP Response
```

### Data Flow for Debt Calculation

```
User Input (Debts + Financial Info)
         │
         ▼
┌─────────────────────────┐
│  Baseline Calculator    │  Calculate minimum obligations,
│  (webapp_specifics)     │  total debt, available income
└──────┬──────────────────┘
       │
       ▼
┌─────────────────────────┐
│  Financial Snapshot     │  Aggregate debts + income
│  Generator              │  Store periodic health check
└──────┬──────────────────┘
       │
       ▼
┌─────────────────────────┐
│  Strategy Calculator    │  Snowball, Avalanche, or
│  (webapp_specifics)     │  custom payoff plans
└──────┬──────────────────┘
       │
       ▼
┌─────────────────────────┐
│  Stress Test Engine     │  Model various scenarios
│  (stress module)        │  (income loss, rate changes)
└──────┬──────────────────┘
       │
       ▼
   Projected Outcomes
   (Debt-free date, interest saved, etc.)
```

---

## 🛠️ Tech Stack

| Component | Technology | Purpose |
|-----------|-----------|---------|
| **Language** | Go 1.25.0 | Fast, type-safe backend |
| **Database** | PostgreSQL 16 | Relational data persistence |
| **Database Driver** | pgx/v5 | High-performance PostgreSQL client |
| **SQL Generation** | sqlc | Type-safe SQL query codegen |
| **Migrations** | goose | Database schema versioning |
| **Auth** | JWT (golang-jwt/v5) | Token-based authentication |
| **Password Hashing** | argon2id | Secure password hashing |
| **OAuth 2.0** | golang.org/x/oauth2 | Google OAuth integration |
| **Decimal Math** | shopspring/decimal | Precise financial calculations |
| **Logging** | go.uber.org/zap | Structured JSON logging |
| **UUID Generation** | google/uuid | Unique identifiers |
| **Env Management** | joho/godotenv | Configuration from .env |
| **Message Queue** | RabbitMQ (AMQP091-go) | Event publishing (optional) |
| **Email Service** | Resend API | Email notifications (optional) |
| **HTTP** | Go net/http + http.ServeMux | Lightweight routing |
| **Containerization** | Docker & Docker Compose | Local development environment |

### Key Dependencies

```go
// Core Dependencies (go.mod highlights)
github.com/golang-jwt/jwt/v5 v5.3.0
github.com/jackc/pgx/v5 v5.10.0
github.com/shopspring/decimal v1.4.0
go.uber.org/zap v1.27.0
golang.org/x/oauth2 v0.36.0
github.com/google/uuid v1.6.0
github.com/alexedwards/argon2id v1.0.0
```

---

## 📊 Database Schema

### Entity-Relationship Diagram

```
┌──────────────────┐
│     users        │
├──────────────────┤
│ id (PK)          │
│ email (UNIQUE)   │
│ password_hash    │
│ monthly_income   │
│ first_name       │
│ last_name        │
│ google_oauth_id  │
│ is_active        │
│ created_at       │
│ updated_at       │
└────────┬─────────┘
         │ (1:N)
         │
         ▼
┌──────────────────────┐         ┌──────────────────┐
│      debts           │◄────────┤  debt_obligations│
├──────────────────────┤         ├──────────────────┤
│ id (PK)              │         │ debt_id (FK)     │
│ user_id (FK)         │         │ type             │
│ name                 │         │ amount           │
│ type (ENUM)          │         │ frequency        │
│ principal_decimal    │         │ next_due_date    │
│ outstanding_balance  │         │ status           │
│ interest_rate        │         │ created_at       │
│ min_payment          │         └──────────────────┘
│ due_date             │
│ is_overdue           │
│ accrued_interest     │
│ monthly_obligation   │
│ minimum_obligation   │
│ fragmentation_weight │
│ created_at           │
│ updated_at           │
└────────┬─────────────┘
         │ (1:N)
         │
         ▼
┌──────────────────────┐
│  payment_history     │
├──────────────────────┤
│ id (PK)              │
│ debt_id (FK)         │
│ amount               │
│ payment_date         │
│ created_at           │
│ updated_at           │
└──────────────────────┘


┌──────────────────────┐
│financial_snapshots   │
├──────────────────────┤
│ id (PK)              │
│ user_id (FK)         │
│ snapshot_date        │
│ monthly_income       │
│ other_income         │
│ fixed_expenses       │
│ variable_expenses    │
│ savings_cash         │
│ investments_value    │
│ total_monthly_emi    │
│ total_debt_balance   │
│ created_at           │
└──────────────────────┘


┌──────────────────────┐
│  financial_events    │
├──────────────────────┤
│ id (PK)              │
│ user_id (FK)         │
│ event_type           │
│ description          │
│ amount               │
│ event_date           │
│ created_at           │
└──────────────────────┘


┌──────────────────────┐
│   analysis_feedback  │
├──────────────────────┤
│ id (PK)              │
│ email (nullable)     │
│ content              │
│ created_at           │
│ user_agent           │
│ ip_address           │
└──────────────────────┘


┌──────────────────────┐
│  calculation_sessions│
├──────────────────────┤
│ id (PK)              │
│ session_id (FK)      │
│ created_at           │
└──────────────────────┘


┌──────────────────────┐
│   finance_steps      │
├──────────────────────┤
│ id (PK)              │
│ session_id (FK)      │
│ monthly_income       │
│ fixed_obligations    │
│ created_at           │
└──────────────────────┘
```

### Core Tables

#### **users**
User accounts with authentication credentials and financial profile.

```sql
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    first_name VARCHAR(255),
    last_name VARCHAR(255),
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    monthly_income DECIMAL(12,2) DEFAULT 0,
    google_oauth_id VARCHAR(255) UNIQUE,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now()
);
```

#### **debts**
Stores all debts with calculated fields for interest and obligations.

```sql
CREATE TABLE debts (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id),
    name VARCHAR(255) NOT NULL,
    type debt_type NOT NULL,  -- 'credit_card', 'personal_loan', etc.
    principal_decimal DECIMAL(12,2) NOT NULL,
    outstanding_balance DECIMAL(12,2) NOT NULL,
    interest_rate DECIMAL(6,5) NOT NULL,
    accrued_interest DECIMAL(12,2),
    monthly_obligation DECIMAL(12,2),
    minimum_obligation DECIMAL(12,2),
    fragmentation_weight DECIMAL(12,2),
    due_date DATE NOT NULL,
    is_overdue BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ
);
```

#### **payment_history**
Records all payments made against debts.

```sql
CREATE TABLE payment_history (
    id UUID PRIMARY KEY,
    debt_id UUID NOT NULL REFERENCES debts(id),
    amount DECIMAL(12,2) NOT NULL,
    payment_date TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now()
);
```

#### **financial_snapshots**
Periodic financial health measurements aggregating user's complete financial picture.

```sql
CREATE TABLE financial_snapshots (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id),
    snapshot_date DATE NOT NULL,
    monthly_income DECIMAL(12,2),
    other_income DECIMAL(12,2),
    fixed_expenses DECIMAL(12,2),
    variable_expenses DECIMAL(12,2),
    total_monthly_emi DECIMAL(12,2),
    total_debt_balance DECIMAL(12,2),
    savings_cash DECIMAL(12,2),
    investments_value DECIMAL(12,2),
    created_at TIMESTAMPTZ DEFAULT now()
);
```

---

## 🔌 API Endpoints

### Authentication

```
POST   /api/login                    - Login with email & password
POST   /api/register                 - Create new account
POST   /api/refresh                  - Refresh JWT token
PUT    /api/users                    - Update user profile
POST   /api/login/google             - Google OAuth ID token login
GET    /api/auth/google              - Start Google OAuth flow
GET    /api/auth/google/callback     - Google OAuth callback
```

### Dashboard & Overview

```
GET    /api/dashboard                - User dashboard (debts summary, metrics)
GET    /api/users/me/financial-events - List financial events
GET    /api/users/me/financial-snapshots/recent - Recent snapshots
GET    /api/payments/recent          - Recent payments across all debts
```

### Debt Management

```
POST   /api/debts                    - Create new debt
GET    /api/debts                    - List all user debts
GET    /api/debts/{id}               - Get debt details
PATCH  /api/debts/{id}               - Update debt
PATCH  /api/debts/{id}/balance       - Update outstanding balance
PATCH  /api/debts/{id}/health        - Update debt health score
POST   /api/debts/{id}/overdue       - Mark debt as overdue
```

### Payments

```
POST   /api/debts/{id}/payments      - Record payment for debt
GET    /api/debts/{id}/payments      - List payments for debt
GET    /api/payments/recent          - Recent payments across debts
```

### Financial Analysis (Webapp)

```
POST   /api/v1/calculate/baseline    - Calculate baseline financial metrics
POST   /api/v1/calculate/strategy    - Generate repayment strategies
POST   /api/v1/snapshot/attach       - Attach financial snapshot to session
POST   /api/v1/snapshot/analyze      - Analyze financial snapshot
POST   /api/v1/snapshot/simulate-loan - Simulate loan scenario
```

### Events & Feedback

```
POST   /api/events                   - Create financial event
POST   /api/feedback                 - Submit analysis feedback
POST   /api/pdf-request              - Request PDF generation
POST   /api/waitlist                 - Join waitlist
```

### Admin (Requires JWT + Email Allowlist)

```
GET    /api/admin/me                 - Get authenticated admin info
GET    /api/admin/analytics          - System analytics & metrics
GET    /api/admin/overview           - Admin overview dashboard
GET    /api/admin/feedback           - List all feedback
GET    /api/admin/feedback/summary   - Feedback summary
GET    /api/admin/users              - Search users
GET    /api/admin/users/{id}         - Get user details
GET    /api/admin/sessions/{session_id} - Get session details
```

### Health Check

```
GET    /api/healthz                  - Server health check
```

---

## 🚀 Setup & Installation

### Prerequisites

- **Go**: 1.25.0 or higher
- **PostgreSQL**: 16 or higher
- **Docker & Docker Compose**: For containerized dev environment (optional)
- **goose**: Database migration tool (`go install github.com/pressly/goose/v3/cmd/goose@latest`)
- **sqlc**: SQL code generation (`go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest`)

### 1. Clone the Repository

```bash
git clone https://github.com/MihirsinhChauhan/debtease-backend-go.git
cd debtease-backend-go
```

### 2. Install Go Dependencies

```bash
go mod download
go mod tidy
```

### 3. Setup Environment Variables

Create a `.env` file in the root directory:

```bash
# Database
DB_URL=postgres://postgres:postgres@localhost:5432/debtease?sslmode=disable

# JWT
JWT_SECRET=your-super-secret-jwt-key-min-32-chars-long

# Server
PORT=8080
PLATFORM=web  # or 'mobile', 'webapp', etc.

# Google OAuth (Optional)
GOOGLE_CLIENT_ID=your-google-client-id
GOOGLE_CLIENT_SECRET=your-google-client-secret
GOOGLE_REDIRECT_URL=http://localhost:8080/api/auth/google/callback
GOOGLE_LOGIN_SUCCESS_REDIRECT=http://localhost:3000/dashboard

# Admin
ADMIN_EMAILS=admin@example.com,another-admin@example.com

# RabbitMQ (Optional)
RABBITMQ_URL=amqp://guest:guest@localhost:5672/
```

### 4. Start PostgreSQL

#### Option A: Docker Compose (Recommended)

```bash
docker compose up -d postgres
# Wait for PostgreSQL to be ready
sleep 5
```

#### Option B: Local PostgreSQL

Ensure PostgreSQL is running on port 5432 with credentials:
- User: `postgres`
- Password: `postgres`
- Database: `debtease`

### 5. Run Database Migrations

```bash
# Install goose if not already installed
go install github.com/pressly/goose/v3/cmd/goose@latest

# Run migrations
goose -dir database/schema postgres "$(DB_URL)" up

# Verify (list applied migrations)
goose -dir database/schema postgres "$(DB_URL)" status
```

### 6. Generate sqlc Code (Optional, if queries changed)

```bash
# Install sqlc if not already installed
go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest

# Generate database code
sqlc generate
```

---

## ▶️ Running the Application

### Option 1: Run Directly

```bash
# Run the server
go run cmd/main.go

# Expected output:
# {"level":"info","msg":"Database connection successful"}
# {"level":"info","msg":"Server starting","port":"8080","platform":"web"}
```

### Option 2: Build and Run

```bash
# Build executable
go build -o debtease-backend cmd/main.go

# Run the built binary
./debtease-backend
```

### Option 3: Using Make

```bash
# Start database
make up

# Run finance_steps migration
make migrate-finance-steps

# Run all migrations
make migrate

# Run integration tests
make test-integration

# One-shot: Start DB, migrate, and run tests
make integration
```

### Verify Server is Running

```bash
# Health check
curl http://localhost:8080/api/healthz
# Response: OK

# Test login endpoint
curl -X POST http://localhost:8080/api/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"test123"}'
```

---

## 💻 Development Guide

### Project Structure

```
debtease-backend-go/
├── cmd/
│   ├── main.go                 # Entry point, route setup
│   └── main_test.go            # Integration tests
│
├── internal/
│   ├── api/
│   │   ├── config.go           # App configuration
│   │   └── routes.go           # Route definitions
│   │
│   ├── auth/                   # Authentication logic
│   │   ├── jwt.go
│   │   └── password.go
│   │
│   ├── database/               # Generated sqlc code
│   │   ├── models.go
│   │   ├── querier.go
│   │   └── *.sql.go            # Auto-generated
│   │
│   ├── dbconn/                 # Database connection
│   │   └── conn.go
│   │
│   ├── handlers/               # HTTP request handlers
│   │   ├── auth.go
│   │   ├── debts.go
│   │   ├── payments.go
│   │   ├── dashboard.go
│   │   ├── users.go
│   │   ├── google_oauth.go
│   │   ├── financial_events.go
│   │   ├── session.go
│   │   └── snapshot_*.go
│   │
│   ├── middleware/             # HTTP middleware
│   │   ├── logging.go
│   │   ├── auth.go
│   │   ├── admin_auth.go
│   │   └── metrics.go
│   │
│   ├── domain/                 # Business logic & domain models
│   │   ├── debt.go
│   │   ├── user.go
│   │   └── payment.go
│   │
│   ├── models/                 # Data structures
│   │   ├── debt.go
│   │   ├── user.go
│   │   └── response.go
│   │
│   ├── logger/                 # Structured logging (zap)
│   │   └── logger.go
│   │
│   ├── projections/            # Financial calculations
│   │   ├── calculator.go
│   │   └── strategies.go
│   │
│   ├── stress/                 # Stress testing engine
│   │   └── stress_test.go
│   │
│   ├─�� notifications/          # Email & notifications
│   │   └── notifier.go
│   │
│   ├── pubsub/                 # RabbitMQ integration
│   │   └── connection.go
│   │
│   ├── ledger/                 # Financial ledger logic
│   │   └── ledger.go
│   │
│   ├── utils/                  # Helper utilities
│   │   ├── validators.go
│   │   ├── converters.go
│   │   └── errors.go
│   │
│   ├── events/                 # Event definitions
│   │   └── events.go
│   │
│   ├── worker/                 # Background workers
│   │   ├── accrual.go
│   │   ├── statement.go
│   │   └── emi.go
│   │
│   └── webapp_specifics/       # Webapp-specific logic
│       ├── handlers/
│       │   ├── admin.go
│       │   ├── calculation.go
│       │   ├── snapshot.go
│       │   └── feedback.go
│       └── models/
│
├── database/
│   ├── schema/                 # SQL migration files
│   │   ├── 20251104004413_create_users.sql
│   │   ├── 20251104131449_create_debts.sql
│   │   ├── 20251104180213_create_payment_history.sql
│   │   ├── 20251105152530_create_credit_card.sql
│   │   ├── 20260308120000_create_finance_steps.sql
│   │   ├── 20260315192052_create_financial_snapshots_table.sql
│   │   ├── 20260519200000_debt_obligation_phase1.sql
│   │   └── ... (more migrations)
│   │
│   └── queries/                # SQL queries for sqlc
│       ├── users.sql
│       ├── debts.sql
│       ├── payments.sql
│       ├── snapshots.sql
│       └── ... (more queries)
│
├── scripts/
│   ├── migrate_finance_steps.go # Custom migration script
│   └── ... (helper scripts)
│
├── experiments/                # Python ML experiments
│   ├── analysis.ipynb
│   └── ... (Jupyter notebooks)
│
├── docs/                       # Documentation
│   ├── API.md
│   └── ARCHITECTURE.md
│
├── postman/                    # Postman collection
│   └── DebtEase.postman_collection.json
│
├── sample data/                # Sample datasets
│   └── ... (CSV, JSON files)
│
├── Makefile                    # Build automation
├── docker-compose.yml          # Local dev environment
├── go.mod & go.sum             # Go dependencies
├── sqlc.yaml                   # sqlc configuration
├── render.yaml                 # Render.com deployment config
└── .env.example                # Example environment variables
```

### Key Implementation Details

#### 1. **Authentication Flow**

```go
// JWT Token Generation (handlers/auth.go)
token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
    "user_id": userID,
    "email":   user.Email,
    "exp":     time.Now().Add(24 * time.Hour).Unix(),
})

// Middleware Validation (middleware/auth.go)
token, _ := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
    return []byte(cfg.JWTSecret), nil
})
userID := token.Claims["user_id"]
```

#### 2. **Debt Calculation Engine**

```go
// Interest Calculation
AccruedInterest = OutstandingBalance * (InterestRate / 365)

// Monthly Obligation
MonthlyObligation = Min(MinPayment, OutstandingBalance + AccruedInterest)

// Fragmentation Weight (for priority)
FragmentationWeight = (OutstandingBalance / TotalDebt) * Priority
```

#### 3. **Database Queries with sqlc**

```sql
-- database/queries/debts.sql
-- name: GetDebtsByUserID :many
SELECT * FROM debts WHERE user_id = $1 ORDER BY due_date;

-- name: UpdateDebtBalance :exec
UPDATE debts SET outstanding_balance = $1 WHERE id = $2;
```

Generated Go code:
```go
// internal/database/debts.sql.go (auto-generated)
func (q *Queries) GetDebtsByUserID(ctx context.Context, userID string) ([]Debt, error) {
    // ... implementation
}
```

#### 4. **Error Handling Pattern**

```go
if err != nil {
    logger.Logger.Errorw("Failed to create debt", "error", err, "user_id", userID)
    http.Error(w, "Failed to create debt", http.StatusInternalServerError)
    return
}
```

#### 5. **Middleware Chaining**

```go
// Protect authenticated endpoints
mux.Handle("GET /api/debts", 
    middleware.Logging(
        middleware.AuthRequired(cfg, 
            handlers.HandleListDebts(cfg))))

// Protect admin endpoints
mux.Handle("GET /api/admin/users",
    middleware.Logging(
        middleware.AdminAuth(cfg, 
            webapp_handlers.HandleAdminUserSearch(cfg))))
```

### Adding a New Endpoint

#### Step 1: Define SQL Query

Create `database/queries/new_feature.sql`:
```sql
-- name: GetNewData :many
SELECT * FROM some_table WHERE user_id = $1;
```

#### Step 2: Generate Code

```bash
sqlc generate
```

#### Step 3: Create Handler

Create `internal/handlers/new_feature.go`:
```go
package handlers

func HandleNewFeature(cfg *api.Config) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        userID := r.Context().Value("user_id").(string)
        
        results, err := cfg.DB.GetNewData(r.Context(), userID)
        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }
        
        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(results)
    })
}
```

#### Step 4: Register Route

In `cmd/main.go`:
```go
newFeature := handlers.HandleNewFeature(cfg)
mux.Handle("GET /api/new-feature", 
    middleware.Logging(middleware.AuthRequired(cfg, newFeature)))
```

---

## 🔄 Common Development Tasks

### Run Tests

```bash
# Unit tests
go test ./internal/...

# Integration tests (requires DB_URL set)
DB_URL="postgres://postgres:postgres@localhost:5432/debtease?sslmode=disable" \
go test ./cmd -tags=integration -v

# Test a specific package
go test ./internal/handlers -v
```

### Format Code

```bash
# Format all Go files
go fmt ./...

# Check code style
go vet ./...
```

### View Database

```bash
# Connect to PostgreSQL
psql postgres://postgres:postgres@localhost:5432/debtease

# List tables
\dt

# Describe table
\d debts

# Run queries
SELECT * FROM users LIMIT 5;
```

### Add Migration

```bash
# Create new migration file
touch database/schema/$(date +%Y%m%d%H%M%S)_add_new_column.sql

# Edit and add goose directives:
-- +goose Up
ALTER TABLE debts ADD COLUMN new_field VARCHAR(255);

-- +goose Down
ALTER TABLE debts DROP COLUMN new_field;

# Apply migration
goose -dir database/schema postgres "$(DB_URL)" up
```

### Debug with Logging

The application uses `go.uber.org/zap` for structured logging:

```go
logger.Logger.Infow("User logged in", "user_id", userID, "email", email)
logger.Logger.Errorw("Database error", "error", err, "query", query)
logger.Logger.Debugw("Request received", "method", r.Method, "path", r.URL.Path)
```

---

## 🐳 Docker Deployment

### Local Development with Docker Compose

```bash
# Start all services (PostgreSQL only, RabbitMQ is commented)
docker compose up -d

# Verify containers
docker compose ps

# View logs
docker compose logs postgres

# Stop services
docker compose down

# Stop and remove volumes
docker compose down -v
```

### Production Deployment

The project includes `render.yaml` for Render.com deployment:

```yaml
services:
  - type: web
    name: debtease-backend
    buildCommand: go build -o debtease-backend cmd/main.go
    startCommand: ./debtease-backend
    envVars:
      - key: DB_URL
        fromDatabase:
          name: debtease-db
          property: connectionString
```

Deploy to Render:
```bash
git push  # Render auto-deploys on push to main
```

---

## 📞 Support & Troubleshooting

### Common Issues

| Issue | Solution |
|-------|----------|
| `DB_URL not set` | Add `DB_URL` to `.env` file |
| `connection refused` | Ensure PostgreSQL is running on port 5432 |
| `JWT_SECRET is required` | Set `JWT_SECRET` in `.env` |
| `migration failed` | Check database exists: `CREATE DATABASE debtease;` |
| `sqlc generation error` | Verify SQL syntax in `database/queries/*.sql` |
| `Port already in use` | Change `PORT` in `.env` or stop other processes |

### Enable Debug Logging

Set environment variable:
```bash
export LOG_LEVEL=debug
go run cmd/main.go
```

### Database Connection Issues

Test connection:
```bash
psql "postgres://postgres:postgres@localhost:5432/debtease?sslmode=disable"
```

---

## 📚 Additional Resources

- [Go Documentation](https://golang.org/doc/)
- [pgx Driver Documentation](https://github.com/jackc/pgx)
- [sqlc Documentation](https://docs.sqlc.dev/)
- [PostgreSQL Documentation](https://www.postgresql.org/docs/)
- [JWT Documentation](https://tools.ietf.org/html/rfc7519)
- [Goose Migration Tool](https://github.com/pressly/goose)

---

## 📄 License

This project is licensed under the MIT License - see LICENSE file for details.

---

## 👤 Author

**Mihir Sinh Chauhan**  
GitHub: [@MihirsinhChauhan](https://github.com/MihirsinhChauhan)

---

## 🤝 Contributing

Contributions are welcome! Please follow these steps:

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit changes (`git commit -m 'Add amazing feature'`)
4. Push to branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

---

## 📞 Questions?

For questions or support:
- Open an issue on GitHub
- Check existing documentation in `/docs` directory
- Review integration tests in `cmd/main_test.go` for usage examples
