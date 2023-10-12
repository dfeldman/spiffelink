# TODO: This is a very very simple Makefile with minimal functionality. 
# It would be nice to have Go version pinning like SPIRE. Also gosec and govet checks. 
# With well-defined versions of those packages of course. 

# Variables
BINARY_NAME = spiffelink
DOCKER_IMAGE_NAME = spiffelink:latest-local

.PHONY: all test install-linters lint gosec govet build docker-build integration-test

# Default target executed when no arguments are given to make.
all: install-grpc install-linters test lint gosec govet build

# Install grpc
install-grpc:
	@echo "Installing grpc..."
	go get google.golang.org/grpc

# Install linters: gosec and golangci-lint
install-linters: install-gosec install-golangci-lint

# Install gosec
install-gosec:
	@echo "Installing gosec..."
	go get github.com/securego/gosec/v2/cmd/gosec

# Install golangci-lint
install-golangci-lint:
	@echo "Installing golangci-lint..."
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.42.1

# Run Go tests
test:
	@echo "Running Go tests..."
	go test -v ./...

# Lint the code using golangci-lint
lint: install-linters
	@echo "Running linters..."
	golangci-lint run

# Run gosec security scanner
gosec:
	@echo "Running gosec..."
	gosec -quiet ./...

# Run go vet
govet:
	@echo "Running go vet..."
	go vet ./...

# Build the Go binary
build:
	@echo "Building Go binary..."
	go build -o $(BINARY_NAME) ./...

# Build the Docker image
docker-build:
	@echo "Building Docker image..."
	docker build -t $(DOCKER_IMAGE_NAME) .

# Run integration tests
integration-test: docker-build
	@echo "Running integration tests..."
	./integ/integtest.sh

