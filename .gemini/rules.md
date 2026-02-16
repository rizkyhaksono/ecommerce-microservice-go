# Project Rules — Ecommerce Microservice Go

## Project Overview

This is a Go microservices project built with **Clean Architecture** principles.

- **Module**: `github.com/gbrayhan/microservices-go`
- **Go Version**: 1.24.2+
- **Entry Point**: `main.go` (root directory, NOT `src/main.go`)
- **Framework**: Gin-Gonic (HTTP router), GORM (ORM), Zap (Structured Logging)
- **Database**: PostgreSQL
- **Auth**: JWT with access + refresh tokens (HS256)
- **Testing**: testify + go-sqlmock + godog/cucumber BDD

## Architecture & Complete Directory Structure

```
main.go                              ← Application entry point
src/
├── domain/                          ← INNERMOST LAYER — NO external dependencies
│   ├── Types.go                     ← Shared types (DataFilters, SortDirection)
│   ├── errors/                      ← Centralized error system
│   │   ├── Errors.go                ← AppError, ErrorType constants, NewAppError(), AppErrorToHTTP()
│   │   ├── Errors_test.go
│   │   └── Gorm.go                  ← GormErr struct for DB error parsing
│   ├── user/                        ← User entity + IUserService interface
│   │   ├── user.go
│   │   └── user_test.go
│   └── medicine/                    ← Medicine entity + IMedicineService interface
│       ├── Medicine.go
│       └── medicine_test.go
├── application/                     ← MIDDLE LAYER — depends ONLY on domain
│   └── usecases/
│       ├── auth/
│       │   ├── auth.go              ← IAuthUseCase + AuthUseCase (login, token refresh)
│       │   └── auth_test.go
│       ├── user/
│       │   ├── user.go              ← IUserUseCase + UserUseCase
│       │   └── user_test.go
│       └── medicine/
│           ├── medicine.go          ← IMedicineUseCase + MedicineUseCase
│           └── medicine_test.go
└── infrastructure/                  ← OUTERMOST LAYER — depends on domain + application
    ├── di/
    │   ├── application_context.go   ← ApplicationContext (all DI wiring)
    │   └── application_context_test.go
    ├── repository/psql/
    │   ├── psql_repository.go       ← InitPSQLDB() database connection
    │   ├── user/
    │   │   ├── user.go              ← UserRepositoryInterface + GORM Repository
    │   │   └── user_test.go
    │   └── medicine/
    │       ├── medicine.go          ← MedicineRepositoryInterface + GORM Repository
    │       └── medicine_test.go
    ├── rest/
    │   ├── controllers/
    │   │   ├── BindTools.go         ← BindJSON(), BindJSONMap() shared helpers
    │   │   ├── Utils.go             ← SortByDataRequest, FieldDateRangeDataRequest
    │   │   ├── auth/
    │   │   │   ├── Auth.go          ← IAuthController + login/refresh handlers
    │   │   │   ├── Auth_test.go
    │   │   │   └── Structures.go    ← Auth request/response structs
    │   │   ├── user/
    │   │   │   ├── User.go          ← IUserController + CRUD handlers
    │   │   │   ├── User_test.go
    │   │   │   └── Validation.go    ← User update validation
    │   │   └── medicine/
    │   │       ├── Medicines.go     ← IMedicineController + CRUD handlers
    │   │       ├── Medicines_test.go
    │   │       └── Validation.go    ← Medicine update validation
    │   ├── middlewares/             ← AuthJWTMiddleware(), ErrorHandler(), GinBodyLog, CommonHeaders
    │   └── routes/
    │       ├── routes.go            ← ApplicationRouter() — main router under /v1
    │       ├── auth.go              ← AuthRoutes()
    │       ├── user.go              ← UserRoutes()
    │       └── medicine.go          ← MedicineRoutes()
    ├── security/
    │   ├── jwt_service.go           ← IJWTService + JWTService (generate/verify tokens)
    │   └── jwt_test.go
    └── logger/                      ← Zap Logger wrapper
Test/                                ← Integration tests (CAPITAL T directory)
└── integration/
    └── main_test.go                 ← Cucumber/godog BDD tests
```

