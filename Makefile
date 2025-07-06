# Makefile for go-await project

# Variables
GOLANGCI_VERSION := v1.61.0
COVERAGE_THRESHOLD := 75
MODULE_NAME := github.com/sachinagada/go-await
GOBIN := $(shell go env GOPATH)/bin

# Color output
GREEN  := $(shell tput -Txterm setaf 2)
YELLOW := $(shell tput -Txterm setaf 3)
RED    := $(shell tput -Txterm setaf 1)
CYAN   := $(shell tput -Txterm setaf 6)
RESET  := $(shell tput -Txterm sgr0)

# Packages to exclude from coverage requirements (space-separated patterns)
COVERAGE_EXCLUDE_PATTERNS := /examples/ /cmd/ /tools/ /scripts/ /mocks/ /testdata/

# Default target
.DEFAULT_GOAL := all

# Main targets
.PHONY: all
all: check build ## Run all checks and build

.PHONY: help
help: ## Show this help message
	@echo "${CYAN}go-await - Promise-like async patterns for Go${RESET}"
	@echo ""
	@echo "${YELLOW}Usage:${RESET} make [target]"
	@echo ""
	@echo "${YELLOW}Main targets:${RESET}"
	@awk 'BEGIN {FS = ":.*?## "} /^(all|build|test|check):.*?## / {printf "  ${CYAN}%-20s${RESET} %s\n", $$1, $$2}' $(MAKEFILE_LIST)
	@echo ""
	@echo "${YELLOW}Testing:${RESET}"
	@awk 'BEGIN {FS = ":.*?## "} /^(coverage|coverage-check|bench):.*?## / {printf "  ${CYAN}%-20s${RESET} %s\n", $$1, $$2}' $(MAKEFILE_LIST)
	@echo ""
	@echo "${YELLOW}Code Quality:${RESET}"
	@awk 'BEGIN {FS = ":.*?## "} /^(fmt|fmt-fix|vet|lint|tidy):.*?## / {printf "  ${CYAN}%-20s${RESET} %s\n", $$1, $$2}' $(MAKEFILE_LIST)
	@echo ""
	@echo "${YELLOW}Documentation:${RESET}"
	@awk 'BEGIN {FS = ":.*?## "} /^(doc|doc-pkg|doc-all|doc-serve):.*?## / {printf "  ${CYAN}%-20s${RESET} %s\n", $$1, $$2}' $(MAKEFILE_LIST)
	@echo ""
	@echo "${YELLOW}Development Tools:${RESET}"
	@awk 'BEGIN {FS = ":.*?## "} /^(install-tools|pre-commit-install|pre-commit-run):.*?## / {printf "  ${CYAN}%-20s${RESET} %s\n", $$1, $$2}' $(MAKEFILE_LIST)
	@echo ""
	@echo "${YELLOW}Other:${RESET}"
	@awk 'BEGIN {FS = ":.*?## "} /^(quick|run-example|security|clean):.*?## / {printf "  ${CYAN}%-20s${RESET} %s\n", $$1, $$2}' $(MAKEFILE_LIST)

.PHONY: check
check: fmt vet lint test ## Run all checks (fmt, vet, lint, test)

# Building
.PHONY: build
build: ## Build the project
	@echo "${GREEN}Building...${RESET}"
	@go build -v $$(go list ./... | grep -v /examples/)

# Testing
.PHONY: test
test: ## Run tests with race detector
	@echo "${GREEN}Running tests...${RESET}"
	@go test -v -race $$(go list ./... | grep -v /examples/)

.PHONY: coverage
coverage: ## Run tests with coverage report
	@echo "${GREEN}Running tests with coverage...${RESET}"
	@go test -v -race -coverprofile=coverage.out -covermode=atomic $$(go list ./... | grep -v /examples/)
	@go tool cover -html=coverage.out -o coverage.html
	@echo "${CYAN}Coverage report generated: coverage.html${RESET}"

