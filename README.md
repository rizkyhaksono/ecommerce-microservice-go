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
