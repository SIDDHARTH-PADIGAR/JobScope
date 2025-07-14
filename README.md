# JobScope - Concurrent Job Processing API

JobScope is a production-grade Go REST API for managing and processing background jobs. Built with concurrency, security, and extensibility in mind, it's designed for real-world deployment scenarios and includes robust job handling, JWT authentication, and rate limiting.

---

## Features

- RESTful API (CRUD-style endpoints)
- Concurrent job processing with Goroutines & Channels
- JWT Authentication for protected routes
- Graceful shutdown support
- Real-time job statistics endpoint
- Rate Limiting (WIP)
- Docker-ready (coming up)

---

## API Endpoints

### Public Endpoints

| Method | Endpoint           | Description                |
|--------|--------------------|----------------------------|
| `GET`  | `/jobs`            | Get all jobs               |
| `GET`  | `/jobs/{id}`       | Get job by ID              |
| `GET`  | `/jobs/stats`      | Get job queue stats        |
| `POST` | `/login`           | Get JWT token (auth)       |

### Protected Endpoints (Require JWT)

| Method   | Endpoint               | Description               |
|----------|------------------------|---------------------------|
| `POST`   | `/jobs`                | Create a new job          |
| `PATCH`  | `/jobs/{id}/status`    | Update job status         |

---

## Authentication

To access protected endpoints:

1. `POST /login` with:
    ```json
    {
      "username": "admin",
      "password": "password"
    }
    ```

2. Copy the `token` from response.

3. Use it as a Bearer token in `Authorization` header:
    ```
    Authorization: Bearer <token>
    ```

---

## Architecture Highlights

- **Job Queue:** Buffered `chan Job` used as a queue.
- **Workers:** Configurable pool of goroutines processing jobs in parallel.
- **Mutex Locking:** Thread-safe job list updates.
- **Job Storage:** Persisted in local JSON file (`jobs.json`).
- **Graceful Shutdown:** Handles OS signals, cleans up workers.
- **Logging:** Logs to `logs/app.log` and stdout.

---

## Getting Started

### Prerequisites

- Go 1.20+
- Git

### Installation

```bash
git clone https://github.com/SIDDHARTH-PADIGAR/go-api
cd go-api
go mod tidy
go run main.go
```
