# GitHub Copilot Instructions — Ecommerce Microservice Go

## Project

Go microservices with Clean Architecture. Module: `github.com/gbrayhan/microservices-go`. Go 1.24.2+. Entry: `main.go` (root).

**Stack**: Gin-Gonic (HTTP), GORM (ORM), PostgreSQL, Zap (logging), JWT (auth), testify + godog (testing).

## Architecture

- `src/domain/` — Entities, service interfaces (`IUserService`, `IMedicineService`), `AppError` types. No external deps.
- `src/application/usecases/` — Use cases implement domain interfaces. Depend on domain only.
- `src/infrastructure/` — Controllers, repositories (GORM/PostgreSQL), DI, security (JWT), logger (Zap), routes.
- `Test/integration/` — Integration tests (capital T directory).
- Dependencies point inward: infrastructure → application → domain.

## Code Patterns

- **Interfaces**: Prefix with `I`. Defined in their respective packages.
- **Constructors**: `NewXxx()` returning interface type. Example: `func NewMedicineUseCase(...) IMedicineUseCase`.
- **DI**: `src/infrastructure/di/application_context.go`. Logger injected into all layers.
- **Errors**: `domain/errors.AppError` with typed errors (`NotFound`, `ValidationError`, etc.). Use `ctx.Error(err)` in controllers.
- **Logging**: Zap structured fields (`zap.Int()`, `zap.String()`, `zap.Error()`). Never use `fmt.Println` or `log`.
- **Mappers**: Repository → Domain: `toDomainMapper()`. Domain → Response: `domainToResponseMapper()`.
- **Routes**: Under `/v1`, per-entity route files. Protected routes use `AuthJWTMiddleware()`.
- **GORM**: Models with `TableName()`, `ColumnsMedicineMapping` for safe column resolution.
- **Validation**: `binding:"required"` tags + separate `Validation.go` files.
- **Testing**: `testify`, `go-sqlmock`, `godog/cucumber`. Tests: `go test -v ./Test/...`

## Rules

- Never import infrastructure from domain
- Never use `fmt.Println` or `log` — use injected Zap logger
- Never return concrete types from constructors
- Never put business logic in controllers
- Always propagate errors via `AppError`
- Always use structured logging
- Follow existing patterns when adding new entities
