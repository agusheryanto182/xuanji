.PHONY: help install dev build start lint \
	docker-build docker-up docker-down docker-logs docker-restart docker-clean

COMPOSE := docker compose

help: ## Show available targets
	@grep -E '^[a-zA-Z0-9_-]+:.*##' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*## "}; {printf "  \033[36m%-18s\033[0m %s\n", $$1, $$2}'

# --- Local (Node) ---

install: ## Install npm dependencies
	npm ci

dev: ## Run Next.js dev server (localhost:3000)
	npm run dev

build: ## Production build (standalone output in .next/)
	npm run build

start: ## Start production server (run build first)
	npm run start

lint: ## Run ESLint
	npm run lint

# --- Docker ---

docker-build: ## Build the production Docker image
	$(COMPOSE) build

docker-up: ## Start the app in Docker (detached)
	$(COMPOSE) up -d

docker-down: ## Stop Docker containers
	$(COMPOSE) down

docker-logs: ## Follow container logs
	$(COMPOSE) logs -f

docker-restart: docker-down docker-up ## Restart Docker stack

docker-clean: docker-down ## Remove containers, image, and build cache
	$(COMPOSE) down --rmi local --volumes
	docker builder prune -f
