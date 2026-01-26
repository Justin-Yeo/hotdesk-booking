# Load environment variables
include .env
export

# Database URL for migrations
DB_URL := $(DATABASE_URL)

.PHONY: help migrate-up migrate-down migrate-status migrate-create migrate-reset

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'

migrate-up: ## Run all pending migrations
	@echo "Running migrations..."
	goose -dir migrations postgres "$(DB_URL)" up

migrate-down: ## Rollback the last migration
	@echo "Rolling back last migration..."
	goose -dir migrations postgres "$(DB_URL)" down

migrate-status: ## Show migration status
	@echo "Migration status:"
	goose -dir migrations postgres "$(DB_URL)" status

migrate-create: ## Create a new migration file (usage: make migrate-create NAME=create_users_table)
	@if [ -z "$(NAME)" ]; then \
		echo "Error: NAME is required. Usage: make migrate-create NAME=create_users_table"; \
		exit 1; \
	fi
	@echo "Creating migration: $(NAME)"
	goose -dir migrations create $(NAME) sql

migrate-reset: ## Reset database (down all, then up all)
	@echo "Resetting database..."
	goose -dir migrations postgres "$(DB_URL)" reset

migrate-version: ## Show current migration version
	@goose -dir migrations postgres "$(DB_URL)" version
