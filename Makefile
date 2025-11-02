.PHONY: build test clean docker-build docker-run run install lint

APP_NAME=tg-forward
DOCKER_IMAGE=tg-forward:latest
GO_FILES=$(shell find . -name '*.go' -type f -not -path "./vendor/*")

build:
	@echo "Building $(APP_NAME)..."
	@go build -o $(APP_NAME) cmd/tg-forward/main.go

build-linux:
	@echo "Building $(APP_NAME) for Linux..."
	@GOOS=linux GOARCH=amd64 go build -o $(APP_NAME)-linux cmd/tg-forward/main.go

install:
	@echo "Installing dependencies..."
	@go mod download
	@go mod tidy

test:
	@echo "Running tests..."
	@go test -v ./...

test-coverage:
	@echo "Running tests with coverage..."
	@go test -cover ./...
	@go test -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html

lint:
	@echo "Running linter..."
	@go fmt ./...
	@go vet ./...

clean:
	@echo "Cleaning..."
	@rm -f $(APP_NAME) $(APP_NAME)-linux
	@rm -f coverage.out coverage.html
	@go clean

run:
	@echo "Running $(APP_NAME)..."
	@go run cmd/tg-forward/main.go

docker-build:
	@echo "Building Docker image..."
	@docker build -t $(DOCKER_IMAGE) .

docker-run:
	@echo "Running Docker container..."
	@docker run -d \
		--name $(APP_NAME) \
		-p 8080:8080 \
		-v $(PWD)/configs:/app/configs \
		$(DOCKER_IMAGE)

docker-stop:
	@echo "Stopping Docker container..."
	@docker stop $(APP_NAME) || true
	@docker rm $(APP_NAME) || true

docker-logs:
	@docker logs -f $(APP_NAME)

setup:
	@echo "Setting up project..."
	@cp configs/config.example.yaml configs/config.yaml 2>/dev/null || true
	@cp configs/rules.example.json configs/rules.json 2>/dev/null || true
	@echo "Please edit configs/config.yaml with your credentials"

all: clean install test build

help:
	@echo "Available targets:"
	@echo "  build           - Build the application"
	@echo "  build-linux     - Build for Linux"
	@echo "  install         - Install dependencies"
	@echo "  test            - Run tests"
	@echo "  test-coverage   - Run tests with coverage report"
	@echo "  lint            - Run linters"
	@echo "  clean           - Clean build artifacts"
	@echo "  run             - Run the application"
	@echo "  docker-build    - Build Docker image"
	@echo "  docker-run      - Run Docker container"
	@echo "  docker-stop     - Stop Docker container"
	@echo "  docker-logs     - Show Docker logs"
	@echo "  setup           - Setup config files from examples"
	@echo "  all             - Clean, install, test, and build"