## Critical Dependency Rules

```
infrastructure ──→ application ──→ domain
     ↑                  ↑              ↑
  outermost          middle        innermost
```

1. **Domain layer has ZERO dependencies** on application or infrastructure
2. **Application layer depends ONLY on domain** (implements domain interfaces, receives repo interfaces)
3. **Infrastructure layer** can import everything
4. **Dependencies ALWAYS point INWARD** — never outward
5. When you add a new feature, verify no import violates these rules

## Dependency Injection

All dependencies are wired in `src/infrastructure/di/application_context.go`:

```go
// Production wiring
func SetupDependencies(logger) (*ApplicationContext, error) {
    db := InitPSQLDB(logger)          // Database connection
    jwtService := NewJWTService()      // JWT service

    // Repos (depend on DB + logger)
    userRepo := user.NewUserRepository(db, logger)
    medicineRepo := medicine.NewMedicineRepository(db, logger)

    // Use cases (depend on repos + logger)
    authUC := auth.NewAuthUseCase(userRepo, jwtService, logger)
    userUC := user.NewUserUseCase(userRepo, logger)
    medicineUC := medicine.NewMedicineUseCase(medicineRepo, logger)

    // Controllers (depend on use cases + logger)
    authCtrl := auth.NewAuthController(authUC, logger)
    userCtrl := user.NewUserController(userUC, logger)
    medicineCtrl := medicine.NewMedicineController(medicineUC, logger)

    return &ApplicationContext{...}
}

// Test wiring — accepts mock implementations
func NewTestApplicationContext(mockUserRepo, mockMedicineRepo, mockJWT, logger) *ApplicationContext
```

## Interface Naming & Definition Pattern

| Layer | Interface | Defined In | Example |
|-------|-----------|------------|---------|
| Domain | Service interface | `domain/<entity>/` | `IMedicineService` |
| Application | Use case interface | `usecases/<entity>/` | `IMedicineUseCase` |
| Infrastructure | Repository interface | `repository/psql/<entity>/` | `MedicineRepositoryInterface` |
| Infrastructure | Controller interface | `controllers/<entity>/` | `IMedicineController` |
| Infrastructure | Security interface | `security/` | `IJWTService` |

**Rules**:
- Service/Use case/Controller interfaces → prefix with `I`
- Repository interfaces → suffix with `RepositoryInterface`
- Constructors → `NewXxx()` returning the **interface type**, NEVER the concrete struct

## Constructor Pattern

```go
// ✅ CORRECT — returns interface type
func NewMedicineUseCase(repo medicine.MedicineRepositoryInterface, log *logger.Logger) IMedicineUseCase {
    return &MedicineUseCase{
        medicineRepository: repo,
        Logger:             log,
    }
}

// ❌ WRONG — returns concrete type
func NewMedicineUseCase(...) *MedicineUseCase { ... }
```

## Error Handling System

### Error Types (defined in `domain/errors/Errors.go`)

| ErrorType | HTTP Status | Default Message |
|-----------|-------------|-----------------|
| `NotFound` | 404 | "record not found" |
| `ValidationError` | 400 | "validation error" |
| `ResourceAlreadyExists` | 409 | "resource already exists" |
| `RepositoryError` | 500 | "error in repository operation" |
| `NotAuthenticated` | 401 | "not Authenticated" |
| `TokenGeneratorError` | 500 | "error in token generation" |
| `NotAuthorized` | 403 | "not authorized" |
| `UnknownError` | 500 | "something went wrong" |

### Creating Errors

