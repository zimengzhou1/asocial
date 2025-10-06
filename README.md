# asocial

> A real-time collaborative canvas chat application where each keystroke is broadcasted to all connected users.

## Overview

asocial is a full-stack real-time collaborative chat application. The backend is built with Go (Gin + Melody for WebSockets), uses Redis for pub/sub messaging, and the frontend is Next.js 15 with Zustand state management. Supports Docker Compose, Kubernetes (minikube), and k3s (production) deployments.

Available on: [https://asocial.page](https://asocial.page), works best on desktop browsers.

## Quick Start

### Prerequisites

- **Docker & Docker Compose**: For Docker Compose deployment
- **Minikube**: For Kubernetes deployment
- **Go 1.22+**: For local development
- **Node.js 18+**: For frontend development

### Running with Kubernetes (Local - Minikube)

```bash
# One-time setup: Start minikube and deploy everything
make k8s-setup

# In a separate terminal, expose services to localhost
make k8s-tunnel

# Access the app at http://localhost
```

### Running with k3s (Production - Remote Server)

Deploy to a remote server and expose via Cloudflare Tunnel (free, automatic HTTPS).

**Prerequisites:**

- Remote server with SSH access
- Cloudflare account + domain on Cloudflare DNS
- (Optional) Tailscale for easier server access

**Setup:**

```bash
ssh your-server
sudo bash scripts/remote-setup.sh  # Installs k3s + Docker + NGINX Ingress

# Configure kubectl on your local machine
scp your-server:/etc/rancher/k3s/k3s.yaml ~/.kube/k3s-config
# Edit k3s-config: replace 127.0.0.1 with server IP or Tailscale IP
export KUBECONFIG=~/.kube/k3s-config

# Deploy application
make remote-deploy

# Setup Cloudflare Tunnel (on server)
ssh your-server
sudo bash scripts/remote-tunnel-setup.sh
```

### Running with Docker Compose

```bash
# Start all services
make docker-up

# View logs
make docker-logs

# Stop services
make docker-down
```

**Access the application:**

- Frontend: http://localhost
- Backend API: http://localhost/api/chat
- Health Check: http://localhost/health
- Traefik Dashboard: http://localhost:8080

### Running Locally (No containers)

```bash
# Install dependencies
go mod download
cd frontend && npm install && cd ..

# Start Redis (required)
docker run -d -p 6379:6379 redis:7-alpine

# Start backend
make run

# Start frontend
cd frontend && npm run dev
```

## Configuration

### Environment Variables

| Variable                          | Description               | Default         |
| --------------------------------- | ------------------------- | --------------- |
| `ASOCIAL_SERVER_PORT`             | Backend HTTP port         | `3001`          |
| `ASOCIAL_SERVER_MAX_CONNECTIONS`  | Max WebSocket connections | `200`           |
| `ASOCIAL_SERVER_MAX_MESSAGE_SIZE` | Max message size (bytes)  | `4096`          |
| `REDIS_ADDR`                      | Redis address             | `redis:6379`    |
| `ASOCIAL_REDIS_PASSWORD`          | Redis password            | `""`            |
| `ASOCIAL_REDIS_DB`                | Redis database number     | `0`             |
| `ASOCIAL_REDIS_CHANNEL`           | Redis pub/sub channel     | `chat:messages` |

## Development

### Available Make Targets

**Kubernetes (Local - Minikube):**

```bash
make k8s-setup            # Setup and deploy to minikube
make k8s-tunnel           # Expose services to localhost
make k8s-status           # Show pod/service status
make k8s-logs             # Tail logs from all pods
make k8s-clean            # Stop cluster (preserves data)
make k8s-delete           # Delete cluster entirely
```

**Remote k3s (Production):**

```bash
make remote-deploy        # Deploy to remote k3s cluster
make remote-status        # Show remote pod status
make remote-logs          # Tail logs from remote pods
make remote-update        # Update to latest Docker images
```

**Docker Compose:**

```bash
make docker-up            # Start all services
make docker-down          # Stop all services
make docker-logs          # View logs
make docker-build         # Build Docker images
```

**Development:**

```bash
make build                # Build backend binary
make run                  # Run backend locally
make test                 # Run all tests
make test-unit            # Run unit tests only
make test-integration     # Run integration tests (requires Redis)
make test-coverage        # Generate test coverage report
make lint                 # Run linter
make fmt                  # Format code
make clean                # Clean build artifacts
```

### Running Tests

```bash
# Unit tests only (fast, no dependencies)
make test-unit

# Integration tests (requires Redis on localhost:6379)
make test-integration

# All tests with coverage report
make test-coverage
```

## Architecture Documentation

**Docker Compose (Local Development):**

```
Browser → Traefik → Frontend / Backend → Redis
```

**Kubernetes - Minikube (Local Testing):**

```
Browser → Ingress-NGINX → Frontend (2 pods) / Backend (3 pods) → Redis (StatefulSet)
```

**k3s + Cloudflare Tunnel (Production):**

```
Internet → Cloudflare CDN → Tunnel → k3s Ingress-NGINX → Frontend/Backend → Redis
```

For detailed architecture information, see:

- [ARCHITECTURE.md](docs/ARCHITECTURE.md) - System design and component interactions
- [OLD_ARCHITECTURE.md](docs/OLD_ARCHITECTURE.md) - Legacy system documentation
