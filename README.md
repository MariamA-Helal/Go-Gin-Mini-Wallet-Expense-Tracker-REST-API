# 🚀 Go Wallet & Expense Tracker REST API - Capstone Project

![Go](https://img.shields.io/badge/Go-1.21+-00ADD8?style=for-the-badge&logo=go)
![Gin](https://img.shields.io/badge/Gin-Framework-00ADD8?style=for-the-badge&logo=go)
![PostgreSQL](https://img.shields.io/badge/PostgreSQL-15-4169E1?style=for-the-badge&logo=postgresql)
![Docker](https://img.shields.io/badge/Docker-Containerized-2496ED?style=for-the-badge&logo=docker)
![Swagger](https://img.shields.io/badge/Swagger-API_Docs-85EA2D?style=for-the-badge&logo=swagger)

A robust, highly scalable, and fully containerized financial RESTful API built with **Go (Golang)**, **Gin Framework**, **GORM**, and **PostgreSQL**. This project serves as a secure mini-wallet system supporting financial transactions, concurrent transfers, and expense tracking.

Developed as a Capstone Backend Engineering Project, focusing on **ACID compliance**, **Concurrency Safety (Race Condition Prevention)**, database architecture, and containerization.

---

## 🧠 System Design & Business Rules

Managing financial data requires strict rules. This project implements the following:

### 1. ACID Transactions & Concurrency Control
*   **Problem:** If two withdrawal requests happen simultaneously, a user might withdraw more money than their balance allows (Race Condition).
*   **Solution:** All financial operations (Deposit, Withdrawal, Transfer) are wrapped in **Database Transactions**. We utilize Row-level locking (`SELECT ... FOR UPDATE`) in PostgreSQL to ensure that balance calculations are atomic and sequential.

### 2. Core Business Rules
*   **Insufficient Funds:** A withdrawal or expense transaction will be strictly blocked at the Service layer if the wallet balance is less than the requested amount.
*   **Rollback Mechanism:** If an error occurs midway through a transfer (e.g., deducting from Sender succeeds, but adding to Receiver fails), the entire transaction is rolled back. No money is lost or created out of thin air.

---

## 🌟 Key Features

### Security & Authentication
* **JWT-Based Authentication:** Secure endpoints using JSON Web Tokens (Bearer Auth).
* **Data Isolation:** Complete data privacy where users can only view, manage, and transfer funds associated with their authenticated accounts.

### Core Financial Logic & Concurrency
* **Wallet Management:** Seamless deposits and withdrawals reflecting real-time balances.
* **Concurrency-Safe Transfers:** Enterprise-grade fund transfers using **Atomic Database Transactions** and **Pessimistic Row-Level Locking (`FOR UPDATE`)** to strictly prevent race conditions, double-spending, and negative balances during simultaneous requests.
* **Transaction History & Summary:** Comprehensive tracking of all financial movements with aggregated summaries for expense analysis.
* **Budget Tracking:** Ability to set custom budgets per expense category and monitor budget status in real-time.

### DevOps & Architecture
* **Clean Architecture:** Strict separation of concerns (Models, Repositories, Services, Handlers, Routers).
* **Dockerized Environment:** Multi-stage Docker builds and `docker-compose` orchestration for zero-configuration setups and database provisioning.
* **Automated Testing:** Extensive unit and concurrency testing to validate transaction safety under heavy concurrent loads.
* **API Documentation:** Interactive Swagger UI integration for seamless testing and endpoint exploration.

---

## 🛠️ Tech Stack & Tooling

* **Language:** Go (Golang)
* **Framework:** Gin (`github.com/gin-gonic/gin`)
* **Database & ORM:** PostgreSQL & GORM (`gorm.io/gorm`)
* **Authentication:** JWT (`golang-jwt/jwt`) & bcrypt for password hashing
* **Containerization:** Docker & Docker Compose
* **Testing:** Go testing package (Concurrency & Unit Tests)
* **Documentation:** Swaggo (Swagger UI)

---

## 📁 Architecture & Project Structure

The codebase is organized into isolated layers to ensure decoupling and ease of testing:

```text
Wallet/
├── cmd/api/                 # 🚀 Entry point of the application
├── docs/                    # 📖 Auto-generated Swagger documentation files
├── internal/
│   ├── database/            # 🐘 PostgreSQL connection & auto-migrations
│   ├── errors/              # 🛑 Centralized custom error handling
│   ├── handler/             # 🌐 HTTP layer, parsing requests, and returning JSON
│   ├── models/              # 📦 Data structures, DTOs, and GORM schema definitions
│   ├── repository/          # 🗄️ Database CRUD logic & Transaction locking (FOR UPDATE)
│   ├── router/              # 🛤️ API route definitions and middleware injection
│   └── service/             # ⚙️ Core business logic and validation
├── Dockerfile               # 🐳 Multi-stage build instructions for the Go API
├── docker-compose.yml       # 🐙 Orchestration for API and Database containers
├── go.mod & go.sum          # 📦 Go module dependencies
└── .env                     # 🔐 Environment variables (Ignored in Git)
```

---

## 🚀 Getting Started

### Prerequisites
* Docker & Docker Desktop installed.
* Git installed.

### Option 1: Run via Docker Hub (Fastest)
You can pull and run the pre-built image directly from Docker Hub without needing the source code:
```bash
docker pull mariamamr286/wallet-api:latest
docker run -p 8080:8080 mariamamr286/wallet-api:latest
```

### Option 2: Build & Run from Source (Development)
1. **Clone the repository:**
   ```bash
   git clone [Your-GitHub-Repository-Link]
   cd Wallet
   ```

2. **Setup Environment Variables:**
   Create a `.env` file in the root directory and add the following configurations:
   ```env
   # Database Configuration
   DB_HOST=postgres_db
   DB_USER=postgres
   DB_PASSWORD=your_password
   DB_NAME=wallet_db
   DB_PORT=5432
   
   # JWT Secret
   JWT_SECRET=your_super_secret_key
   ```

3. **Start the application using Docker Compose:**
   ```bash
   docker-compose up --build
   ```
   *The database will initialize automatically, and GORM will handle auto-migrations.*

---

## 📡 API Documentation & Endpoints

### 📖 Swagger Documentation
| Method | Endpoint | Description | Access |
|--------|----------|-------------|--------|
| `GET`  | `/swagger/*any` | Interactive API Documentation (Swagger UI) | Public |

### 🔐 Authentication
| Method | Endpoint | Description | Access |
|--------|----------|-------------|--------|
| `POST` | `/api/signup` | Register a new user | Public |
| `POST` | `/api/login` | Authenticate and receive JWT | Public |

### 💰 Wallet & Transactions
| Method | Endpoint | Description | Access |
|--------|----------|-------------|--------|
| `GET`  | `/api/wallet/` | Retrieve current wallet balance | User |
| `POST` | `/api/wallet/deposit` | Deposit funds into wallet | User |
| `POST` | `/api/wallet/withdraw` | Withdraw funds from wallet | User |
| `POST` | `/api/wallet/transfer` | Securely transfer funds to another user | User |
| `GET`  | `/api/wallet/transactions` | Get paginated transaction history | User |
| `GET`  | `/api/wallet/transactions/summary` | Get expense/income summary | User |

### 📊 Budget Management
| Method | Endpoint | Description | Access |
|--------|----------|-------------|--------|
| `POST` | `/api/wallet/budgets` | Set a new budget for a category | User |
| `PUT`  | `/api/wallet/budgets/:category` | Update an existing budget | User |
| `GET`  | `/api/wallet/budgets/status` | View current spending vs. budget limits | User |

*(Note: Protected API calls require the `Authorization` header formatted as: `Bearer <token>`)*

---

## 🧪 Testing Strategy
Testing financial logic is critical. Our test suite includes:

*   **Business Rule Tests:** Asserts that `Withdraw()` returns an `ErrInsufficientFunds` when balance < amount.
*   **Transaction Rollback Tests:** Verifies that a database failure during a multi-step operation (like transfers) strictly rolls back the state, leaving sender balances unchanged.
*   **Concurrency & Deadlock Prevention Tests:** Uses Goroutines to simulate simultaneous withdrawals and bidirectional transfers, ensuring Row-level locking (`SELECT ... FOR UPDATE`) prevents race conditions and data anomalies.

To run the full suite of tests:
```bash
go test ./... -v
```

---

## 👤 Author
**Mariam Amr Helal**
* LinkedIn: [https://www.linkedin.com/in/mariam-helal-464994212/](https://www.linkedin.com/in/mariam-helal-464994212/)
* DockerHub: [https://hub.docker.com/r/mariamamr286/wallet-api](https://hub.docker.com/r/mariamamr286/wallet-api)