.PHONY: coverage-check
coverage-check: ## Check if test coverage meets threshold (75%)
	@echo "${GREEN}Checking test coverage (minimum $(COVERAGE_THRESHOLD)%)...${RESET}"
	@echo "${GREEN}Coverage by package:${RESET}"
	@rm -f /tmp/coverage_failed; \
	go test -coverprofile=coverage.out -covermode=atomic $$(go list ./... | grep -v /examples/) 2>&1 | grep "coverage:" | while read -r line; do \
		pkg=$$(echo "$$line" | awk '{print $$2}'); \
		coverage=$$(echo "$$line" | awk '{print $$5}' | sed 's/%//' | cut -d. -f1); \
		excluded=0; \
		for pattern in $(COVERAGE_EXCLUDE_PATTERNS); do \
			if echo "$$pkg" | grep -q "$$pattern"; then \
				excluded=1; \
				break; \
			fi; \
		done; \
		if [ "$$excluded" -eq 1 ]; then \
			printf "  %-60s %s%% %s\n" "$$pkg" "$$coverage" "${YELLOW}(excluded)${RESET}"; \
		elif [ -z "$$coverage" ] || [ "$$coverage" = "statements" ] || [ "$$pkg" = "?" ]; then \
			continue; \
		elif [ "$$coverage" -lt "$(COVERAGE_THRESHOLD)" ]; then \
			printf "  %-60s ${RED}%s%%${RESET} < $(COVERAGE_THRESHOLD)%%\n" "$$pkg" "$$coverage"; \
			touch /tmp/coverage_failed; \
		else \
			printf "  %-60s ${GREEN}%s%%${RESET}\n" "$$pkg" "$$coverage"; \
		fi; \
	done; \
	if [ -f /tmp/coverage_failed ]; then \
		rm -f /tmp/coverage_failed; \
		echo "${RED}Coverage check failed! Some packages are below $(COVERAGE_THRESHOLD)%${RESET}"; \
		exit 1; \
	else \
		echo "${GREEN}All packages meet minimum coverage threshold!${RESET}"; \
	fi

.PHONY: bench
bench: ## Run benchmarks
	@echo "${GREEN}Running benchmarks...${RESET}"
	@go test -bench=. -benchmem ./...

# Code Quality
.PHONY: fmt
fmt: ## Check code formatting
	@echo "${GREEN}Checking formatting...${RESET}"
	@output=$$(gofmt -l .); \
	if [ -n "$$output" ]; then \
		echo "${RED}The following files need formatting:${RESET}"; \
		echo "$$output"; \
		echo "${YELLOW}Run 'make fmt-fix' to fix them${RESET}"; \
		exit 1; \
	else \
		echo "${GREEN}All files are properly formatted${RESET}"; \
	fi

.PHONY: fmt-fix
fmt-fix: ## Fix code formatting
	@echo "${GREEN}Fixing formatting...${RESET}"
	@gofmt -w .
	@goimports -w .

.PHONY: vet
vet: ## Run go vet
	@echo "${GREEN}Running go vet...${RESET}"
	@go vet $$(go list ./... | grep -v /examples/)

.PHONY: lint
lint: ## Run golangci-lint
	@echo "${GREEN}Running golangci-lint...${RESET}"
	@if ! command -v golangci-lint >/dev/null 2>&1; then \
		echo "${RED}golangci-lint not installed. Run 'make install-tools' to install it${RESET}"; \
		exit 1; \
	fi
	@golangci-lint run ./...

.PHONY: tidy
tidy: ## Run go mod tidy and verify
	@echo "${GREEN}Running go mod tidy...${RESET}"
	@go mod tidy -v
	@go mod verify

