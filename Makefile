BINARY_NAME=kong2eg
BINARY_PATH=./bin/$(BINARY_NAME)
SOURCE_PATH=./cmd/kong2eg

.PHONY: build build-all build-linux build-darwin build-windows compress clean test help

build: ## Build the binary for current platform
	@mkdir -p bin
	go build -o $(BINARY_PATH) $(SOURCE_PATH)

build-all: build-linux build-darwin build-windows ## Build for all platforms

build-linux: ## Build for Linux (amd64 and arm64)
	@mkdir -p bin
	GOOS=linux GOARCH=amd64 go build -o bin/$(BINARY_NAME)-linux-amd64 $(SOURCE_PATH)
	GOOS=linux GOARCH=arm64 go build -o bin/$(BINARY_NAME)-linux-arm64 $(SOURCE_PATH)

build-darwin: ## Build for macOS (amd64 and arm64)
	@mkdir -p bin
	GOOS=darwin GOARCH=amd64 go build -o bin/$(BINARY_NAME)-darwin-amd64 $(SOURCE_PATH)
	GOOS=darwin GOARCH=arm64 go build -o bin/$(BINARY_NAME)-darwin-arm64 $(SOURCE_PATH)

build-windows: ## Build for Windows (amd64 and arm64)
	@mkdir -p bin
	GOOS=windows GOARCH=amd64 go build -o bin/$(BINARY_NAME)-windows-amd64.exe $(SOURCE_PATH)
	GOOS=windows GOARCH=arm64 go build -o bin/$(BINARY_NAME)-windows-arm64.exe $(SOURCE_PATH)

compress: ## Compress binary files with gzip
	@echo "Compressing binary files..."
	@for file in bin/$(BINARY_NAME)-*; do \
		if [ -f "$$file" ]; then \
			echo "Compressing $$file..."; \
			gzip -9 "$$file"; \
		fi \
	done
	@echo "Compression complete!"

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
