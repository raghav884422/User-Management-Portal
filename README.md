# 📋 User API — Go Backend Development Task

A RESTful API built with **Go + GoFiber + PostgreSQL + SQLC** that manages users with their date of birth and dynamically calculates their age on every fetch.

---

## 🏗️ Project Structure

```
user-api/
├── cmd/
│   └── server/
│       └── main.go              # Application entry point
├── config/
│   └── config.go                # Environment-based configuration
├── db/
│   ├── migrations/
│   │   └── 001_create_users.sql # SQL schema migration
│   └── sqlc/
│       ├── query.sql            # SQLC query definitions
│       ├── query.sql.go         # Generated query implementations
│       ├── models.go            # Generated data models
│       ├── db.go                # Generated DB wrapper
│       └── querier.go           # Generated Querier interface
├── internal/
│   ├── handler/
│   │   └── user_handler.go      # HTTP handlers + global error handler
│   ├── logger/
│   │   └── logger.go            # Uber Zap logger setup
│   ├── middleware/
│   │   └── middleware.go        # RequestID + RequestLogger middleware
│   ├── models/
│   │   ├── user.go              # API request/response models + CalculateAge
│   │   └── user_test.go         # Unit tests for age calculation
│   ├── repository/
│   │   └── user_repository.go   # Data-access layer (wraps SQLC)
│   ├── routes/
│   │   └── user_routes.go       # Fiber route registration
│   └── service/
│       └── user_service.go      # Business logic layer
├── .env.example                 # Environment variable template
├── docker-compose.yml           # Docker Compose (API + PostgreSQL)
├── Dockerfile                   # Multi-stage Docker build
├── go.mod                       # Go module definition
├── go.sum                       # Dependency checksums
├── sqlc.yaml                    # SQLC code generation config
└── README.md                    # This file
```

---

## 🛠️ Tech Stack

