ifneq ($(wildcard .env),)
include .env
export
else
$(warning WARNING: .env file not found! Using .env.example)
include .env.example
export
endif

BASE_STACK := docker compose -f docker-compose.yml
INTEGRATION_TEST_STACK := $(BASE_STACK) -f docker-compose-integration-test.yml

POSTGRES_CONTAINER := redis_playground_postgres
REDIS_CONTAINER := redis_playground_redis

POSTGRES_USER := user
POSTGRES_DB := redis_playground_db

# ==============================================================================
# Help
# ==============================================================================

.PHONY: help

help: ## Display help
	@awk 'BEGIN {FS=":.*##"; printf "\nUsage:\n  make \033[36m<TARGET>\033[0m\n"} \
	/^[a-zA-Z_-]+:.*?##/ {printf "  \033[36m%-28s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# ==============================================================================
# Docker
# ==============================================================================

compose-up: ## Start PostgreSQL and Redis
	$(BASE_STACK) up --build -d postgres redis
.PHONY: compose-up

compose-up-all: ## Start all services
	$(BASE_STACK) up --build -d
.PHONY: compose-up-all

compose-down: ## Stop all services
	$(BASE_STACK) down --remove-orphans
.PHONY: compose-down

compose-clean: ## Stop containers and remove volumes
	$(BASE_STACK) down -v --remove-orphans
.PHONY: compose-clean

compose-restart: ## Restart all services
	$(BASE_STACK) restart
.PHONY: compose-restart

compose-logs: ## Follow docker compose logs (example: make compose-logs postgres)
	$(BASE_STACK) logs -f $(filter-out $@,$(MAKECMDGOALS))
.PHONY: compose-logs

compose-ps: ## Show running containers
	$(BASE_STACK) ps
.PHONY: compose-ps

compose-up-integration-test: ## Run integration tests
	$(INTEGRATION_TEST_STACK) up --build \
		--abort-on-container-exit \
		--exit-code-from integration-test; \
	exit_code=$$?; \
	$(INTEGRATION_TEST_STACK) down --remove-orphans; \
	exit $$exit_code
.PHONY: compose-up-integration-test

# ==============================================================================
# Docker CLI
# ==============================================================================

redis-cli: ## Open Redis CLI
	docker exec -it $(REDIS_CONTAINER) redis-cli
.PHONY: redis-cli

psql: ## Open PostgreSQL CLI
	docker exec -it $(POSTGRES_CONTAINER) \
		psql -U $(POSTGRES_USER) -d $(POSTGRES_DB)
.PHONY: psql

# ==============================================================================
# Development
# ==============================================================================

deps: ## Download & verify dependencies
	go mod tidy
	go mod verify
.PHONY: deps

deps-audit: ## Check dependency vulnerabilities
	govulncheck ./...
.PHONY: deps-audit

swag-v1: ## Generate Swagger
	swag init --parseDependency -g internal/controller/restapi/router.go
.PHONY: swag-v1

fix-diff: ## Show go fix diff
	go fix -diff ./...
.PHONY: fix-diff

format: ## Format source code
	go fix ./...
	gofumpt -l -w .
	gci write . --skip-generated -s standard -s default
.PHONY: format

run: deps swag-v1 ## Run application
	go mod download && \
	CGO_ENABLED=0 go run -tags migrate ./cmd/app
.PHONY: run

air: deps swag-v1 ## Run application with Air
	go mod download && \
	air -c .air.toml
.PHONY: air

seed: ## Seed database
	go run ./cmd/seed
.PHONY: seed

# ==============================================================================
# Linters
# ==============================================================================

linter-golangci: ## Run golangci-lint
	golangci-lint run
.PHONY: linter-golangci

linter-hadolint: ## Run hadolint
	git ls-files --exclude='Dockerfile*' --ignored | xargs hadolint
.PHONY: linter-hadolint

linter-dotenv: ## Run dotenv-linter
	dotenv-linter
.PHONY: linter-dotenv

# ==============================================================================
# Testing
# ==============================================================================

test: ## Run unit tests
	go test -v -race -covermode atomic -coverprofile=coverage.txt ./internal/... ./pkg/...
.PHONY: test

integration-test: ## Run integration tests
	go clean -testcache
	go test -v ./integration-test/...
.PHONY: integration-test

mock: ## Generate mocks
	mockgen -source ./internal/repo/contracts.go -package usecase_test > ./internal/usecase/mocks_repo_test.go
	mockgen -source ./internal/usecase/contracts.go -package usecase_test > ./internal/usecase/mocks_usecase_test.go
	mockgen -source ./pkg/logger/logger.go -package usecase_test > ./internal/usecase/mocks_logger_test.go
.PHONY: mock

# ==============================================================================
# Migration
# ==============================================================================

migrate-create: ## Create migration
	migrate create -ext sql -dir migrations '$(word 2,$(MAKECMDGOALS))'
.PHONY: migrate-create

migrate-up: ## Apply migrations
	migrate -path migrations -database '$(PG_URL)?sslmode=disable' up
.PHONY: migrate-up

migrate-down: ## Rollback all migrations
	migrate -path migrations -database '$(PG_URL)?sslmode=disable' down
.PHONY: migrate-down

migrate-down-step: ## Rollback N migrations
	migrate -path migrations -database '$(PG_URL)?sslmode=disable' down $(STEP)
.PHONY: migrate-down-step

# ==============================================================================
# Tools
# ==============================================================================

bin-deps: ## Install development tools
	go install tool
	go install -tags postgres github.com/golang-migrate/migrate/v4/cmd/migrate
.PHONY: bin-deps

validate: format linter-golangci test ## Format, lint and test
.PHONY: validate

pre-commit: deps swag-v1 mock validate ## Run pre-commit checks
.PHONY: pre-commit

# Allow extra arguments (e.g. make compose-logs postgres)
%:
	@: