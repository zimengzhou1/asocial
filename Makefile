.PHONY: help build test test-unit test-integration test-coverage clean docker-build docker-up docker-down docker-logs run dev lint fmt vet tidy k8s-setup k8s-deploy k8s-logs k8s-status k8s-tunnel k8s-clean k8s-delete remote-deploy remote-status remote-logs remote-update

# Default target
.DEFAULT_GOAL := help

## help: Show this help message
help:
	@echo "Available targets:"
	@echo ""
	@grep -E '^##' $(MAKEFILE_LIST) | sed 's/^## /  /'

## build: Build the backend binary
build:
	@echo "Building backend..."
	go build -o bin/server ./main.go
	@echo "Build complete: bin/server"

## run: Run the backend locally (auto-starts Redis if needed)
run: redis-local build
	@echo "Running server..."
	REDIS_ADDR=localhost:6379 ./bin/server

## redis-local: Start Redis locally for development
redis-local:
	@echo "Starting Redis on localhost:6379..."
	@if docker ps | grep -q asocial-redis-local; then \
		echo "Redis already running"; \
	else \
		docker run -d --name asocial-redis-local -p 6379:6379 redis:7-alpine && \
		echo "Redis started"; \
	fi

## redis-stop: Stop local Redis
redis-stop:
	@echo "Stopping Redis..."
	@docker stop asocial-redis-local 2>/dev/null || true
	@docker rm asocial-redis-local 2>/dev/null || true
	@echo "Redis stopped"

## test: Run all tests
test:
	@echo "Running all tests..."
	go test -v ./...

## test-unit: Run unit tests only
test-unit:
	@echo "Running unit tests..."
	go test -v -short ./...

## test-integration: Run integration tests (requires Redis)
test-integration:
	@echo "Running integration tests..."
	@echo "Note: Redis must be running on localhost:6379"
	go test -v -count=1 ./tests/integration

## test-coverage: Run tests with coverage report
test-coverage:
	@echo "Running tests with coverage..."
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

## lint: Run linter (requires golangci-lint)
lint:
	@if command -v golangci-lint > /dev/null; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not installed. Install from: https://golangci-lint.run/usage/install/"; \
	fi

## fmt: Format Go code
fmt:
	@echo "Formatting code..."
	go fmt ./...
	@echo "Code formatted"

## vet: Run go vet
vet:
	@echo "Running go vet..."
	go vet ./...
	@echo "Vet complete"

## clean: Remove build artifacts and caches
clean:
	@echo "Cleaning build artifacts..."
	rm -rf bin/
	rm -f coverage.out coverage.html
	go clean -cache -testcache
	@echo "Clean complete"

## docker-build: Build Docker images
docker-build:
	@echo "Building Docker images..."
	docker compose build
	@echo "Docker build complete"

## docker-up: Start all services with Docker Compose
docker-up:
	@echo "Starting services..."
	docker compose up -d
	@echo "Services started"
	@echo ""
	@echo "Access the application:"
	@echo "  Frontend: http://localhost"
	@echo "  Backend:  http://localhost/api/chat"
	@echo "  Health:   http://localhost/health"
	@echo "  Traefik:  http://localhost:8080"
	@echo ""
	@echo "View logs: make docker-logs"

## docker-down: Stop all services
docker-down:
	@echo "Stopping services..."
	docker compose down
	@echo "Services stopped"

## docker-down-volumes: Stop all services and remove volumes
docker-down-volumes:
	@echo "Stopping services and removing volumes..."
	docker compose down -v
	@echo "Services stopped and volumes removed"

## docker-logs: Show Docker Compose logs
docker-logs:
	docker compose logs -f

## docker-restart: Restart all services
docker-restart: docker-down docker-up

## check: Run all checks (fmt, vet, lint, test)
check: fmt vet test
	@echo "All checks passed"

## k8s-setup: Setup and deploy to local Kubernetes (minikube)
k8s-setup:
	@./scripts/k8s-setup.sh

## k8s-deploy: Apply Kubernetes manifests (without full setup)
k8s-deploy:
	@echo "Applying Kubernetes manifests..."
	kubectl apply -f k8s/namespace.yaml
	kubectl apply -f k8s/redis/
	kubectl apply -f k8s/backend/
	kubectl apply -f k8s/frontend/
	kubectl apply -f k8s/ingress.yaml
	@echo "Manifests applied"

## k8s-logs: Tail logs from all pods
k8s-logs:
	@./scripts/k8s-logs.sh

## k8s-status: Show status of all Kubernetes resources
k8s-status:
	@echo "Kubernetes Resources:"
	@echo ""
	@kubectl get pods,svc,ingress -n asocial

## k8s-tunnel: Start minikube tunnel (run in separate terminal)
k8s-tunnel:
	@echo "Starting minikube tunnel..."
	@echo "Keep this running and access app at http://localhost"
	minikube tunnel

## k8s-clean: Delete Kubernetes resources and stop minikube
k8s-clean:
	@./scripts/k8s-teardown.sh

## k8s-delete: Completely delete minikube cluster
k8s-delete:
	@echo "⚠️  This will delete the entire minikube cluster and all data"
	@read -p "Are you sure? [y/N] " -n 1 -r; \
	echo; \
	if [[ $$REPLY =~ ^[Yy]$$ ]]; then \
		minikube delete; \
		echo "✅ Cluster deleted"; \
	else \
		echo "Cancelled"; \
	fi

## remote-deploy: Deploy to remote k3s cluster
remote-deploy:
	@./scripts/remote-deploy.sh

## remote-status: Show status of remote k3s deployment
remote-status:
	@./scripts/remote-status.sh

## remote-logs: Tail logs from remote k3s pods
remote-logs:
	@./scripts/remote-logs.sh

## remote-update: Update remote deployment with latest images
remote-update:
	@./scripts/remote-update.sh