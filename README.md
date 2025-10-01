# Asocial

> A real-time collaborative canvas chat application where users can type anywhere on the canvas, with each keystroke broadcasted instantly to all connected users.

## Overview

Asocial is a full-stack web application demonstrating real-time communication using WebSockets, user presence tracking, clean architecture patterns, and modern cloud-native design principles. Built with Go and Next.js, it showcases production-ready development practices including Docker containerization, Redis-based presence tracking with TTL, Zustand state management, structured logging, health checks, and comprehensive testing.

## Technology Stack

### Backend

- **Language**: Go 1.22.1
- **Framework**: Gin (HTTP routing)
- **WebSocket**: Melody
- **Pub/Sub**: Redis
- **Configuration**: Viper
- **Logging**: slog (structured JSON logging)

### Frontend

- **Framework**: Next.js 14 (App Router)
- **Language**: TypeScript
- **State Management**: Zustand
- **Styling**: TailwindCSS
- **Gestures**: @use-gesture/react
- **Real-time**: Native WebSocket API with custom hook

### Infrastructure

- **Containerization**: Docker & Docker Compose
- **Reverse Proxy**: Traefik
- **Cache/Pub-Sub/Presence**: Redis 7 (Pub/Sub + Sets with TTL)
- **Orchestration** (planned): Kubernetes

## Architecture

```
┌─────────────┐       ┌─────────────┐       ┌─────────────┐
│   Frontend  │◄─────►│   Traefik   │◄─────►│   Backend   │
│  (Next.js)  │       │  (Routing)  │       │     (Go)    │
└─────────────┘       └─────────────┘       └──────┬──────┘
                                                   │
                                                   ▼
                                            ┌─────────────┐
                                            │    Redis    │
                                            │  (Pub/Sub)  │
                                            └─────────────┘
```

### Clean Architecture

```
internal/
├── domain/       # Business entities (Message, Position)
├── service/      # Business logic (MessageService)
├── pubsub/       # Pub/Sub abstraction (Redis implementation)
├── handler/      # HTTP/WebSocket handlers
└── config/       # Configuration management
```

## Quick Start

### Prerequisites

- **Docker & Docker Compose**: For running the application
- **Go 1.22+**: For local development
- **Node.js 18+**: For frontend development

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

### Running Locally

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

### Configuration File

Edit `config.yaml`:

```yaml
server:
  port: "3001"
  max_connections: 200
  max_message_size: 4096

redis:
  addr: "redis:6379"
  password: ""
  db: 0
  channel: "chat:messages"
```

## Development

### Available Make Targets

```bash
make help                 # Show all available targets
make build                # Build the backend binary
make test                 # Run all tests
make test-unit            # Run unit tests only
make test-integration     # Run integration tests (requires Redis)
make test-coverage        # Generate test coverage report
make lint                 # Run linter
make fmt                  # Format code
make clean                # Clean build artifacts
make docker-build         # Build Docker images
make docker-up            # Start all services
make docker-down          # Stop all services
make docker-logs          # View logs
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

### Project Structure

```
asocial/
├── cmd/server/           # Application entry point
├── internal/
│   ├── domain/           # Business entities (Message, Position, MessageType)
│   ├── service/          # Business logic (MessageService)
│   ├── pubsub/           # Redis pub/sub + presence tracking
│   ├── handler/          # HTTP/WebSocket handlers + presence lifecycle
│   └── config/           # Configuration loader
├── frontend/             # Next.js application
│   ├── src/
│   │   ├── app/          # Next.js 14 app router pages
│   │   ├── components/   # React components (Canvas, Messages, Layout)
│   │   ├── hooks/        # Custom hooks (useWebSocket)
│   │   └── stores/       # Zustand state management (chatStore)
│   ├── public/           # Static assets
│   └── package.json
├── tests/
│   └── integration/      # Integration tests
├── docs/                 # Documentation
├── config.yaml           # Application configuration
├── docker-compose.yml    # Docker Compose setup
├── Makefile              # Build automation
└── README.md            # This file
```

## Features

- ✅ Real-time bidirectional communication via WebSockets
- ✅ **User presence tracking with TTL-based cleanup**
- ✅ **Real-time user join/leave notifications**
- ✅ **Frontend state management with Zustand**
- ✅ Multi-channel support
- ✅ Scalable pub/sub architecture with Redis
- ✅ Clean architecture with separation of concerns
- ✅ Structured JSON logging
- ✅ Health check endpoints for Kubernetes
- ✅ Graceful shutdown handling
- ✅ Docker containerization
- ✅ Sub-pixel precision canvas positioning
- ✅ Pan/zoom/pinch gestures
- ✅ Type-safe TypeScript frontend

## Roadmap

- [ ] **Phase 2**: CI/CD pipeline with GitHub Actions
- [ ] **Phase 3**: Kubernetes deployment manifests
- [ ] **Phase 4**: Horizontal Pod Autoscaling
- [ ] **Phase 5**: Observability (Prometheus + Grafana)
- [ ] **Phase 6**: Authentication (Firebase/JWT)
- [ ] **Phase 7**: Rate limiting
- [ ] **Phase 8**: Message persistence

See [docs/PLANNING.md](docs/PLANNING.md) for detailed roadmap.

## Architecture Documentation

For detailed architecture information, see:

- [ARCHITECTURE.md](docs/ARCHITECTURE.md) - System design and component interactions
- [OLD_ARCHITECTURE.md](docs/OLD_ARCHITECTURE.md) - Legacy system documentation
- [PLANNING.md](docs/PLANNING.md) - Implementation phases and progress