| Technology | Purpose |
|---|---|
| [Go 1.22](https://go.dev/) | Language |
| [GoFiber v2](https://gofiber.io/) | HTTP framework |
| [PostgreSQL 16](https://www.postgresql.org/) | Database |
| [SQLC](https://sqlc.dev/) | Type-safe SQL code generation |
| [Uber Zap](https://github.com/uber-go/zap) | Structured logging |
| [go-playground/validator](https://github.com/go-playground/validator) | Request validation |
| [pgx/v5](https://github.com/jackc/pgx) | PostgreSQL driver |
| [godotenv](https://github.com/joho/godotenv) | .env file loading |

---

## 🗄️ Database Schema

```sql
CREATE TABLE users (
    id   SERIAL PRIMARY KEY,
    name TEXT   NOT NULL,
    dob  DATE   NOT NULL
);
```

> **Note:** The `age` field is **not** stored in the database. It is calculated dynamically using Go's `time` package whenever a user is fetched.

---

## 🔌 API Endpoints

| Method | Path | Description |
|--------|------|-------------|
| `POST` | `/users` | Create a new user |
| `GET` | `/users` | List all users (with pagination) |
| `GET` | `/users/:id` | Get a user by ID (includes `age`) |
| `PUT` | `/users/:id` | Update a user |
| `DELETE` | `/users/:id` | Delete a user |
| `GET` | `/health` | Health check |

### Create User — `POST /users`

**Request:**
```json
{ "name": "Alice", "dob": "1990-05-10" }
```

**Response `201 Created`:**
```json
{ "id": 1, "name": "Alice", "dob": "1990-05-10" }
```

### Get User by ID — `GET /users/:id`

**Response `200 OK`:**
```json
{ "id": 1, "name": "Alice", "dob": "1990-05-10", "age": 35 }
```

### Update User — `PUT /users/:id`

**Request:**
```json
{ "name": "Alice Updated", "dob": "1991-03-15" }
```

**Response `200 OK`:**
```json
{ "id": 1, "name": "Alice Updated", "dob": "1991-03-15" }
```

### Delete User — `DELETE /users/:id`

**Response `204 No Content`**

### List Users — `GET /users?page=1&page_size=10`

**Response `200 OK`:**
```json
{
  "data": [
    { "id": 1, "name": "Alice", "dob": "1990-05-10", "age": 35 }
  ],
  "total": 1,
  "page": 1,
  "page_size": 10,
  "total_pages": 1
}
```

---

## 🚀 Getting Started

### Prerequisites

- [Go 1.22+](https://go.dev/dl/)
- [PostgreSQL 16+](https://www.postgresql.org/download/) (or Docker)
- [SQLC](https://docs.sqlc.dev/en/latest/overview/install.html) (optional, for code re-generation)
- [Docker & Docker Compose](https://docs.docker.com/get-docker/) (optional)

---

### Option A — Run with Docker Compose (Recommended)

```bash
# 1. Clone the project
git clone https://github.com/yourusername/user-api.git
cd user-api

# 2. Copy environment file
cp .env.example .env

# 3. Start the database and API (migrations run automatically)
docker compose up --build

# 4. The API is now available at http://localhost:8080
```

---

### Option B — Run Locally

#### 1. Install Dependencies

```bash
go mod download
```

#### 2. Set Up Environment Variables

```bash
cp .env.example .env
# Edit .env and fill in your DATABASE_URL
```

#### 3. Run Database Migration

> Ensure PostgreSQL is running, then execute the migration:

```bash
# Using psql directly
psql -U postgres -d userdb -f db/migrations/001_create_users.sql

# Or using your preferred migration tool, e.g. golang-migrate:
migrate -path db/migrations -database "$DATABASE_URL" up
```

#### 4. (Optional) Regenerate SQLC Code

> Only needed if you modify the SQL queries.

```bash
# Install sqlc (if not installed)
go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest

# Generate
sqlc generate
```

#### 5. Start the Server

```bash
go run ./cmd/server
```

The server starts on `http://localhost:8080` by default.

---

## 🧪 Run Unit Tests

```bash
go test ./internal/models/... -v
```

Expected output:
```
=== RUN   TestCalculateAge
--- PASS: TestCalculateAge (0.00s)
=== RUN   TestCalculateAge_SpecificDates
--- PASS: TestCalculateAge_SpecificDates (0.00s)
PASS
```

Run all tests:
```bash
go test ./...
```

---

## 📡 Example cURL Requests

### Create a User

```bash
curl -X POST http://localhost:8080/users \
  -H "Content-Type: application/json" \
  -d '{"name": "Alice", "dob": "1990-05-10"}'
```

### Get User by ID (includes age)

```bash
curl http://localhost:8080/users/1
```

### Update a User

```bash
curl -X PUT http://localhost:8080/users/1 \
  -H "Content-Type: application/json" \
  -d '{"name": "Alice Updated", "dob": "1991-03-15"}'
```

### Delete a User

```bash
curl -X DELETE http://localhost:8080/users/1
```

### List Users (with pagination)

```bash
# Page 1, 10 per page (default)
curl "http://localhost:8080/users"

# Page 2, 5 per page
curl "http://localhost:8080/users?page=2&page_size=5"
```

### Health Check

```bash
curl http://localhost:8080/health
```

---

## 🪵 Logging

The application uses **Uber Zap** for structured logging:

- **Development** mode: colored, human-readable console output
- **Production** mode (set `APP_ENV=production`): structured JSON output

Every request is automatically logged with:
- `request_id` — unique UUID injected per request
- `method` — HTTP method
- `path` — request path
- `status` — response HTTP status code
- `duration` — request processing time
- `ip` — client IP address

---

## 🔐 Environment Variables

| Variable | Default | Description |
|---|---|---|
| `APP_ENV` | `development` | Application environment (`development` / `production`) |
| `APP_PORT` | `8080` | Port the server listens on |
| `DATABASE_URL` | `postgres://postgres:postgres@localhost:5432/userdb?sslmode=disable` | PostgreSQL connection string |
| `POSTGRES_USER` | `postgres` | PostgreSQL user (Docker Compose only) |
| `POSTGRES_PASSWORD` | `postgres` | PostgreSQL password (Docker Compose only) |
| `POSTGRES_DB` | `userdb` | PostgreSQL database name (Docker Compose only) |

---

## 📁 Submission

- Push code to a GitHub repository
- Share the repo link
- Ensure this `README.md` contains clear setup and run instructions ✅
