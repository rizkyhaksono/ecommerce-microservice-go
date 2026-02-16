# Go Microservice Coding Guidelines

Panduan ini berlaku untuk project **ecommerce-microservice-go** yang menggunakan **Clean Architecture** dengan Go.

---

## 1. Project Structure

Setiap entity baru (misal: `order`, `product`, `category`) harus mengikuti pola yang sama:

```
src/
├── domain/<entity>/
│   ├── <entity>.go           # Struct entity + service interface
│   └── <entity>_test.go      # Unit test untuk domain logic
├── application/usecases/<entity>/
│   ├── <entity>.go           # Use case interface + implementasi
│   └── <entity>_test.go      # Unit test use case (mock repository)
└── infrastructure/
    ├── repository/psql/<entity>/
    │   ├── <entity>.go       # GORM model + repository implementasi
    │   └── <entity>_test.go  # Unit test repository (sqlmock)
    └── rest/
        ├── controllers/<entity>/
        │   ├── <Entity>.go        # Controller + request/response structs
        │   ├── <Entity>_test.go   # Controller unit test
        │   └── Validation.go      # Custom validasi untuk update
        └── routes/<entity>.go     # Route registration
```

---

## 2. Membuat Entity Baru (Step-by-Step)

### Step 1: Domain Layer

```go
// src/domain/order/order.go
package order

import (
    "time"
    "github.com/gbrayhan/microservices-go/src/domain"
)

// Entity — plain struct tanpa dependency external
type Order struct {
    ID        int
    UserID    int
    Status    string
    Total     float64
    CreatedAt time.Time
    UpdatedAt time.Time
}

// Result type untuk search/pagination
type SearchResultOrder struct {
    Data       *[]Order
    Total      int64
    Page       int
    PageSize   int
    TotalPages int
}

// Service interface — didefinisikan di domain, diimplementasi oleh use case
type IOrderService interface {
    GetAll() (*[]Order, error)
    GetByID(id int) (*Order, error)
    Create(order *Order) (*Order, error)
    Delete(id int) error
    Update(id int, orderMap map[string]any) (*Order, error)
    SearchPaginated(filters domain.DataFilters) (*SearchResultOrder, error)
}
```

### Step 2: Application Layer (Use Case)

```go
// src/application/usecases/order/order.go
package order

import (
    "github.com/gbrayhan/microservices-go/src/domain"
    orderDomain "github.com/gbrayhan/microservices-go/src/domain/order"
    logger "github.com/gbrayhan/microservices-go/src/infrastructure/logger"
    "github.com/gbrayhan/microservices-go/src/infrastructure/repository/psql/order"
    "go.uber.org/zap"
)

// Use case interface
type IOrderUseCase interface {
    GetAll() (*[]orderDomain.Order, error)
    GetByID(id int) (*orderDomain.Order, error)
    Create(order *orderDomain.Order) (*orderDomain.Order, error)
    Delete(id int) error
    Update(id int, orderMap map[string]any) (*orderDomain.Order, error)
    SearchPaginated(filters domain.DataFilters) (*orderDomain.SearchResultOrder, error)
}

// Concrete struct — unexported fields
type OrderUseCase struct {
    orderRepository order.OrderRepositoryInterface
    Logger          *logger.Logger
}

// Constructor — return interface, bukan concrete type
func NewOrderUseCase(repo order.OrderRepositoryInterface, log *logger.Logger) IOrderUseCase {
    return &OrderUseCase{
        orderRepository: repo,
        Logger:          log,
    }
}

func (s *OrderUseCase) GetByID(id int) (*orderDomain.Order, error) {
    s.Logger.Info("Getting order by ID", zap.Int("id", id))
    return s.orderRepository.GetByID(id)
}

// ... implement remaining methods
```

### Step 3: Repository Layer

