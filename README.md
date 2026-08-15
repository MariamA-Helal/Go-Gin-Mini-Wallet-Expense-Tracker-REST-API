# 💰 FinTech Mini-Wallet & Expense Tracker REST API

> A production-ready, highly secure, and concurrent mini-wallet and expense-tracking backend built with **Go**, **Gin**, **GORM**, and **PostgreSQL**. Designed with strict transactional consistency to handle financial assets without race conditions.

---

## 🚀 Project Overview
Unlike standard toy CRUD applications, this project focuses on **money-movement logic**, data atomicity, and concurrency safety. It features strict RBAC, JWT authentication, pagination, financial auditing, monthly budgeting with automated warnings, and a complete Swagger UI playground.

---

## 🏛️ Architecture & Folder Structure
The project follows **Clean Architecture** principles, separating concerns across distinct layers:

```text
wallet-api/
├── cmd/
│   └── api/
│       └── main.go          # Application entrypoint & dependency injection
├── internal/
│   ├── database/            # PostgreSQL & GORM connection setup
│   ├── errors/              # Centralized custom error definitions
│   ├── handler/             # HTTP REST controllers (Gin framework)
│   ├── middleware/          # JWT Auth, RBAC & Rate Limiter security layers
│   ├── models/              # GORM database entities & DTOs
│   ├── repository/          # Data access layer (with Row-Level Locking)
│   ├── router/              # API route groupings & Swagger bindings
│   └── service/             # Core business logic, validation & budget rules
├── docs/                    # Auto-generated Swagger documentation files
├── .env.example             # Environment variables template
├── Dockerfile               # Container build configuration
├── docker-compose.yml       # Multi-container orchestration (App + DB)
└── README.md                # Project documentation
```

## 💡 Business Logic & Software Engineering Highlights
### 1. Concurrency & Data Integrity (FinTech-Grade)
#### **Row-Level Locking (SELECT ... FOR UPDATE):** Used during withdrawals and deposits to prevent Lost Updates and Race Conditions under heavy concurrent requests.

#### **Deadlock Prevention in Transfers:** Peer-to-peer transfers utilize a deterministic locking order (sorting wallet IDs in ascending order before acquiring locks) to completely eliminate circular wait scenarios and database deadlocks.

**Atomic Transactions:** All balance-modifying operations run inside strict GORM transaction blocks with automatic rollbacks on failure.

### 2. Security & Hardening 
#### **JWT Authentication & RBAC:** Secure token-based access with Role-Based Access Control (Users manage only their own wallets; Admins have read-only visibility across accounts).

#### **Rate Limiting Middleware:** IP-based request throttling protecting sensitive endpoints (/signup, /login) against brute-force attacks.

#### **Information Disclosure Defense:** Internal database errors are securely logged on the server side while clients receive sanitized, generic error responses.

#### **Strict Input Validation:** Custom validation rules enforce secure password complexities (Uppercase, Lowercase, Number, Symbol) and strictly formatted usernames.

### 3. Smart Budgeting & Alerts
#### **Users can set monthly spending limits per category (e.g., "Food", "Rent").**

#### When a withdrawal or transfer is made, the system evaluates the monthly summary via SQL aggregation (GROUP BY) and appends a dynamic warning if the budget cap is exceeded (without blocking the transaction).

## 🛠️ Tech Stack
### Language: Go (Golang)

### Web Framework: Gin

### ORM: GORM

### Database: PostgreSQL

### Authentication: JWT (golang-jwt/jwt/v5) & Bcrypt

### Documentation: Swagger (swaggo/swag)

### Containerization: Docker & Docker Compose

## 🔌 API Endpoints & Testing (Swagger UI)
You can interact with and test all API endpoints directly using the built-in Swagger UI playground.

### 1. Start the server (or via Docker).

### 2. Open your browser and navigate to:

```Plaintext
http://localhost:8080/swagger/index.html
```
### 3. Use /api/signup or /api/login to obtain your Bearer Token.

### 4. Click Authorize at the top right of the Swagger UI, enter Bearer <your_token>, and test all protected endpoints (/wallet, /wallet/deposit, /wallet/withdraw, /wallet/transfer, /wallet/budgets, etc.).


## 🐳 Running the Application
### Option 1: Running with Docker Compose
To run the app and PostgreSQL database together seamlessly using containers:

``` Bash
docker-compose up --build
```

### Option 2: Running Locally without Docker
Clone the repository and configure your environment variables:

```Bash
cp .env.example .env
```

Run the application:

```Bash
go run cmd/api/main.go
```

## ⭐️ Developed with precision, security, and clean architecture.