```go
// Custom message
domainErrors.NewAppError(errors.New("medicine id is invalid"), domainErrors.ValidationError)

// Default message for type
domainErrors.NewAppErrorWithType(domainErrors.NotFound)
```

### In Controllers — delegate to ErrorHandler middleware

```go
// ✅ CORRECT
_ = ctx.Error(appError)
return

// ❌ WRONG — never manually set error responses
ctx.JSON(http.StatusNotFound, gin.H{"error": "not found"})
```

### In Repositories — wrap all DB errors

```go
if err == gorm.ErrRecordNotFound {
    return nil, domainErrors.NewAppErrorWithType(domainErrors.NotFound)
}
// Parse GORM duplicate key errors via JSON marshal/unmarshal
byteErr, _ := json.Marshal(tx.Error)
var gormErr domainErrors.GormErr
json.Unmarshal(byteErr, &gormErr)
if gormErr.Number == 1062 {
    return nil, domainErrors.NewAppErrorWithType(domainErrors.ResourceAlreadyExists)
}
```

## Structured Logging with Zap

Every struct that needs logging receives `*logger.Logger` via constructor injection.

### DO
```go
s.Logger.Info("Creating medicine", zap.String("name", med.Name))
s.Logger.Error("Failed to create", zap.Error(err), zap.Int("id", id))
s.Logger.Warn("Medicine not found", zap.Int("id", id))
s.Logger.Info("Pagination search", zap.Int("page", p), zap.Int("pageSize", ps))
```

### DON'T
```go
fmt.Println("creating medicine")          // ❌
log.Printf("error: %v", err)              // ❌
s.Logger.Info(fmt.Sprintf("id=%d", id))   // ❌ wasteful string formatting
```

### When to Log
- **Controller**: Request start, validation errors, success with result IDs
- **Use Case**: Operation start with key parameters
- **Repository**: DB operations, warnings for not-found, detailed error info

## Mapper Pattern

Three separate struct types exist per entity:

1. **Domain Entity** (`domain/<entity>/`) — pure business struct, no tags
2. **GORM Model** (`repository/psql/<entity>/`) — with `gorm:` tags
3. **Response Struct** (`controllers/<entity>/`) — with `json:` tags

Mappers convert between them:

```go
// Repository: GORM → Domain
func (m *Medicine) toDomainMapper() *domainMedicine.Medicine { ... }
func arrayToDomainMapper(models *[]Medicine) *[]domainMedicine.Medicine { ... }

// Controller: Domain → Response
func domainToResponseMapper(m *medicineDomain.Medicine) *ResponseMedicine { ... }
func arrayDomainToResponseMapper(m *[]medicineDomain.Medicine) *[]ResponseMedicine { ... }
```

**NEVER expose domain entities directly in API responses.**

## Database (GORM) Patterns

### Model with explicit table name
```go
type Medicine struct {
    ID          int       `gorm:"primaryKey"`
    Name        string    `gorm:"unique"`
    EANCode     string    `gorm:"unique"`
    CreatedAt   time.Time `gorm:"autoCreateTime:milli"`
    UpdatedAt   time.Time `gorm:"autoUpdateTime:milli"`
}
func (*Medicine) TableName() string { return "medicines" }
```

### Column Mapping (REQUIRED for dynamic queries)
```go
var ColumnsMedicineMapping = map[string]string{
    "id":          "id",
    "name":        "name",
    "eanCode":     "ean_code",    // camelCase → snake_case
    "createdAt":   "created_at",
}
```
Always resolve column names through this map. Never use user input directly in SQL.

## REST API Patterns

### Route Registration
- All routes under `/v1` via `router.Group("/v1")`
- Per-entity route files: `routes/medicine.go`, `routes/user.go`, `routes/auth.go`
- Protected routes use `middlewares.AuthJWTMiddleware()`

