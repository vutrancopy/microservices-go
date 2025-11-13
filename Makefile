# Ride Hailing Backend - Makefile
# =====================================

.PHONY: help build run stop clean test migrate-up migrate-down logs lint coverage docker-build docker-up docker-down docker-logs

# Default target
.DEFAULT_GOAL := help

# Colors for output
BLUE := \033[0;34m
GREEN := \033[0;32m
YELLOW := \033[0;33m
RED := \033[0;31m
NC := \033[0m # No Color

# Variables
DOCKER_COMPOSE := docker-compose
GO := go
APP_NAME := ride-hailing-api
COVERAGE_FILE := coverage.out

## help: Display this help message
help:
	@echo "$(BLUE)Ride Hailing Backend - Available Commands$(NC)"
	@echo "==========================================="
	@awk 'BEGIN {FS = ":.*##"; printf "\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  $(GREEN)%-20s$(NC) %s\n", $$1, $$2 } /^##@/ { printf "\n$(YELLOW)%s$(NC)\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ Docker Commands

## docker-build: Build Docker images
docker-build:
	@echo "$(BLUE)Building Docker images...$(NC)"
	$(DOCKER_COMPOSE) build

## build: Alias for docker-build
build: docker-build

## docker-up: Start all services with Docker Compose
docker-up:
	@echo "$(BLUE)Starting Docker Compose services...$(NC)"
	$(DOCKER_COMPOSE) up -d
	@echo "$(GREEN)Services started successfully!$(NC)"
	@echo "API: http://localhost:8080"
	@echo "Adminer: http://localhost:8081"
	@echo "PostgreSQL: localhost:5433"

## run: Start all services (same as docker-up)
run: docker-up

## docker-down: Stop and remove all containers
docker-down:
	@echo "$(YELLOW)Stopping Docker Compose services...$(NC)"
	$(DOCKER_COMPOSE) down

## stop: Stop all services
stop: docker-down

## docker-logs: Show logs from all services
docker-logs:
	$(DOCKER_COMPOSE) logs -f

## logs: Show logs (alias for docker-logs)
logs: docker-logs

## docker-restart: Restart all services
docker-restart: docker-down docker-up

## clean: Remove all containers, volumes, and images
clean:
	@echo "$(RED)Cleaning up Docker resources...$(NC)"
	$(DOCKER_COMPOSE) down -v --remove-orphans
	@echo "$(GREEN)Cleanup complete!$(NC)"

##@ Development Commands

## dev: Run the application locally (without Docker)
dev:
	@echo "$(BLUE)Running application locally...$(NC)"
	$(GO) run main.go

## install: Install Go dependencies
install:
	@echo "$(BLUE)Installing dependencies...$(NC)"
	$(GO) mod download
	$(GO) mod tidy

## fmt: Format Go code
fmt:
	@echo "$(BLUE)Formatting code...$(NC)"
	$(GO) fmt ./...

## vet: Run go vet
vet:
	@echo "$(BLUE)Running go vet...$(NC)"
	$(GO) vet ./...

## lint: Run linter
lint: fmt vet
	@echo "$(BLUE)Running linter...$(NC)"
	@if command -v golangci-lint > /dev/null; then \
		golangci-lint run ./...; \
	else \
		echo "$(YELLOW)golangci-lint not installed. Skipping...$(NC)"; \
	fi

##@ Testing Commands

## test: Run all tests
test:
	@echo "$(BLUE)Running tests...$(NC)"
	$(GO) test -v ./...

## test-unit: Run unit tests only
test-unit:
	@echo "$(BLUE)Running unit tests...$(NC)"
	$(GO) test -v -short ./...

## test-integration: Run integration tests
test-integration:
	@echo "$(BLUE)Running integration tests...$(NC)"
	@if [ -f scripts/run-integration-test.bash ]; then \
		./scripts/run-integration-test.bash; \
	else \
		echo "$(YELLOW)Integration test script not found$(NC)"; \
	fi

## coverage: Run tests with coverage
coverage:
	@echo "$(BLUE)Running tests with coverage...$(NC)"
	$(GO) test -coverprofile=$(COVERAGE_FILE) -covermode=atomic ./...
	$(GO) tool cover -html=$(COVERAGE_FILE) -o coverage.html
	@echo "$(GREEN)Coverage report generated: coverage.html$(NC)"

## coverage-report: Show coverage report in terminal
coverage-report:
	@echo "$(BLUE)Generating coverage report...$(NC)"
	$(GO) test -coverprofile=$(COVERAGE_FILE) ./...
	$(GO) tool cover -func=$(COVERAGE_FILE)

##@ Database Commands

## migrate-up: Run database migrations (inside container)
migrate-up:
	@echo "$(BLUE)Running database migrations...$(NC)"
	@echo "$(YELLOW)Note: Migrations are automatically run on container startup$(NC)"
	$(DOCKER_COMPOSE) exec api ./microservice migrate-up

## migrate-down: Rollback database migrations (inside container)
migrate-down:
	@echo "$(YELLOW)Rolling back database migrations...$(NC)"
	$(DOCKER_COMPOSE) exec api ./microservice migrate-down

## db-shell: Connect to PostgreSQL database
db-shell:
	@echo "$(BLUE)Connecting to database...$(NC)"
	$(DOCKER_COMPOSE) exec db psql -U postgres -d ride_hailing_db

## db-reset: Reset database (WARNING: destroys all data)
db-reset:
	@echo "$(RED)WARNING: This will destroy all database data!$(NC)"
	@read -p "Are you sure? [y/N] " -n 1 -r; \
	echo; \
	if [[ $$REPLY =~ ^[Yy]$$ ]]; then \
		$(DOCKER_COMPOSE) down -v; \
		$(DOCKER_COMPOSE) up -d; \
		echo "$(GREEN)Database reset complete!$(NC)"; \
	fi

##@ Utility Commands

## status: Show status of all services
status:
	@echo "$(BLUE)Service Status:$(NC)"
	$(DOCKER_COMPOSE) ps

## shell: Open shell in API container
shell:
	@echo "$(BLUE)Opening shell in API container...$(NC)"
	$(DOCKER_COMPOSE) exec api sh

## health: Check API health
health:
	@echo "$(BLUE)Checking API health...$(NC)"
	@curl -f http://localhost:8080/v1/user/ || echo "$(RED)API is not responding$(NC)"

## env: Copy .env.example to .env if not exists
env:
	@if [ ! -f .env ]; then \
		echo "$(BLUE)Creating .env file from .env.example...$(NC)"; \
		cp .env.example .env; \
		echo "$(GREEN).env file created! Please review and update values.$(NC)"; \
	else \
		echo "$(YELLOW).env file already exists$(NC)"; \
	fi

## version: Show Go and Docker versions
version:
	@echo "$(BLUE)Version Information:$(NC)"
	@echo "Go version: $$(go version)"
	@echo "Docker version: $$(docker --version)"
	@echo "Docker Compose version: $$(docker-compose --version)"
