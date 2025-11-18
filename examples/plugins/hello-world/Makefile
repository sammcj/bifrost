.PHONY: all build clean install help test

# Note: Go plugins only support Linux and macOS (Darwin)
# - Native builds: Use native Go compiler
# - Linux cross-platform: Use Docker with appropriate platform
# - macOS cross-arch: Use GOOS/GOARCH (amd64 <-> arm64 works on macOS)
# - macOS builds require macOS host

# Colors
COLOR_RESET   = \033[0m
COLOR_INFO    = \033[36m
COLOR_SUCCESS = \033[32m
COLOR_WARNING = \033[33m
COLOR_ERROR   = \033[31m
COLOR_BOLD    = \033[1m

# Plugin name
PLUGIN_NAME = hello-world
OUTPUT_DIR = build

# Platform detection
UNAME_S := $(shell uname -s)
ifeq ($(UNAME_S),Linux)
	PLUGIN_EXT = .so
	PLATFORM = linux
	HOST_OS = linux
endif
ifeq ($(UNAME_S),Darwin)
	PLUGIN_EXT = .so
	PLATFORM = darwin
	HOST_OS = darwin
endif

# Architecture detection
UNAME_M := $(shell uname -m)
ifeq ($(UNAME_M),x86_64)
	ARCH = amd64
	HOST_ARCH = amd64
endif
ifeq ($(UNAME_M),arm64)
	ARCH = arm64
	HOST_ARCH = arm64
endif
ifeq ($(UNAME_M),aarch64)
	ARCH = arm64
	HOST_ARCH = arm64
endif

# Determine if we need Docker for cross-compilation
USE_DOCKER = no
ifneq ($(HOST_OS),linux)
	USE_DOCKER = yes
endif

# Output file
OUTPUT = $(OUTPUT_DIR)/$(PLUGIN_NAME)$(PLUGIN_EXT)

help: ## Show this help message
	@echo '$(COLOR_BOLD)Usage:$(COLOR_RESET) make [target]'
	@echo ''
	@echo '$(COLOR_BOLD)Host System:$(COLOR_RESET)'
	@echo '  OS:           $(HOST_OS)'
	@echo '  Architecture: $(HOST_ARCH)'
	@echo ''
	@echo '$(COLOR_BOLD)Available targets:$(COLOR_RESET)'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  $(COLOR_INFO)%-20s$(COLOR_RESET) %s\n", $$1, $$2}' $(MAKEFILE_LIST)

_clean_build_dir:
	@echo "$(COLOR_INFO)Cleaning build directory...$(COLOR_RESET)"
	@rm -rf $(OUTPUT_DIR)
	@echo "$(COLOR_SUCCESS)✓ Build directory cleaned$(COLOR_RESET)"

build: _clean_build_dir ## Build the plugin for current platform
	@echo "$(COLOR_INFO)Building plugin for $(PLATFORM)/$(ARCH)...$(COLOR_RESET)"
	@mkdir -p $(OUTPUT_DIR)
	go build --buildmode=plugin -o $(OUTPUT) main.go
	@echo "$(COLOR_SUCCESS)✓ Plugin built successfully: $(OUTPUT)$(COLOR_RESET)"

build-linux-amd64: _clean_build_dir ## Build the plugin for Linux AMD64
ifeq ($(HOST_OS),linux)
ifeq ($(HOST_ARCH),amd64)
	@echo "$(COLOR_INFO)Building plugin for linux/amd64 (native)...$(COLOR_RESET)"
	@mkdir -p $(OUTPUT_DIR)
	go build --buildmode=plugin -o $(OUTPUT_DIR)/$(PLUGIN_NAME)-linux-amd64.so main.go
	@echo "$(COLOR_SUCCESS)✓ Plugin built successfully: $(OUTPUT_DIR)/$(PLUGIN_NAME)-linux-amd64.so$(COLOR_RESET)"
else
	@echo "$(COLOR_INFO)Building plugin for linux/amd64 using Docker...$(COLOR_RESET)"
	@mkdir -p $(OUTPUT_DIR)
	docker run --rm --platform linux/amd64 -v "$(PWD):/work" -w /work \
		-e GOOS=linux -e GOARCH=amd64 -e CGO_ENABLED=0 \
		golang:1.24.3-alpine \
		go build --buildmode=plugin -o $(OUTPUT_DIR)/$(PLUGIN_NAME)-linux-amd64.so main.go
	@echo "$(COLOR_SUCCESS)✓ Plugin built successfully: $(OUTPUT_DIR)/$(PLUGIN_NAME)-linux-amd64.so$(COLOR_RESET)"
endif
else
	@echo "$(COLOR_INFO)Building plugin for linux/amd64 using Docker...$(COLOR_RESET)"
	@mkdir -p $(OUTPUT_DIR)
	docker run --rm --platform linux/amd64 -v "$(PWD):/work" -w /work \
		-e GOOS=linux -e GOARCH=amd64 -e CGO_ENABLED=0 \
		golang:1.24.3-alpine \
		go build --buildmode=plugin -o $(OUTPUT_DIR)/$(PLUGIN_NAME)-linux-amd64.so main.go
	@echo "$(COLOR_SUCCESS)✓ Plugin built successfully: $(OUTPUT_DIR)/$(PLUGIN_NAME)-linux-amd64.so$(COLOR_RESET)"
endif

build-linux-arm64: _clean_build_dir ## Build the plugin for Linux ARM64
ifeq ($(HOST_OS),linux)
ifeq ($(HOST_ARCH),arm64)
	@echo "$(COLOR_INFO)Building plugin for linux/arm64 (native)...$(COLOR_RESET)"
	@mkdir -p $(OUTPUT_DIR)
	go build -buildmode=plugin -o $(OUTPUT_DIR)/$(PLUGIN_NAME)-linux-arm64.so main.go
	@echo "$(COLOR_SUCCESS)✓ Plugin built successfully: $(OUTPUT_DIR)/$(PLUGIN_NAME)-linux-arm64.so$(COLOR_RESET)"
else
	@echo "$(COLOR_INFO)Building plugin for linux/arm64 using Docker...$(COLOR_RESET)"
	@mkdir -p $(OUTPUT_DIR)
	docker run --rm --platform linux/arm64 -v "$(PWD):/work" -w /work \
		-e GOOS=linux -e GOARCH=arm64 -e CGO_ENABLED=0 \
		golang:1.24.3-alpine \
		go build --buildmode=plugin -o $(OUTPUT_DIR)/$(PLUGIN_NAME)-linux-arm64.so main.go
	@echo "$(COLOR_SUCCESS)✓ Plugin built successfully: $(OUTPUT_DIR)/$(PLUGIN_NAME)-linux-arm64.so$(COLOR_RESET)"
endif
else
	@echo "$(COLOR_INFO)Building plugin for linux/arm64 using Docker...$(COLOR_RESET)"
	@mkdir -p $(OUTPUT_DIR)
	docker run --rm --platform linux/arm64 -v "$(PWD):/work" -w /work \
		-e GOOS=linux -e GOARCH=arm64 -e CGO_ENABLED=0 \
		golang:1.24.3-alpine \
		go build --buildmode=plugin -o $(OUTPUT_DIR)/$(PLUGIN_NAME)-linux-arm64.so main.go
	@echo "$(COLOR_SUCCESS)✓ Plugin built successfully: $(OUTPUT_DIR)/$(PLUGIN_NAME)-linux-arm64.so$(COLOR_RESET)"
endif

build-darwin-amd64: _clean_build_dir ## Build the plugin for macOS AMD64
ifeq ($(HOST_OS),darwin)
ifeq ($(HOST_ARCH),amd64)
	@echo "$(COLOR_INFO)Building plugin for darwin/amd64 (native)...$(COLOR_RESET)"
	@mkdir -p $(OUTPUT_DIR)
	go build --buildmode=plugin -o $(OUTPUT_DIR)/$(PLUGIN_NAME)-darwin-amd64.so main.go
	@echo "$(COLOR_SUCCESS)✓ Plugin built successfully: $(OUTPUT_DIR)/$(PLUGIN_NAME)-darwin-amd64.so$(COLOR_RESET)"
else
	@echo "$(COLOR_INFO)Building plugin for darwin/amd64 (cross-compile)...$(COLOR_RESET)"
	@mkdir -p $(OUTPUT_DIR)
	GOOS=darwin GOARCH=amd64 go build --buildmode=plugin -o $(OUTPUT_DIR)/$(PLUGIN_NAME)-darwin-amd64.so main.go
	@echo "$(COLOR_SUCCESS)✓ Plugin built successfully: $(OUTPUT_DIR)/$(PLUGIN_NAME)-darwin-amd64.so$(COLOR_RESET)"
endif
else
	@echo "$(COLOR_WARNING)⚠ Cannot build macOS plugins on non-macOS systems$(COLOR_RESET)"
	@exit 1
endif

build-darwin-arm64: _clean_build_dir ## Build the plugin for macOS ARM64
ifeq ($(HOST_OS),darwin)
ifeq ($(HOST_ARCH),arm64)
	@echo "$(COLOR_INFO)Building plugin for darwin/arm64 (native)...$(COLOR_RESET)"
	@mkdir -p $(OUTPUT_DIR)
	go build --buildmode=plugin -o $(OUTPUT_DIR)/$(PLUGIN_NAME)-darwin-arm64.so main.go
	@echo "$(COLOR_SUCCESS)✓ Plugin built successfully: $(OUTPUT_DIR)/$(PLUGIN_NAME)-darwin-arm64.so$(COLOR_RESET)"
else
	@echo "$(COLOR_INFO)Building plugin for darwin/arm64 (cross-compile)...$(COLOR_RESET)"
	@mkdir -p $(OUTPUT_DIR)
	GOOS=darwin GOARCH=arm64 go build --buildmode=plugin -o $(OUTPUT_DIR)/$(PLUGIN_NAME)-darwin-arm64.so main.go
	@echo "$(COLOR_SUCCESS)✓ Plugin built successfully: $(OUTPUT_DIR)/$(PLUGIN_NAME)-darwin-arm64.so$(COLOR_RESET)"
endif
else
	@echo "$(COLOR_WARNING)⚠ Cannot build macOS plugins on non-macOS systems$(COLOR_RESET)"
	@exit 1
endif

build-all: _clean_build_dir ## Build for all supported platforms (based on host OS)
ifeq ($(HOST_OS),darwin)
	@$(MAKE) build-linux-amd64 build-linux-arm64 build-darwin-amd64 build-darwin-arm64
else ifeq ($(HOST_OS),linux)
	@$(MAKE) build-linux-amd64 build-linux-arm64
	@echo "$(COLOR_WARNING)⚠ Skipping macOS builds (requires macOS host)$(COLOR_RESET)"
else
	@echo "$(COLOR_ERROR)✗ Unsupported host OS: $(HOST_OS)$(COLOR_RESET)"
	@exit 1
endif

# Alternative: Cross-compile with CGO (EXPERIMENTAL - may not work reliably)
# Requires: brew install FiloSottile/musl-cross/musl-cross (for macOS)
build-linux-amd64-cgo: _clean_build_dir ## [EXPERIMENTAL] Build for Linux AMD64 using cross-compilation with CGO
	@echo "$(COLOR_WARNING)⚠ Building plugin for linux/amd64 with CGO (experimental)...$(COLOR_RESET)"
	@mkdir -p $(OUTPUT_DIR)
	GOOS=linux GOARCH=amd64 CC=x86_64-linux-musl-gcc \
		go build -buildmode=plugin -o $(OUTPUT_DIR)/$(PLUGIN_NAME)-linux-amd64.so main.go
	@echo "$(COLOR_SUCCESS)✓ Plugin built successfully: $(OUTPUT_DIR)/$(PLUGIN_NAME)-linux-amd64.so$(COLOR_RESET)"

clean: _clean_build_dir ## Remove build artifacts
	@echo "$(COLOR_INFO)Cleaning build artifacts...$(COLOR_RESET)"
	@rm -rf $(OUTPUT_DIR)
	@echo "$(COLOR_SUCCESS)✓ Clean complete$(COLOR_RESET)"

install: build ## Build and install the plugin to Bifrost plugins directory
	@echo "$(COLOR_INFO)Installing plugin...$(COLOR_RESET)"
	@mkdir -p ~/.bifrost/plugins
	@cp $(OUTPUT) ~/.bifrost/plugins/
	@echo "$(COLOR_SUCCESS)✓ Plugin installed to ~/.bifrost/plugins/$(PLUGIN_NAME)$(PLUGIN_EXT)$(COLOR_RESET)"

test: _clean_build_dir ## Run tests
	@echo "$(COLOR_INFO)Running tests...$(COLOR_RESET)"
	go test -v ./...

deps: ## Download dependencies
	@echo "$(COLOR_INFO)Downloading dependencies...$(COLOR_RESET)"
	go mod download
	go mod tidy
	@echo "$(COLOR_SUCCESS)✓ Dependencies updated$(COLOR_RESET)"

.DEFAULT_GOAL := help
