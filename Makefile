include .env
export $(shell sed 's/=.*//' .env)

start:
	@go run main.go
lint:
	@golangci-lint run ./...
tests:
	@go test -v ./Test/...
tests-%:
	@go test -v ./Test/... -run=$(shell echo $* | sed 's/_/./g')
testsum:
	@cd Test && gotestsum --format testname
integration-test:
	@./scripts/run-integration-test.bash
swagger:
	@swag init --parseDependency --parseInternal
swagger-docker:
	@docker run --rm -v ./:/code ghcr.io/swaggo/swag:latest init --parseDependency --parseInternal
migration-%:
	@migrate create -ext sql -dir src/database/migrations create-table-$(subst :,_,$*)
migration-docker-%:
	@docker run --rm -v ./src/database/migrations:/migrations migrate/migrate create -ext sql -dir /migrations $(subst :,_,$*)
migrate-up:
	@migrate -database "postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=disable" -path src/database/migrations up
migrate-down:
	@migrate -database "postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=disable" -path src/database/migrations down
migrate-docker-up:
	@docker run --rm --network host -v ./src/database/migrations:/migrations migrate/migrate -path=/migrations/ -database postgres://$(DB_USER):$(DB_PASSWORD)@localhost:$(DB_PORT)/$(DB_NAME)?sslmode=disable up
migrate-docker-down:
	@docker run --rm --network host -v ./src/database/migrations:/migrations migrate/migrate -path=/migrations/ -database postgres://$(DB_USER):$(DB_PASSWORD)@localhost:$(DB_PORT)/$(DB_NAME)?sslmode=disable down 1
migrate-docker-production-up:
	@docker run --rm --network host -v ./src/database/migrations:/migrations migrate/migrate -path=/migrations/ -database postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=disable up
migrate-docker-production-down:
	@docker run --rm --network host -v ./src/database/migrations:/migrations migrate/migrate -path=/migrations/ -database postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=disable down 1
docker:
	@docker compose up --build -d
docker-test:
	@docker compose up -d && make tests
docker-down:
	@docker compose down --rmi all --volumes --remove-orphans
docker-cache:
	@docker builder prune -f