### Controller Handler Flow
1. Log operation start
2. Bind/validate request (`controllers.BindJSON` or `controllers.BindJSONMap`)
3. Map request → domain entity
4. Call service method
5. Handle error → `ctx.Error(err); return`
6. Map domain → response
7. Log success with IDs
8. `ctx.JSON(http.StatusOK, response)`

### Request Binding
```go
controllers.BindJSON(ctx, &request)     // Struct binding with validation tags
controllers.BindJSONMap(ctx, &reqMap)   // Map binding for partial updates
```

### Validation tags
```go
type NewMedicineRequest struct {
    Name string `json:"name" binding:"required"`
}
```
Separate update validation in `Validation.go` per controller.

## Testing

### Stack
- `testify` for assertions (`assert`, `require`) and mocks
- `go-sqlmock` for database mocking
- `godog/cucumber` for BDD integration tests

### Locations
- Unit tests: `*_test.go` alongside source files
- Integration tests: `Test/integration/` (capital T)
- Test DI: `NewTestApplicationContext(mockUserRepo, mockMedicineRepo, mockJWT, logger)`

### Naming: `Test<Method>_<Scenario>`
```
TestCreate_Success
TestCreate_DuplicateError
TestGetByID_NotFound
TestUpdate_ValidationError
```

### Running
```bash
make tests                          # go test -v ./Test/...
make tests-TestGetByID_Success      # specific test
make integration-test               # BDD integration tests
```

## Adding a New Entity

When adding a new entity, create these files in order:

1. `src/domain/<entity>/<entity>.go` — Entity struct + `I<Entity>Service` interface
2. `src/application/usecases/<entity>/<entity>.go` — `I<Entity>UseCase` + `<Entity>UseCase` + `New<Entity>UseCase()`
3. `src/infrastructure/repository/psql/<entity>/<entity>.go` — GORM model + `<Entity>RepositoryInterface` + `Repository` + mappers
4. `src/infrastructure/rest/controllers/<entity>/<Entity>.go` — Request/Response structs + `I<Entity>Controller` + `Controller` + mappers
5. `src/infrastructure/rest/controllers/<entity>/Validation.go` — Update validation
6. `src/infrastructure/rest/routes/<entity>.go` — Route group
7. Update `src/infrastructure/di/application_context.go` — Wire repo → use case → controller
8. Update `src/infrastructure/rest/routes/routes.go` — Register routes

## Code Quality

| Tool | Command | Purpose |
|------|---------|---------|
| golangci-lint | `golangci-lint run ./...` | Comprehensive linting |
| staticcheck | `staticcheck ./...` | Static analysis |
| go vet | `go vet ./...` | Go vet checks |
| trivy | `trivy fs .` | Security/vulnerability scanning |
| go fmt | `go fmt ./...` | Formatting |
| goimports | `goimports -w .` | Import organization |

Pre-commit hooks managed by `lefthook.yml` — runs all of the above plus `go build` and unit tests.

## DO NOT

- ❌ Import infrastructure packages from domain layer
- ❌ Use `fmt.Println` or standard `log` package — use injected Zap logger
- ❌ Return concrete types from constructors — return interfaces
- ❌ Put business logic in controllers — controllers handle HTTP only
- ❌ Hardcode column names in dynamic queries — use column mapping
- ❌ Skip error handling — all errors must be returned or passed to `ctx.Error()`
- ❌ Expose domain entities directly to API — use response structs with mappers
- ❌ Create new dependencies without wiring in `ApplicationContext`
- ❌ Use `errors.New()` for domain errors — use `AppError` types
- ❌ Log with string interpolation — use Zap structured fields

## DO

- ✅ Follow the existing entity pattern exactly when adding features
- ✅ Write tests for every layer (unit + integration)
- ✅ Use structured Zap logging at every layer
- ✅ Propagate errors via `AppError` system
- ✅ Keep domain entities pure (no tags, no external deps)
- ✅ Register everything through `ApplicationContext`
- ✅ Run `make lint` before committing
- ✅ Target ≥ 80% test coverage
