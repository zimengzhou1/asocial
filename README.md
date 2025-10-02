# asocial

> A real-time collaborative canvas chat application where each keystroke broadcasted to all connected users.

## Overview

asocial is a full-stack real-time collaborative chat application. The backend is built with Go (Gin + Melody for WebSockets), uses Redis for pub/sub messaging, and the frontend is Next.js 15 with Zustand state management. Supports both Docker Compose and Kubernetes (minikube) deployments.

## Architecture

**Docker Compose (Development):**

```
Browser → Traefik → Frontend (Next.js) / Backend (Go) → Redis
```

**Kubernetes (Production-like):**

```
Browser → Ingress-NGINX → Frontend (2 pods) / Backend (3 pods) → Redis (StatefulSet)
```

## Quick Start

### Prerequisites

- **Docker & Docker Compose**: For Docker Compose deployment
- **Minikube**: For Kubernetes deployment
- **Go 1.22+**: For local development
- **Node.js 18+**: For frontend development

### Running with Kubernetes

```bash
# One-time setup: Start minikube and deploy everything
make k8s-setup

# In a separate terminal, expose services to localhost
make k8s-tunnel

# Access the app at http://localhost
```

**Other K8s commands:**

```bash
make k8s-status        # View pod status
make k8s-logs          # Tail logs from all pods
make k8s-clean         # Stop cluster (keeps data)
make k8s-delete        # Delete cluster entirely
```

**How it works:**

- Minikube creates a local Kubernetes cluster
- Backend runs with 3 replicas (load balanced)
- Frontend runs with 2 replicas
- Redis StatefulSet with persistent storage
- Ingress-NGINX routes traffic (`/api/*` → backend, `/*` → frontend)
- `minikube tunnel` exposes Ingress to localhost:80

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

# Start frontend (in another terminal)
cd frontend && npm run dev
```

## API Documentation

### WebSocket Endpoint

#### `GET /api/chat`

Upgrades HTTP connection to WebSocket for real-time bidirectional communication.

**Query Parameters:**

- `user_id` (required): Unique identifier for the user
- `channel_id` (optional): Channel to join (default: "default")

**Example:**

```javascript
const ws = new WebSocket(
  "ws://localhost/api/chat?user_id=user123&channel_id=default"
);

ws.onopen = () => {
  console.log("Connected");
};

ws.onmessage = (event) => {
  const message = JSON.parse(event.data);
  console.log("Received:", message);
};

// Send a message
const message = {
  message_id: crypto.randomUUID(),
  channel_id: "default",
  user_id: "user123",
  payload: "Hello, World!",
  position: { x: 100, y: 200 },
  timestamp: Date.now(),
};
ws.send(JSON.stringify(message));
```

**Message Format:**

```typescript
interface Message {
  type: "chat" | "user_joined" | "user_left" | "user_sync"; // Message type
  message_id?: string; // Unique message identifier (chat only)
  channel_id: string; // Channel name
  user_id: string; // Sender user ID
  payload?: string; // Message content (chat only)
  position?: Position; // Canvas position (chat only)
  users?: string[]; // User list (user_sync only)
  timestamp: number; // Unix timestamp (milliseconds)
}

interface Position {
  x: number; // X coordinate on canvas (float for sub-pixel precision)
  y: number; // Y coordinate on canvas (float for sub-pixel precision)
}
```

**Message Types:**

- `chat`: User-sent chat message with content and position
- `user_joined`: Broadcast when a user connects (presence event)
- `user_left`: Broadcast when a user disconnects (presence event)
- `user_sync`: Sent to new users with current channel user list

### Health Check Endpoints

#### `GET /health` : **Liveness probe** - checks if the application is running.

#### `GET /ready` : **Readiness probe** - checks if the application can accept traffic (Redis connection healthy).

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

**Kubernetes:**

```bash
make k8s-setup            # Setup and deploy to minikube
make k8s-tunnel           # Expose services to localhost
make k8s-status           # Show pod/service status
make k8s-logs             # Tail logs from all pods
make k8s-clean            # Stop cluster (preserves data)
make k8s-delete           # Delete cluster entirely
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

For detailed architecture information, see:

- [ARCHITECTURE.md](docs/ARCHITECTURE.md) - System design and component interactions
- [OLD_ARCHITECTURE.md](docs/OLD_ARCHITECTURE.md) - Legacy system documentation
