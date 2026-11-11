SHELL := /bin/bash
BACKEND := backend
FRONTEND := frontend
YARN ?= corepack yarn

.PHONY: help test backend-test frontend-test frontend-typecheck frontend-build lint migrations-lint migrations-checksum fixtures test-upgrade release-check

help: ## Show available targets
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-24s %s\n", $$1, $$2}'

test: backend-test frontend-test ## Run backend and frontend unit tests

backend-test: ## Run all Go tests (includes upgrade rehearsal)
	cd $(BACKEND) && go test ./...

frontend-test: ## Run frontend unit tests
	cd $(FRONTEND) && $(YARN) test

frontend-typecheck: ## Type-check the frontend and E2E sources
	cd $(FRONTEND) && $(YARN) tsc --noEmit
	cd $(FRONTEND) && $(YARN) tsc -p tsconfig.node.json --noEmit

frontend-build: ## Build the production frontend
	cd $(FRONTEND) && REACT_APP_API_URL="" $(YARN) build

lint: migrations-lint frontend-typecheck ## Run repository checks

migrations-lint: ## Verify migration checksums, gaps and destructive statements
	cd $(BACKEND) && go run ./cmd/migrationtool lint

migrations-checksum: ## Record a newly added migration in migrations.lock
	cd $(BACKEND) && go run ./cmd/migrationtool checksum -update
	cd $(BACKEND) && go run ./cmd/migrationtool lint

fixtures: ## Rebuild committed upgrade fixtures
	$(MAKE) -C $(BACKEND) fixtures

test-upgrade: ## Rehearse upgrades from every committed fixture
	$(MAKE) -C $(BACKEND) test-upgrade

release-check: migrations-lint test-upgrade backend-test frontend-typecheck frontend-test frontend-build ## Full pre-release gate
	@echo "Release checks passed. Remember to build the Windows/Linux artifacts with scripts/build-windows-exe.sh."
