.PHONY: test
test:
	@go test ./...

.PHONY: install-tools
install-tools: ## Install tools
	@go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.5.0
	@go install golang.org/x/vuln/cmd/govulncheck@latest

check-tool-%:
	@which $* > /dev/null || (echo "Install $* with 'make install-tools'"; exit 1 )

.PHONY: lint
lint: check-tool-golangci-lint
	@golangci-lint run ./...

.PHONY: vulncheck
vulncheck: check-tool-govulncheck
	@govulncheck ./...

BIN_DIR ?= $(shell pwd)/bin

build:
	@mkdir -p $(BIN_DIR)
	@go build -o $(BIN_DIR) ./cmd/chasky 

install:
	@BIN_DIR=$(shell go env GOPATH)/bin $(MAKE) build

generate:
	@go generate ./...