```go
// src/infrastructure/repository/psql/order/order.go
package order

import (
    domainErrors "github.com/gbrayhan/microservices-go/src/domain/errors"
    domainOrder "github.com/gbrayhan/microservices-go/src/domain/order"
    logger "github.com/gbrayhan/microservices-go/src/infrastructure/logger"
    "go.uber.org/zap"
    "gorm.io/gorm"
)

// Repository interface
type OrderRepositoryInterface interface {
    GetAll() (*[]domainOrder.Order, error)
    GetByID(id int) (*domainOrder.Order, error)
    Create(order *domainOrder.Order) (*domainOrder.Order, error)
    Delete(id int) error
    Update(id int, orderMap map[string]any) (*domainOrder.Order, error)
    SearchPaginated(filters domain.DataFilters) (*domainOrder.SearchResultOrder, error)
}

// GORM model — terpisah dari domain entity
type Order struct {
    ID        int       `gorm:"primaryKey"`
    UserID    int       `gorm:"not null"`
    Status    string    `gorm:"default:'pending'"`
    Total     float64   `gorm:"type:decimal(10,2)"`
    CreatedAt time.Time `gorm:"autoCreateTime:milli"`
    UpdatedAt time.Time `gorm:"autoUpdateTime:milli"`
}

func (*Order) TableName() string {
    return "orders"
}

// Column mapping — WAJIB untuk dynamic queries (prevent SQL injection)
var ColumnsOrderMapping = map[string]string{
    "id":        "id",
    "userId":    "user_id",
    "status":    "status",
    "total":     "total",
    "createdAt": "created_at",
    "updatedAt": "updated_at",
}

type Repository struct {
    DB     *gorm.DB
    Logger *logger.Logger
}

func NewOrderRepository(DB *gorm.DB, log *logger.Logger) OrderRepositoryInterface {
    return &Repository{DB: DB, Logger: log}
}

// Mapper — GORM model ke domain entity
func (o *Order) toDomainMapper() *domainOrder.Order {
    return &domainOrder.Order{
        ID:        o.ID,
        UserID:    o.UserID,
        Status:    o.Status,
        Total:     o.Total,
        CreatedAt: o.CreatedAt,
        UpdatedAt: o.UpdatedAt,
    }
}

func (r *Repository) GetByID(id int) (*domainOrder.Order, error) {
    var order Order
    err := r.DB.Where("id = ?", id).First(&order).Error
    if err != nil {
        if err == gorm.ErrRecordNotFound {
            r.Logger.Warn("Order not found", zap.Int("id", id))
            return nil, domainErrors.NewAppErrorWithType(domainErrors.NotFound)
        }
        r.Logger.Error("Error getting order", zap.Error(err), zap.Int("id", id))
        return nil, domainErrors.NewAppErrorWithType(domainErrors.UnknownError)
    }
    return order.toDomainMapper(), nil
}

// ... implement remaining methods
```

### Step 4: Controller Layer

```go
// src/infrastructure/rest/controllers/order/Order.go
package order

import (
    "net/http"
    "strconv"

    domainError "github.com/gbrayhan/microservices-go/src/domain/errors"
    orderDomain "github.com/gbrayhan/microservices-go/src/domain/order"
    logger "github.com/gbrayhan/microservices-go/src/infrastructure/logger"
    "github.com/gbrayhan/microservices-go/src/infrastructure/rest/controllers"
    "github.com/gin-gonic/gin"
    "go.uber.org/zap"
)

// Request struct — dengan validation tags
type NewOrderRequest struct {
    UserID int     `json:"userId" binding:"required"`
    Total  float64 `json:"total" binding:"required,gt=0"`
}

// Response struct — terpisah dari domain entity
type ResponseOrder struct {
    ID     int     `json:"id"`
    UserID int     `json:"userId"`
    Status string  `json:"status"`
    Total  float64 `json:"total"`
}

// Controller interface
type IOrderController interface {
    NewOrder(ctx *gin.Context)
    GetAllOrders(ctx *gin.Context)
    GetOrderByID(ctx *gin.Context)
    UpdateOrder(ctx *gin.Context)
    DeleteOrder(ctx *gin.Context)
}

type Controller struct {
    orderService orderDomain.IOrderService
    Logger       *logger.Logger
}

func NewOrderController(svc orderDomain.IOrderService, log *logger.Logger) IOrderController {
    return &Controller{orderService: svc, Logger: log}
}

// Mapper — domain ke response (JANGAN expose domain entity langsung)
func domainToResponseMapper(o *orderDomain.Order) *ResponseOrder {
    return &ResponseOrder{
        ID:     o.ID,
        UserID: o.UserID,
        Status: o.Status,
        Total:  o.Total,
    }
}

func (c *Controller) NewOrder(ctx *gin.Context) {
    c.Logger.Info("Creating new order")
    var req NewOrderRequest
    if err := controllers.BindJSON(ctx, &req); err != nil {
        c.Logger.Error("Error binding JSON", zap.Error(err))
        _ = ctx.Error(domainError.NewAppError(err, domainError.ValidationError))
        return
    }
    // ...
}
```

### Step 5: Routes

```go
// src/infrastructure/rest/routes/order.go
package routes

import (
    "github.com/gbrayhan/microservices-go/src/infrastructure/rest/controllers/order"
    "github.com/gbrayhan/microservices-go/src/infrastructure/rest/middlewares"
    "github.com/gin-gonic/gin"
)

func OrderRoutes(router *gin.RouterGroup, controller order.IOrderController) {
    ord := router.Group("/order")
    ord.Use(middlewares.AuthJWTMiddleware())
    {
        ord.GET("/", controller.GetAllOrders)
        ord.POST("/", controller.NewOrder)
        ord.GET("/:id", controller.GetOrderByID)
        ord.PUT("/:id", controller.UpdateOrder)
        ord.DELETE("/:id", controller.DeleteOrder)
    }
}
```

### Step 6: Register di DI & Routes

```go
// src/infrastructure/di/application_context.go
// Tambahkan field:
//   OrderController    orderController.IOrderController
//   OrderRepository    order.OrderRepositoryInterface
//   OrderUseCase       orderUseCase.IOrderUseCase

// Dalam SetupDependencies():
//   orderRepo := order.NewOrderRepository(db, loggerInstance)
//   orderUC := orderUseCase.NewOrderUseCase(orderRepo, loggerInstance)
//   orderCtrl := orderController.NewOrderController(orderUC, loggerInstance)

// src/infrastructure/rest/routes/routes.go
// Tambahkan: OrderRoutes(v1, appContext.OrderController)
```