# Development Tools
.PHONY: install-tools
install-tools: ## Install development tools (golangci-lint, pre-commit, pkgsite)
	@echo "${GREEN}Installing development tools...${RESET}"
	# Install golangci-lint
	@if ! command -v golangci-lint >/dev/null 2>&1; then \
		echo "${CYAN}Installing golangci-lint $(GOLANGCI_VERSION)...${RESET}"; \
		curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(GOBIN) $(GOLANGCI_VERSION); \
	else \
		echo "${GREEN}golangci-lint already installed${RESET}"; \
	fi
	# Install pre-commit
	@if ! command -v pre-commit >/dev/null 2>&1; then \
		echo "${CYAN}Installing pre-commit...${RESET}"; \
		pip install --user pre-commit || pip3 install --user pre-commit || echo "${RED}Failed to install pre-commit. Please install Python and pip first.${RESET}"; \
	else \
		echo "${GREEN}pre-commit already installed${RESET}"; \
	fi
	# Install pkgsite for documentation
	@if ! command -v pkgsite >/dev/null 2>&1; then \
		echo "${CYAN}Installing pkgsite...${RESET}"; \
		go install golang.org/x/pkgsite/cmd/pkgsite@latest; \
	else \
		echo "${GREEN}pkgsite already installed${RESET}"; \
	fi
	# Download dependencies
	@go mod download
	@echo "${GREEN}All tools installed!${RESET}"

.PHONY: quick
quick: fmt-fix test ## Quick check (format and test)

.PHONY: run-example
run-example: ## Run the basic example
	@echo "${GREEN}Running basic example...${RESET}"
	@cd examples/basic && go run main.go

.PHONY: security
security: ## Check for security vulnerabilities
	@echo "${GREEN}Checking for vulnerabilities...${RESET}"
	@if command -v nancy >/dev/null 2>&1; then \
		go list -json -deps ./... | nancy sleuth; \
	else \
		echo "${YELLOW}Nancy not installed. Install with: go install github.com/sonatype-nexus-community/nancy@latest${RESET}"; \
		echo "${CYAN}Running go list to check for outdated dependencies...${RESET}"; \
		go list -u -m all; \
	fi

# Documentation
.PHONY: doc
doc: ## View package documentation in terminal
	@echo "${GREEN}Showing documentation for current package...${RESET}"
	@go doc -all .

.PHONY: doc-pkg
doc-pkg: ## View documentation for specific package (usage: make doc-pkg PKG=retry)
	@if [ -z "$(PKG)" ]; then \
		echo "${RED}Please specify a package: make doc-pkg PKG=retry${RESET}"; \
		exit 1; \
	fi
	@echo "${GREEN}Showing documentation for package $(PKG)...${RESET}"
	@go doc -all ./$(PKG)

.PHONY: doc-all
doc-all: ## View all documentation including unexported symbols
	@echo "${GREEN}Showing all documentation (including unexported)...${RESET}"
	@go doc -all -u .

.PHONY: doc-serve
doc-serve: ## Start documentation web server
	@echo "${GREEN}Starting documentation server...${RESET}"
	@if command -v pkgsite >/dev/null 2>&1; then \
		pkgsite -http=:8080 & \
		echo "${GREEN}Documentation server started at http://localhost:8080${RESET}"; \
		echo "${YELLOW}Browse to http://localhost:8080/$(MODULE_NAME)${RESET}"; \
	elif command -v godoc >/dev/null 2>&1; then \
		godoc -http=:6060 & \
		echo "${GREEN}Documentation server started at http://localhost:6060${RESET}"; \
		echo "${YELLOW}Browse to http://localhost:6060/pkg/$(MODULE_NAME)${RESET}"; \
	else \
		echo "${YELLOW}Neither pkgsite nor godoc found. Install with:${RESET}"; \
		echo "  go install golang.org/x/pkgsite/cmd/pkgsite@latest"; \
		echo "  go install golang.org/x/tools/cmd/godoc@latest"; \
		exit 1; \
	fi

# Pre-commit Integration
.PHONY: pre-commit-install
pre-commit-install: ## Install pre-commit hooks
	@echo "${GREEN}Installing pre-commit hooks...${RESET}"
	@pre-commit install
	@echo "${GREEN}Pre-commit hooks installed${RESET}"

.PHONY: pre-commit-run
pre-commit-run: ## Run pre-commit on all files
	@echo "${GREEN}Running pre-commit on all files...${RESET}"
	@pre-commit run --all-files

# Utilities
.PHONY: clean
clean: ## Remove build artifacts and coverage files
	@echo "${GREEN}Cleaning...${RESET}"
	@rm -f coverage.out coverage.html
	@go clean -cache -testcache -modcache
	@echo "${GREEN}Clean complete${RESET}"
