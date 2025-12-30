# PayFlow | Fintech Ledger System

PayFlow is an industry-grade mock backend engine designed to handle financial transactions with the same rigor as modern banking systems. Built with **Go**, **Postgres**, and **Redis**, it solves the "Double-Spend" and "Data Loss" problems using distributed systems patterns.

## 🏗 Architectural Pillars

### 1. Immutable Double-Entry Ledger
Unlike simple "balance update" applications, PayFlow utilizes a strictly immutable ledger. Every transfer creates atomic entries (debits and credits) tied to a unique transaction ID.
* **Auditability:** Reconstruct any account balance by summing its historical entries.
* **Integrity:** Balances are denormalized for read performance but verified against the entry sum via a **Global Audit** check.

### 2. Transactional Outbox Pattern
To ensure reliability, the system decouples the database update from the external bank/provider call. 
* When a transfer occurs, a "Task" is written to the database in the **same transaction** as the ledger update. 
* A dedicated **Background Worker** polls these tasks and executes the payout, ensuring that even if the server crashes, no payout is ever lost.

### 3. Distributed Idempotency
Financial systems must handle "Retry" logic gracefully. PayFlow implements a **Redis-backed Middleware** using unique idempotency keys (`X-Idempotency-Key`).
* Prevents accidental double-charging if a user clicks "Pay" twice.
* Utilizes atomic `SETNX` operations to ensure sub-millisecond locking.

### 4. Concurrency & Locking
* **Pessimistic Locking:** Uses `SELECT FOR UPDATE` to lock specific account rows during transfers, preventing race conditions.
* **Worker Optimization:** Uses `FOR UPDATE SKIP LOCKED` in background processing, allowing workers to scale horizontally without processing the same task twice.

---

## 🛠 Tech Stack

| Component | Technology | Role |
| :--- | :--- | :--- |
| **Language** | **Go (Golang)** | High-concurrency execution and type safety. |
| **API Framework** | **Gin Gonic** | High-performance routing and middleware. |
| **Primary DB** | **PostgreSQL** | ACID-compliant storage for the ledger and outbox. |
| **Cache/Locking** | **Redis** | Distributed idempotency store. |
| **Tooling** | **Docker** | Containerized environment for consistent deployment. |

---

## 🚦 Getting Started

### Prerequisites
* Docker & Docker Compose
* Go 1.21+

### Installation

1. **Clone & Setup**
   ```bash
   git clone [https://github.com/TharunNo1/payflow.git](https://github.com/TharunNo1/payflow.git)
   cd payflow
   ```

2. **Spin up Infrastructure**
    ```bash
    docker-compose up -d
    ```

3. **Run Migrations & Start**
    ```bash
    go run cmd/api/main.go
    ```

## 🧪 Testing the Flow
1. **Seed Accounts**
    Log into the DB (make db-shell) and run:

    ```sql
    INSERT INTO accounts (id, owner_name, balance) VALUES 
    ('00000000-0000-0000-0000-000000000001', 'Treasury', 1000000), 
    ('00000000-0000-0000-0000-000000000002', 'User_Alpha', 0);
    ```

2. **Perform a Transfer**
    ```bash
    curl -X POST http://localhost:8080/transfer \
        -H "Content-Type: application/json" \
        -H "X-Idempotency-Key: unique-tx-101" \
        -d '{
            "from_account_id": "00000000-0000-0000-0000-000000000001",
            "to_account_id": "00000000-0000-0000-0000-000000000002",
            "amount": 5000
        }'
    ```

3. **Verify Audit Sum**
    ```bash
    curl http://localhost:8080/health
    # Response: {"status":"healthy","ledger_integrity_sum":0}
    ```

## 📄 Key Design Decisions
- Integer Math: All currency is handled as int64 (cents) to avoid floating-point precision errors.

- Graceful Shutdown: Listens for SIGINT/SIGTERM to allow active database transactions to commit before exiting.

- Structured Logging: Uses clear event tracing for easier debugging in distributed environments.

## ⚖️ License
Distributed under the MIT License.
