# Ecommerce Microservices (Go)

A production-ready e-commerce system built with Go, converted from a modular monolith to a **Microservices Architecture**. It features 4 independent services, an API Gateway, and dedicated databases for each service.

## 🏗️ Architecture

The system is composed of the following services:

| Service | Port | Description | Database |
| :--- | :--- | :--- | :--- |
| **API Gateway** | `9090` | Reverse proxy, CORS, Request Logging | - |
| **User Service** | `9091` | Authentication (JWT), User Management | `user_db` |
| **Catalog Service** | `9092` | Product & Category Management | `catalog_db` |
| **Order Service** | `9093` | Order Processing & History | `order_db` |

### Tech Stack
- **Language**: Go 1.24+
- **Framework**: Gin Web Framework
- **Database**: PostgreSQL (GORM)
- **Infrastructure**: Docker, Docker Compose
- **Logging**: Zap (Structured Logging)
- **Observability**: OpenTelemetry (traces + metrics + logs) → Grafana LGTM stack (Loki, Grafana, Tempo, Prometheus) via an OpenTelemetry Collector
- **Documentation**: Swagger (Swaggo)

## 📂 Project Structure

```bash
.
├── pkg/                # Shared code (Logger, Errors, Middleware, Security, DB)
├── services/
│   ├── gateway/        # API Gateway (Reverse Proxy)
│   ├── user/           # User & Auth Service
│   ├── catalog/        # Product & Category Service
│   └── order/          # Order Service
├── docker-compose.yml  # Orchestration for all services + databases
├── Makefile            # Development commands
└── go.work             # Go workspace for local development
```

## 🚀 Getting Started

### Prerequisites
- Docker & Docker Compose
- Go 1.24+ (optional, for local dev)
- Make (optional)

### Quick Start (Docker)

1. **Clone the repository**
   ```bash
   git clone <repo-url>
   cd ecommerce-microservice-go
   ```

2. **Start all services**
   ```bash
   make up
   # OR
   docker compose up -d --build
   ```

3. **Verify**
   Check if all containers are running:
   ```bash
   docker compose ps
   ```

### Accessing Endpoints

All requests go through the **API Gateway** on port `9090`.

**Health Check:**
```bash
curl http://localhost:9090/v1/health
```

**Auth (Login):**
```bash
POST http://localhost:9090/v1/auth/login
{
    "email": "admin@example.com",
    "password": "admin123"
}
```

**Products (Public):**
```bash
GET http://localhost:9090/v1/product/
```

**Orders (Protected - Requires Bearer Token):**
```bash
GET http://localhost:9090/v1/order/
Authorization: Bearer <your-access-token>
```

## 📊 Observability

Every service is instrumented with **OpenTelemetry** and exports the three pillars
(traces, metrics, logs) over OTLP to an **OpenTelemetry Collector**, which fans them
out to the **Grafana LGTM** backends. All of this starts automatically with
`docker compose up -d --build`.

| UI / Endpoint | URL | Purpose |
| :--- | :--- | :--- |
| **Grafana** | http://localhost:3000 | Single pane of glass (anonymous admin, dev only) |
| **Prometheus** | http://localhost:9099 | Metrics store (host port 9099 → container 9090) |
| **Tempo** | http://localhost:3200 | Traces store |
| **Loki** | http://localhost:3100 | Logs store |
| **OTel Collector** | `:4317` gRPC / `:4318` HTTP | OTLP ingest from services |

What you get:
- **Traces**: a single trace follows a request `gateway → service → DB` (W3C context
  propagation + GORM query spans).
- **Logs**: Zap logs are written to both stdout (`docker logs`) and Loki, tagged with
  `trace_id`/`span_id` so Grafana links a log line to its trace and back.
- **Metrics**: HTTP server latency/throughput, Go runtime, and DB pool stats per service.

### Verify end-to-end
```bash
docker compose up -d --build
docker compose ps                                   # all containers healthy

# Generate traffic through the gateway
curl http://localhost:9090/v1/health
curl -X POST http://localhost:9090/v1/auth/login \
     -H 'Content-Type: application/json' \
     -d '{"email":"admin@example.com","password":"admin123"}'
```
Then open Grafana (http://localhost:3000) → **Explore**:
- **Tempo** → Search → open a trace and confirm `gateway → user-service` with nested DB spans.
- **Loki** → `{service_name="user-service"}` → click a line's `TraceID` field to jump to its trace.
- **Prometheus** → query `http_server_request_duration_seconds_bucket` or `go_goroutine_count`.

## 🛠️ Development

### Local Build
To build all services locally:
```bash
make sync       # Sync go.work dependencies
go build ./...  # Build everything
```

### Swagger Documentation
To regenerate Swagger documentation for all services:
```bash
make swagger
```
*Note: Swagger UI is currently available per-service during development if enabled in code, but typically accessed via endpoint discovery.*

### Clean Up
To stop services and remove volumes (reset databases):
```bash
make clean
```

## 📝 License
MIT License
