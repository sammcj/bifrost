# Makefile for Bifrost

# Variables
HOST ?= localhost
PORT ?= 8080
APP_DIR ?= 
PROMETHEUS_LABELS ?=
LOG_STYLE ?= json
LOG_LEVEL ?= info

# Colors for output
RED=\033[0;31m
GREEN=\033[0;32m
YELLOW=\033[1;33m
BLUE=\033[0;34m
CYAN=\033[0;36m
NC=\033[0m # No Color

.PHONY: all help dev build-ui build run install-air clean test install-ui setup-workspace work-init work-clean docs build-docker-image cleanup-enterprise deploy-to-fly-io

all: help

# Default target
help: ## Show this help message
	@echo "$(BLUE)Bifrost Development - Available Commands:$(NC)"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  $(GREEN)%-15s$(NC) %s\n", $$1, $$2}'
	@echo ""
	@echo "$(YELLOW)Environment Variables:$(NC)"
	@echo "  HOST              Server host (default: localhost)"
	@echo "  PORT              Server port (default: 8080)"
	@echo "  PROMETHEUS_LABELS Labels for Prometheus metrics"
	@echo "  LOG_STYLE Logger output format: json|pretty (default: json)"
	@echo "  LOG_LEVEL Logger level: debug|info|warn|error (default: info)"
	@echo "  APP_DIR           App data directory inside container (default: /app/data)"

cleanup-enterprise: ## Clean up enterprise directories if present
	@echo "$(GREEN)Cleaning up enterprise...$(NC)"
	@if [ -d "ui/app/enterprise" ]; then rm -rf ui/app/enterprise; fi
	@echo "$(GREEN)Enterprise cleaned up$(NC)"

install-ui: cleanup-enterprise
	@which node > /dev/null || (echo "$(RED)Error: Node.js is not installed. Please install Node.js first.$(NC)" && exit 1)
	@which npm > /dev/null || (echo "$(RED)Error: npm is not installed. Please install npm first.$(NC)" && exit 1)
	@echo "$(GREEN)Node.js and npm are installed$(NC)"
	@cd ui && npm install
	@which next > /dev/null || (echo "$(YELLOW)Installing nextjs...$(NC)" && npm install -g next)
	@echo "$(GREEN)UI deps are in sync$(NC)"

install-air: ## Install air for hot reloading (if not already installed)
	@which air > /dev/null || (echo "$(YELLOW)Installing air for hot reloading...$(NC)" && go install github.com/air-verse/air@latest)
	@echo "$(GREEN)Air is ready$(NC)"

dev: install-ui install-air setup-workspace ## Start complete development environment (UI + API with proxy)
	@echo "$(GREEN)Starting Bifrost complete development environment...$(NC)"
	@echo "$(YELLOW)This will start:$(NC)"
	@echo "  1. UI development server (localhost:3000)"
	@echo "  2. API server with UI proxy (localhost:$(PORT))"
	@echo "$(CYAN)Access everything at: http://localhost:$(PORT)$(NC)"
	@echo ""
	@echo "$(YELLOW)Starting UI development server...$(NC)"
	@cd ui && npm run dev &
	@sleep 3
	@echo "$(YELLOW)Starting API server with UI proxy...$(NC)"
	@$(MAKE) setup-workspace >/dev/null
	@cd transports/bifrost-http && BIFROST_UI_DEV=true air -c .air.toml -- \
		-host "$(HOST)" \
		-port "$(PORT)" \
		-log-style "$(LOG_STYLE)" \
		-log-level "$(LOG_LEVEL)" \
		$(if $(PROMETHEUS_LABELS),-prometheus-labels "$(PROMETHEUS_LABELS)") \
		$(if $(APP_DIR),-app-dir "$(APP_DIR)")

build-ui: install-ui ## Build ui
	@echo "$(GREEN)Building ui...$(NC)"
	@rm -rf ui/.next
	@cd ui && npm run build && npm run copy-build

build: build-ui ## Build bifrost-http binary
	@echo "$(GREEN)Building bifrost-http...$(NC)"
	@cd transports/bifrost-http && GOWORK=off go build -o ../../tmp/bifrost-http .
	@echo "$(GREEN)Built: tmp/bifrost-http$(NC)"

build-docker-image: build-ui ## Build Docker image
	@echo "$(GREEN)Building Docker image...$(NC)"
	$(eval GIT_SHA=$(shell git rev-parse --short HEAD))
	@docker build -f transports/Dockerfile -t bifrost -t bifrost:$(GIT_SHA) -t bifrost:latest .
	@echo "$(GREEN)Docker image built: bifrost, bifrost:$(GIT_SHA), bifrost:latest$(NC)"

docker-run: ## Run Docker container
	@echo "$(GREEN)Running Docker container...$(NC)"
	@docker run -e APP_PORT=$(PORT) -e APP_HOST=0.0.0.0 -p $(PORT):$(PORT) -e LOG_LEVEL=$(LOG_LEVEL) -e LOG_STYLE=$(LOG_STYLE) -v $(shell pwd):/app/data  bifrost

docs: ## Prepare local docs
	@echo "$(GREEN)Preparing local docs...$(NC)"
	@cd docs && npx --yes mintlify@latest dev

run: build ## Build and run bifrost-http (no hot reload)
	@echo "$(GREEN)Running bifrost-http...$(NC)"
	@./tmp/bifrost-http \
		-host "$(HOST)" \
		-port "$(PORT)" \
		-log-style "$(LOG_STYLE)" \
		-log-level "$(LOG_LEVEL)" \
		$(if $(PROMETHEUS_LABELS),-prometheus-labels "$(PROMETHEUS_LABELS)")
		$(if $(APP_DIR),-app-dir "$(APP_DIR)")

clean: ## Clean build artifacts and temporary files
	@echo "$(YELLOW)Cleaning build artifacts...$(NC)"
	@rm -rf tmp/
	@rm -f transports/bifrost-http/build-errors.log
	@rm -rf transports/bifrost-http/tmp/
	@echo "$(GREEN)Clean complete$(NC)"

test: ## Run tests for bifrost-http
	@echo "$(GREEN)Running bifrost-http tests...$(NC)"
	@cd transports/bifrost-http && GOWORK=off go test -v ./...

test-core: ## Run core tests
	@echo "$(GREEN)Running core tests...$(NC)"
	@cd core && go test -v ./...

test-plugins: ## Run plugin tests
	@echo "$(GREEN)Running plugin tests...$(NC)"
	@cd plugins && find . -name "*.go" -path "*/tests/*" -o -name "*_test.go" | head -1 > /dev/null && \
		for dir in $$(find . -name "*_test.go" -exec dirname {} \; | sort -u); do \
			echo "Testing $$dir..."; \
			cd $$dir && go test -v ./... && cd - > /dev/null; \
		done || echo "No plugin tests found"

test-all: test-core test-plugins test ## Run all tests

# Quick start with example config
quick-start: ## Quick start with example config and maxim plugin
	@echo "$(GREEN)Quick starting Bifrost with example configuration...$(NC)"
	@$(MAKE) dev

# Linting and formatting
lint: ## Run linter for Go code
	@echo "$(GREEN)Running golangci-lint...$(NC)"
	@golangci-lint run ./...

fmt: ## Format Go code
	@echo "$(GREEN)Formatting Go code...$(NC)"
	@gofmt -s -w .
	@goimports -w .

# Workspace helpers
setup-workspace: ## Set up Go workspace with all local modules for development
	@echo "$(GREEN)Setting up Go workspace for local development...$(NC)"
	@echo "$(YELLOW)Cleaning existing workspace...$(NC)"
	@rm -f go.work go.work.sum || true
	@echo "$(YELLOW)Initializing new workspace...$(NC)"
	@go work init ./core ./framework ./transports
	@echo "$(YELLOW)Adding plugin modules...$(NC)"
	@for plugin_dir in ./plugins/*/; do \
		if [ -d "$$plugin_dir" ] && [ -f "$$plugin_dir/go.mod" ]; then \
			echo "  Adding plugin: $$(basename $$plugin_dir)"; \
			go work use "$$plugin_dir"; \
		fi; \
	done
	@echo "$(YELLOW)Syncing workspace...$(NC)"
	@go work sync
	@echo "$(GREEN)✓ Go workspace ready with all local modules$(NC)"
	@echo ""
	@echo "$(CYAN)Local modules in workspace:$(NC)"
	@go list -m all | grep "github.com/maximhq/bifrost" | grep -v " v" | sed 's/^/  ✓ /'
	@echo ""
	@echo "$(CYAN)Remote modules (no local version):$(NC)"
	@go list -m all | grep "github.com/maximhq/bifrost" | grep " v" | sed 's/^/  → /'
	@echo ""
	@echo "$(YELLOW)Note: go.work files are not committed to version control$(NC)"

work-init: ## Create local go.work to use local modules for development (legacy)
	@echo "$(YELLOW)⚠️  work-init is deprecated, use 'make setup-workspace' instead$(NC)"
	@$(MAKE) setup-workspace

work-clean: ## Remove local go.work
	@rm -f go.work go.work.sum || true
	@echo "$(GREEN)Removed local go.work files$(NC)"


# Deployment scripts for different platforms

deploy-to-fly-io: ## Deploy to Fly.io (Usage: make deploy-to-fly-io APP_NAME=your-app-name)
	@echo "$(BLUE)Starting Fly.io deployment...$(NC)"
	@echo ""
	@# Check if APP_NAME is provided
	@if [ -z "$(APP_NAME)" ]; then \
		echo "$(RED)Error: APP_NAME is required$(NC)"; \
		echo "$(YELLOW)Usage: make deploy-to-fly-io APP_NAME=your-app-name$(NC)"; \
		exit 1; \
	fi
	@echo "$(YELLOW)Checking prerequisites...$(NC)"
	@# Check if docker is installed
	@which docker > /dev/null || (echo "$(RED)Error: Docker is not installed. Please install Docker first.$(NC)" && exit 1)
	@echo "$(GREEN)✓ Docker is installed$(NC)"
	@# Check if flyctl is installed
	@which flyctl > /dev/null || (echo "$(RED)Error: flyctl is not installed. Please install flyctl first.$(NC)" && exit 1)
	@echo "$(GREEN)✓ flyctl is installed$(NC)"
	@# Check if app exists on Fly.io
	@flyctl status -a $(APP_NAME) > /dev/null 2>&1 || (echo "$(RED)Error: App '$(APP_NAME)' not found on Fly.io$(NC)" && echo "$(YELLOW)Create the app first with: flyctl launch --name $(APP_NAME)$(NC)" && exit 1)
	@echo "$(GREEN)✓ App '$(APP_NAME)' exists on Fly.io$(NC)"
	@echo ""
	@# Check if fly.toml exists, create temp if needed
	@if [ -f "fly.toml" ]; then \
		echo "$(GREEN)✓ Using existing fly.toml$(NC)"; \
	else \
		echo "$(YELLOW)fly.toml not found in current directory$(NC)"; \
		echo "$(CYAN)Would you like to create a temporary fly.toml with 2 vCPU configuration?$(NC)"; \
		echo "$(CYAN)(It will be removed after deployment)$(NC)"; \
		printf "Create temporary fly.toml? [y/N]: "; read response; \
		case "$$response" in \
			[yY][eE][sS]|[yY]) \
				echo "$(YELLOW)Creating temporary fly.toml with 2 vCPU configuration...$(NC)"; \
				echo "app = '$(APP_NAME)'" > fly.toml; \
				echo "primary_region = 'iad'" >> fly.toml; \
				echo "" >> fly.toml; \
				echo "[build]" >> fly.toml; \
				echo "  image = 'registry.fly.io/$(APP_NAME):latest'" >> fly.toml; \
				echo "" >> fly.toml; \
				echo "[http_service]" >> fly.toml; \
				echo "  internal_port = 8080" >> fly.toml; \
				echo "  force_https = true" >> fly.toml; \
				echo "  auto_stop_machines = true" >> fly.toml; \
				echo "  auto_start_machines = true" >> fly.toml; \
				echo "  min_machines_running = 0" >> fly.toml; \
				echo "" >> fly.toml; \
				echo "[[vm]]" >> fly.toml; \
				echo "  memory = '2gb'" >> fly.toml; \
				echo "  cpu_kind = 'shared'" >> fly.toml; \
				echo "  cpus = 2" >> fly.toml; \
				echo "$(GREEN)✓ Created temporary fly.toml with 2 vCPU configuration$(NC)"; \
				touch .fly.toml.tmp.marker; \
				;; \
			*) \
				echo "$(RED)Deployment cancelled. Please create a fly.toml file or run 'flyctl launch' first.$(NC)"; \
				exit 1; \
				;; \
		esac; \
	fi
	@echo ""
	@echo "$(YELLOW)Building Docker image...$(NC)"
	@$(MAKE) build-docker-image
	@echo ""
	@echo "$(YELLOW)Tagging image for Fly.io registry...$(NC)"
	@docker tag bifrost:latest registry.fly.io/$(APP_NAME):latest
	$(eval GIT_SHA=$(shell git rev-parse --short HEAD))
	@docker tag bifrost:$(GIT_SHA) registry.fly.io/$(APP_NAME):$(GIT_SHA)
	@echo "$(GREEN)✓ Tagged: registry.fly.io/$(APP_NAME):latest$(NC)"
	@echo "$(GREEN)✓ Tagged: registry.fly.io/$(APP_NAME):$(GIT_SHA)$(NC)"
	@echo ""
	@echo "$(YELLOW)Pushing to Fly.io registry...$(NC)"
	@echo "$(YELLOW)Authenticating with Fly.io...$(NC)"
	@flyctl auth docker
	@echo "$(GREEN)✓ Authenticated with Fly.io$(NC)"
	@echo ""
	@echo "$(YELLOW)Pushing image to Fly.io registry...$(NC)"
	@docker push registry.fly.io/$(APP_NAME):latest
	@docker push registry.fly.io/$(APP_NAME):$(GIT_SHA)
	@echo "$(GREEN)✓ Image pushed to registry$(NC)"
	@echo ""
	@echo "$(YELLOW)Deploying to Fly.io...$(NC)"
	@flyctl deploy -a $(APP_NAME)
	@echo ""
	@echo "$(GREEN)✓ Deployment complete!$(NC)"
	@echo "$(CYAN)App URL: https://$(APP_NAME).fly.dev$(NC)"
	@echo ""
	@# Clean up temporary fly.toml if we created it
	@if [ -f ".fly.toml.tmp.marker" ]; then \
		echo "$(YELLOW)Cleaning up temporary fly.toml...$(NC)"; \
		rm -f fly.toml .fly.toml.tmp.marker; \
		echo "$(GREEN)✓ Temporary fly.toml removed$(NC)"; \
	fi