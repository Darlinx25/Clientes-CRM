.PHONY: migrate-up migrate-down migrate-create migrate-status migrate-version migrations-lint migrations-checksum fixtures test-upgrade help

# Database path
DB_PATH ?= meerkat.db
MIGRATIONS_DIR = ./database/migrations

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-20s %s\n", $$1, $$2}'

migrate-up: ## Run all pending migrations
	@echo "Running migrations on $(DB_PATH)..."
	@go run cmd/migrate/main.go up
	@echo "Migrations applied successfully!"

migrate-down: ## Rollback the last migration
	@echo "Rolling back last migration on $(DB_PATH)..."
	@go run cmd/migrate/main.go down
	@echo "Migration rolled back successfully!"

migrate-create: ## Create a new migration file (usage: make migrate-create NAME=create_users_table)
	@if [ -z "$(NAME)" ]; then \
		echo "Error: NAME is required. Usage: make migrate-create NAME=your_migration_name"; \
		exit 1; \
	fi
	@migrate create -ext sql -dir $(MIGRATIONS_DIR) -seq $(NAME)
	@echo "Migration files created in $(MIGRATIONS_DIR)!"

migrate-status: ## Show current migration version
	@echo "Current migration status for $(DB_PATH):"
	@go run cmd/migrate/main.go version

migrate-version: ## Show available migration files
	@echo "Available migration files:"
	@ls -1 $(MIGRATIONS_DIR)/*.sql 2>/dev/null || echo "No migration files found"

migrations-lint: ## Check migration checksums, version gaps and destructive statements
	@go run ./cmd/migrationtool lint

migrations-checksum: ## Add newly created migrations to migrations.lock (never rewrites history)
	@go run ./cmd/migrationtool checksum -update
	@go run ./cmd/migrationtool lint

fixtures: ## Rebuild the upgrade-rehearsal fixtures from committed migrations
	@go run ./cmd/mkfixture -version 23 -out database/testdata/fixtures/v000023.db -force
	@go run ./cmd/mkfixture -version 39 -out database/testdata/fixtures/v000039.db -force

test-upgrade: ## Run the upgrade rehearsal against every committed fixture
	@go test ./database -run 'TestUpgrade|TestDirty|TestNewDatabase|TestNewer|TestMigration|TestLatest' -v

# Development helpers
dev: ## Build and run the server in development mode
	@go build -o meerkat . && ./meerkat

test: ## Run all tests
	@go test ./... -v

build: ## Build the application
	@go build -o meerkat .

clean: ## Clean build artifacts
	@rm -f meerkat
	@echo "Build artifacts cleaned!"
