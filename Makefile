BINARY_NAME=kong2eg
BINARY_PATH=./bin/$(BINARY_NAME)
SOURCE_PATH=./cmd/kong2eg

.PHONY: build clean test run-example help

build: ## Build the binary
	@mkdir -p bin
	go build -o $(BINARY_PATH) $(SOURCE_PATH)

clean: ## Clean build artifacts
	rm -rf bin/

test: ## Run tests
	go test ./...

deps: ## Download dependencies
	go mod tidy
	go mod download

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

.DEFAULT_GOAL := help