---

## 3. Error Handling

### Gunakan AppError

```go
// ✅ Benar
return nil, domainErrors.NewAppErrorWithType(domainErrors.NotFound)
return nil, domainErrors.NewAppError(errors.New("custom message"), domainErrors.ValidationError)

// ❌ Salah
return nil, errors.New("not found")
return nil, fmt.Errorf("something went wrong")
```

### Di Controller — gunakan ctx.Error()

```go
// ✅ Benar — error middleware akan handle response
_ = ctx.Error(err)
return

// ❌ Salah — jangan manual set response untuk error
ctx.JSON(http.StatusNotFound, gin.H{"error": "not found"})
```

---

## 4. Logging

### Selalu pakai Zap structured logging

```go
// ✅ Benar
s.Logger.Info("Creating order", zap.Int("userId", order.UserID), zap.Float64("total", order.Total))
s.Logger.Error("Failed to create order", zap.Error(err), zap.Int("userId", order.UserID))
s.Logger.Warn("Order not found", zap.Int("id", id))

// ❌ Salah
fmt.Println("creating order")
log.Printf("error: %v", err)
s.Logger.Info(fmt.Sprintf("Creating order for user %d", order.UserID))
```

---

## 5. Testing

### Unit Test Pattern

```go
func TestGetByID_Success(t *testing.T) {
    // Arrange
    mockRepo := new(MockOrderRepository)
    logger, _ := logger.NewDevelopmentLogger()
    uc := NewOrderUseCase(mockRepo, logger)

    expected := &orderDomain.Order{ID: 1, Status: "pending"}
    mockRepo.On("GetByID", 1).Return(expected, nil)

    // Act
    result, err := uc.GetByID(1)

    // Assert
    assert.NoError(t, err)
    assert.Equal(t, expected, result)
    mockRepo.AssertExpectations(t)
}
```

### Jalankan Tests

```bash
# Semua test
make tests

# Test spesifik
make tests-TestGetByID_Success

# Integration test
make integration-test

# Dengan coverage
go test -v -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

---

## 6. Database Migration

```bash
# Buat migration baru
make migration-order

# Jalankan migration
make migrate-up

# Rollback
make migrate-down

# Via Docker
make migrate-docker-up
make migrate-docker-down
```

### Migration file template

```sql
-- 000X_create-table-order.up.sql
CREATE TABLE IF NOT EXISTS orders (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id),
    status VARCHAR(50) DEFAULT 'pending',
    total DECIMAL(10,2) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_orders_user_id ON orders(user_id);
CREATE INDEX idx_orders_status ON orders(status);
```

```sql
-- 000X_create-table-order.down.sql
DROP TABLE IF EXISTS orders;
```

---

## 7. API Response Format

### Success — return data langsung

```json
// Single entity
{"id": 1, "userId": 2, "status": "pending", "total": 150.00}

// Collection
[{"id": 1, ...}, {"id": 2, ...}]

// Paginated
{
  "data": [...],
  "total": 100,
  "page": 1,
  "pageSize": 10,
  "totalPages": 10
}
```

### Error — di-handle oleh ErrorHandler middleware

```json
{"error": "record not found"}
```

---

## 8. Environment Variables

Semua config dari environment variables, TIDAK hardcode.

```bash
# .env file (copy dari .env.example)
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=devPassword123
DB_NAME=boilerplate_go
SERVER_PORT=8080
JWT_ACCESS_SECRET_KEY=secretKey
JWT_ACCESS_TIME_MINUTE=15
JWT_REFRESH_SECRET_KEY=refreshKey
JWT_REFRESH_TIME_HOUR=168
```

---

## 9. Pre-Commit Checklist

Sebelum commit, pastikan:

```bash
make lint                    # golangci-lint
go vet ./...                 # Go vet
staticcheck ./...            # Static analysis
make tests                   # Unit tests pass
go mod tidy                  # Dependencies clean
```

Atau otomatis via `lefthook` (sudah dikonfigurasi di `lefthook.yml`).

---

## 10. Aturan Penting

| Rule | Detail |
|------|--------|
| Clean Architecture | Domain TIDAK boleh import infrastructure/application |
| Interface First | Definisikan interface sebelum implementasi |
| Constructor Pattern | `NewXxx()` return interface type |
| Error Handling | Selalu pakai `AppError`, JANGAN `errors.New()` langsung di response |
| Logging | Zap structured logging, JANGAN `fmt.Println` |
| Mapper Required | JANGAN expose domain entity langsung ke API response |
| Column Mapping | Pakai `ColumnsXxxMapping` untuk dynamic queries |
| DI Wiring | Register semua dependency di `ApplicationContext` |
| Test Coverage | Target ≥ 80% |
| Git Hooks | `lefthook` otomatis lint + test sebelum commit |
