# Asocial - System Architecture

> A real-time collaborative canvas chat application

---

## Table of Contents

- [System Architecture](#system-architecture)
- [Technology Stack](#technology-stack)
- [Architecture Decisions](#architecture-decisions)
- [Clean Architecture Pattern](#clean-architecture-pattern)
- [Kubernetes-Native Features](#kubernetes-native-features)

---

## System Architecture

### High-Level Overview

```
┌────────────────────────────────────────────────────────────────────────┐
│                          Kubernetes Cluster                            │
│                                                                        │
│  ┌────────────────────────────────────────────────────────────────┐    │
│  │                 Ingress-NGINX Controller                       │    │
│  │  • TLS Termination (cert-manager)                              │    │
│  │  • Path-based routing (/api/*, /*)                             │    │
│  │  • Rate limiting (100 req/min per IP)                          │    │
│  │  • WebSocket upgrade support                                   │    │
│  └──────────────┬──────────────────────────┬──────────────────────┘    │
│                 │                          │                           │
│    ┌────────────▼────────────┐   ┌─────────▼──────────────────────┐    │
│    │   Frontend Service      │   │    Backend Service             │    │
│    │   Type: ClusterIP       │   │    Type: ClusterIP             │    │
│    │   Port: 80              │   │    Port: 8080                  │    │
│    └────────────┬────────────┘   └─────────┬──────────────────────┘    │
│                 │                          │                           │
│    ┌────────────▼────────────┐   ┌─────────▼──────────────────────┐    │
│    │   Frontend Pods         │   │    Backend Pods                │    │
│    │   ┌──────────────────┐  │   │   ┌──────────────────────────┐ │    │
│    │   │ Next.js 14       │  │   │   │ Go (Gin + Melody)        │ │    │
│    │   │ Node.js 20       │  │   │   │ • WebSocket Handler      │ │    │
│    │   │ Port: 3000       │  │   │   │ • HTTP API               │ │    │
│    │   └──────────────────┘  │   │   │ • Health Checks          │ │    │
│    │  Replicas: 2-5          │   │   │ • Prometheus Metrics     │ │    │
│    │  Strategy: RollingUpdate│   │   └──────────────────────────┘ │    │
│    │  Resources:             │   │   Replicas: 3-10 (HPA)         │    │
│    │  - CPU: 100m-500m       │   │   Strategy: RollingUpdate      │    │
│    │  - Memory: 128Mi-512Mi  │   │   Resources:                   │    │
│    └─────────────────────────┘   │   - CPU: 100m-1000m            │    │
│                                  │   - Memory: 64Mi-256Mi         │    │
│                                  └─────────┬──────────────────────┘    │
│                                            │                           │
│                                  ┌─────────▼──────────────────────┐    │
│                                  │    Redis Service               │    │
│                                  │    Type: ClusterIP             │    │
│                                  │    Port: 6379                  │    │
│                                  └─────────┬──────────────────────┘    │
│                                            │                           │
│                                   ┌────────▼───────────────────────┐   │
│                                   │    Redis StatefulSet           │   │
│                                   │   ┌──────────────────────────┐ │   │
│                                   │   │ Redis 7.x                │ │   │
│                                   │   │ • Pub/Sub (messages)     │ │   │
│                                   │   │ • Cache (sessions)       │ │   │
│                                   │   │ • Rate limiting counters │ │   │
│                                   │   └──────────────────────────┘ │   │
│                                   │ Replicas: 3 (High Availability)│   │
│                                   │   Persistence: Enabled (PVC)   │   │
│                                   │   Storage: 5Gi                 │   │
│                                   └────────────────────────────────┘   │
│                                                                        │
│  ┌────────────────────────────────────────────────────────────────┐    │
│  │                   Observability Stack (Optional - Phase 5)     │    │
│  │                                                                │    │
│  │  ┌──────────────┐     ┌──────────────┐     ┌──────────────┐    │    │
│  │  │ Prometheus   │────▶│   Grafana    │     │     Loki     │    │    │
│  │  │              │     │              │◀────│              │    │    │
│  │  │ • Metrics    │     │ • Dashboards │     │ • Logs       │    │    │
│  │  │ • Alerts     │     │ • Queries    │     │ • Search     │    │    │
│  │  │ • TSDB       │     │ • Alerts     │     │ • Aggregation│    │    │
│  │  └──────────────┘     └──────────────┘     └──────────────┘    │    │
│  │                                                                │    │
│  └────────────────────────────────────────────────────────────────┘    │
│                                                                        │
└────────────────────────────────────────────────────────────────────────┘
```

### Message Flow Diagram

```
User A                 Frontend Pod              Backend Pod             Redis            Backend Pod           Frontend Pod              User B
  │                        │                          │                    │                   │                      │                      │
  │  1. Click Canvas       │                          │                    │                   │                      │                      │
  ├───────────────────────▶│                          │                    │                   │                      │                      │
  │                        │                          │                    │                   │                      │                      │
  │  2. Type "Hello"       │                          │                    │                   │                      │                      │
  ├───────────────────────▶│                          │                    │                   │                      │                      │
  │                        │                          │                    │                   │                      │                      │
  │                        │  3. WebSocket msg        │                    │                   │                      │                      │
  │                        ├─────────────────────────▶│                    │                   │                      │                      │
  │                        │  {user_id, msg_id,       │                    │                   │                      │                      │
  │                        │   payload, position}     │                    │                   │                      │                      │
  │                        │                          │                    │                   │                      │                      │
  │                        │                          │  4. PUBLISH        │                   │                      │                      │
  │                        │                          │  channel:default   │                   │                      │                      │
  │                        │                          ├───────────────────▶│                   │                      │                      │
  │                        │                          │                    │                   │                      │                      │
  │                        │                          │                    │  5. SUBSCRIBE     │                      │                      │
  │                        │                          │                    │  channel:default  │                      │                      │
  │                        │                          │                    ├──────────────────▶│                      │                      │
  │                        │                          │                    │                   │                      │                      │
  │                        │                          │                    │                   │  6. Filter (user_id) │                      │
  │                        │                          │                    │                   │  Broadcast to others │                      │
  │                        │                          │                    │                   ├─────────────────────▶│                      │
  │                        │                          │                    │                   │  WebSocket           │                      │
  │                        │                          │                    │                   │                      │                      │
  │                        │                          │                    │                   │                      │  7. Render "Hello"   │
  │                        │                          │                    │                   │                      ├─────────────────────▶│
  │                        │                          │                    │                   │                      │  at position (x,y)   │
  │                        │                          │                    │                   │                      │                      │
  │                        │                          │                    │                   │                      │  8. Fade after 5s    │
  │                        │                          │                    │                   │                      ├─────────────────────▶│
```

### Component Interactions

```
┌─────────────────────────────────────────────────────────────────────────┐
│                         Backend Service (Go)                            │
│                                                                         │
│  ┌──────────────────────────────────────────────────────────────────┐   │
│  │                      HTTP Layer (Gin)                            │   │
│  │  ┌────────────┐  ┌────────────┐  ┌────────────┐  ┌────────────┐  │   │
│  │  │ WebSocket  │  │   Health   │  │  Metrics   │  │   API      │  │   │
│  │  │  Handler   │  │   /health  │  │ /metrics   │  │  Routes    │  │   │
│  │  └──────┬─────┘  └──────┬─────┘  └──────┬─────┘  └──────┬─────┘  │   │
│  └─────────┼───────────────┼───────────────┼───────────────┼────────┘   │
│            │               │               │               │            │
│  ┌─────────▼───────────────▼───────────────▼───────────────▼────────┐   │
│  │                      Middleware Layer                            │   │
│  │  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐          │   │
│  │  │  Auth    │  │ Logging  │  │ Metrics  │  │  CORS    │          │   │
│  │  │  (JWT)   │  │ (slog)   │  │(Promeths)│  │          │          │   │
│  │  └──────────┘  └──────────┘  └──────────┘  └──────────┘          │   │
│  └─────────────────────────────────┬────────────────────────────────┘   │
│                                    │                                    │
│  ┌─────────────────────────────────▼────────────────────────────────┐   │
│  │                        Service Layer                             │   │
│  │  ┌─────────────────────────────────────────────────────────────┐ │   │
│  │  │              MessageService (Business Logic)                │ │   │
│  │  │  • Validate message format                                  │ │   │
│  │  │  • Enforce rate limits (100 msg/min per user)               │ │   │
│  │  │  • Handle message lifecycle                                 │ │   │
│  │  │  • Coordinate pub/sub                                       │ │   │
│  │  └────────────────────────┬────────────────────────────────────┘ │   │
│  │                           │                                      │   │
│  │  ┌────────────────────────▼────────────────────────────────────┐ │   │
│  │  │              BroadcastService                               │ │   │
│  │  │  • Filter recipients (exclude sender)                       │ │   │
│  │  │  • Handle channel-based routing                             │ │   │
│  │  │  • Manage WebSocket connections pool                        │ │   │
│  │  └────────────────────────┬────────────────────────────────────┘ │   │
│  └───────────────────────────┼──────────────────────────────────────┘   │
│                              │                                          │
│  ┌───────────────────────────▼──────────────────────────────────────┐   │
│  │                     Repository Layer                             │   │
│  │  ┌─────────────────────────────────────────────────────────────┐ │   │
│  │  │                   RedisRepository                           │ │   │
│  │  │  • Session storage (user → connection mapping)              │ │   │
│  │  │  • Rate limit counters (INCR with TTL)                      │ │   │
│  │  │  • Message cache (optional, for persistence)                │ │   │
│  │  └────────────────────────┬────────────────────────────────────┘ │   │
│  │                           │                                      │   │
│  │  ┌────────────────────────▼────────────────────────────────────┐ │   │
│  │  │                   RedisPubSub                               │ │   │
│  │  │  • PUBLISH messages to channel                              │ │   │
│  │  │  • SUBSCRIBE to channels                                    │ │   │
│  │  │  • Handle reconnection logic                                │ │   │
│  │  └────────────────────────┬────────────────────────────────────┘ │   │
│  └───────────────────────────┼──────────────────────────────────────┘   │
│                              │                                          │
│                      ┌───────▼────────┐                                 │
│                      │  Redis Client  │                                 │
│                      │  (go-redis/v9) │                                 │
│                      └────────────────┘                                 │
└─────────────────────────────────────────────────────────────────────────┘
```

---

## Technology Stack

### Backend (Go)

| Component                | Technology               | Purpose                                  | Version |
| ------------------------ | ------------------------ | ---------------------------------------- | ------- |
| **HTTP Framework**       | Gin                      | Fast HTTP router, middleware support     | v1.9+   |
| **WebSocket**            | Melody                   | WebSocket library built on Gorilla       | v1.2+   |
| **Pub/Sub**              | go-redis/redis/v9        | Redis client with pub/sub support        | v9.0+   |
| **Config**               | Viper                    | Configuration management                 | v1.18+  |
| **Logging**              | slog                     | Structured logging (Go 1.21+)            | stdlib  |
| **Metrics**              | prometheus/client_golang | Prometheus metrics exporter              | v1.19+  |
| **Dependency Injection** | Wire                     | Compile-time DI (optional, can refactor) | v0.6+   |
| **Testing**              | testify                  | Assertions and mocking                   | v1.9+   |
| **Mocking**              | gomock                   | Mock generation                          | v1.6+   |
| **Validation**           | go-playground/validator  | Struct validation                        | v10.19+ |
| **UUID**                 | google/uuid              | Unique ID generation                     | v1.6+   |

### Frontend (Next.js)

| Component      | Technology           | Purpose                  | Version |
| -------------- | -------------------- | ------------------------ | ------- |
| **Framework**  | Next.js              | React framework with SSR | v14.1+  |
| **Language**   | TypeScript           | Type safety              | v5.0+   |
| **UI Library** | React                | UI components            | v18.0+  |
| **Styling**    | TailwindCSS          | Utility-first CSS        | v3.3+   |
| **Canvas**     | react-zoom-pan-pinch | Pan/zoom functionality   | v3.4+   |
| **Auth**       | NextAuth.js          | Authentication           | v5.0+   |
| **Firebase**   | Firebase SDK         | User authentication      | v10.11+ |
| **WebSocket**  | native WebSocket API | Real-time communication  | -       |

### Infrastructure

| Component             | Technology        | Purpose                          | Environment    |
| --------------------- | ----------------- | -------------------------------- | -------------- |
| **Container Runtime** | Docker            | Containerization                 | All            |
| **Orchestration**     | Kubernetes        | Container orchestration          | All            |
| **Local K8s**         | Minikube or Kind  | Local development cluster        | Dev            |
| **Ingress**           | Ingress-NGINX     | Load balancing, routing          | All            |
| **TLS**               | cert-manager      | Certificate management           | Prod           |
| **Registry**          | Docker Hub / GHCR | Container image registry         | All            |
| **CI/CD**             | GitHub Actions    | Automated testing and deployment | All            |
| **Monitoring**        | Prometheus        | Metrics collection               | Prod (Phase 5) |
| **Visualization**     | Grafana           | Metrics dashboards               | Prod (Phase 5) |
| **Logging**           | Loki              | Log aggregation                  | Prod (Phase 5) |
| **Homelab K8s**       | K3s               | Lightweight K8s for homelab      | Future         |

### Database/Cache

| Component       | Technology    | Purpose                        |
| --------------- | ------------- | ------------------------------ |
| **Pub/Sub**     | Redis Pub/Sub | Real-time message broadcasting |
| **Cache**       | Redis         | Session storage, rate limiting |
| **Persistence** | Redis RDB/AOF | Optional message persistence   |

---

## Architecture Decisions

### 1. Replace Kafka with Redis Pub/Sub

**Problem:** Current architecture uses Apache Kafka + Zookeeper for message brokering.

**Decision:** Replace with Redis Pub/Sub

**Rationale:**

| Factor                   | Kafka + Zookeeper                             | Redis Pub/Sub                         | Winner                 |
| ------------------------ | --------------------------------------------- | ------------------------------------- | ---------------------- |
| **Resource Usage**       | ~1GB memory, 3 containers                     | ~50MB memory, 1 container             | ✅ Redis               |
| **Complexity**           | High (ZooKeeper, partitions, consumer groups) | Low (simple pub/sub)                  | ✅ Redis               |
| **Latency**              | 5-10ms                                        | 1-2ms                                 | ✅ Redis               |
| **Throughput**           | 1M+ msg/sec                                   | 100K+ msg/sec                         | Kafka (but not needed) |
| **Message Persistence**  | Yes (configurable retention)                  | No (ephemeral)                        | Kafka (but not needed) |
| **Operational Overhead** | High (topic management, rebalancing)          | Low (single instance)                 | ✅ Redis               |
| **Learning Curve**       | Steep                                         | Gentle                                | ✅ Redis               |
| **Multi-purpose**        | No                                            | Yes (cache + pub/sub + rate limiting) | ✅ Redis               |
| **K8s Footprint**        | 3 pods (ZooKeeper + 2 Kafka)                  | 1 pod (3 with HA)                     | ✅ Redis               |

**Why It Works:**

- **Scale:** Chat app with 100-1000 concurrent users, not millions
- **Ephemeral messages:** Users don't need message history beyond 5 seconds
- **Simplicity:** Redis eliminates operational complexity
- **Dual purpose:** Same Redis instance handles session caching and rate limiting
- **Homelab friendly:** 70% less resource usage

**When to Use Kafka Instead:**

- Message persistence required (audit logs, analytics)
- Millions of messages per second
- Complex stream processing (aggregations, windowing)
- Guaranteed delivery with replay capability

### 2. Clean Architecture Pattern

**Problem:** Current code mixes concerns (HTTP handlers directly call Kafka, business logic scattered).

**Decision:** Implement Clean Architecture (Hexagonal Architecture)

**Structure:**

```
Domain Layer (Business Entities)
       ↓
Service Layer (Business Logic)
       ↓
Repository Layer (Data Access)
       ↓
Infrastructure Layer (External Systems)
```

**Benefits:**

- **Testability:** Each layer tested independently
- **Maintainability:** Clear separation of concerns
- **Flexibility:** Swap implementations (Redis → PostgreSQL) without changing business logic
- **CV Value:** Shows understanding of software design principles

### 3. Kubernetes-Native Design

**Decision:** Design specifically for Kubernetes deployment from the start

**Principles:**

- **12-Factor App:** Stateless processes, config via environment, port binding
- **Health Checks:** Liveness and readiness probes
- **Graceful Shutdown:** Handle SIGTERM properly (5-30s drain period)
- **Observability:** Structured logs, metrics, tracing
- **Resource Limits:** Define requests/limits for proper scheduling
- **Horizontal Scaling:** Stateless backend pods (no local state)

### 4. Observability First

**Decision:** Build observability into the application from day one

**Approach:**

- **Logging:** Structured JSON logs with trace IDs (slog)
- **Metrics:** Prometheus metrics for all key operations
  - `websocket_connections_active{pod="backend-1"}`
  - `messages_published_total{channel="default"}`
  - `message_broadcast_latency_seconds{quantile="0.99"}`
  - `http_request_duration_seconds{handler="/api/chat", status="200"}`
- **Tracing:** Correlation IDs for request tracking (optional: OpenTelemetry)

**Why Early:**

- Debugging K8s issues is hard without proper observability
- Easier to build in than add later
- Shows production-ready thinking

### 5. Authentication Strategy

**Decision:** Keep Firebase Auth (already implemented)

**Flow:**

1. Frontend authenticates with Firebase
2. Frontend receives JWT token
3. WebSocket connection sends JWT in query param or header
4. Backend validates JWT with Firebase Admin SDK
5. User ID extracted and used for message attribution

**Alternative Considered:** Roll our own auth → Too complex, not the focus of this project

---

## Clean Architecture Pattern

### Directory Structure

```
backend/
├── cmd/
│   └── server/
│       └── main.go                    # Application entry point
│
├── internal/                          # Private application code
│   │
│   ├── domain/                        # Business entities (models)
│   │   ├── message.go                 # Message entity
│   │   ├── user.go                    # User entity
│   │   ├── channel.go                 # Channel entity
│   │   └── errors.go                  # Domain-specific errors
│   │
│   ├── service/                       # Business logic layer
│   │   ├── message_service.go         # Message business logic
│   │   │   • CreateMessage()
│   │   │   • ValidateMessage()
│   │   │   • BroadcastMessage()
│   │   ├── user_service.go            # User management
│   │   └── rate_limiter.go            # Rate limiting logic
│   │
│   ├── repository/                    # Data access layer
│   │   ├── interface.go               # Repository interfaces
│   │   ├── redis_repository.go        # Redis implementation
│   │   │   • SaveSession()
│   │   │   • GetSession()
│   │   │   • IncrementRateLimit()
│   │   └── memory_repository.go       # In-memory (for testing)
│   │
│   ├── pubsub/                        # Pub/Sub abstraction
│   │   ├── interface.go               # PubSub interface
│   │   ├── redis_pubsub.go            # Redis pub/sub implementation
│   │   │   • Publish()
│   │   │   • Subscribe()
│   │   │   • HandleReconnect()
│   │   └── memory_pubsub.go           # In-memory (for testing)
│   │
│   ├── handler/                       # HTTP/WebSocket handlers (thin layer)
│   │   ├── websocket_handler.go       # WebSocket connection handling
│   │   │   • HandleConnect()
│   │   │   • HandleMessage()
│   │   │   • HandleDisconnect()
│   │   ├── health_handler.go          # Health check endpoints
│   │   │   • GET /health (liveness)
│   │   │   • GET /ready (readiness)
│   │   ├── metrics_handler.go         # Prometheus metrics endpoint
│   │   │   • GET /metrics
│   │   └── api_handler.go             # REST API (if needed)
│   │
│   ├── middleware/                    # HTTP middleware
│   │   ├── auth.go                    # JWT validation
│   │   ├── logging.go                 # Request logging
│   │   ├── metrics.go                 # Metrics collection
│   │   ├── cors.go                    # CORS headers
│   │   └── recovery.go                # Panic recovery
│   │
│   └── config/                        # Configuration structs
│       ├── config.go                  # Config struct definitions
│       └── validator.go               # Config validation
│
├── pkg/                               # Public libraries (reusable)
│   ├── logger/
│   │   └── logger.go                  # Structured logging setup
│   ├── validator/
│   │   └── validator.go               # Input validation utilities
│   └── tracing/
│       └── trace.go                   # Correlation ID generation
│
├── tests/
│   ├── unit/                          # Unit tests (per package)
│   │   ├── service_test.go
│   │   ├── repository_test.go
│   │   └── handler_test.go
│   ├── integration/                   # Integration tests
│   │   ├── redis_test.go              # Test with real Redis
│   │   └── websocket_test.go          # Test WebSocket flow
│   └── e2e/                           # End-to-end tests
│       └── chat_test.go               # Full user flow
│
├── scripts/
│   ├── generate-mocks.sh              # Generate gomock mocks
│   └── run-tests.sh                   # Run all tests with coverage
│
├── config/
│   ├── config.yaml                    # Default config
│   ├── config.dev.yaml                # Development overrides
│   └── config.prod.yaml               # Production overrides
│
├── Dockerfile                         # Multi-stage production build
├── Dockerfile.dev                     # Development hot-reload
├── go.mod
├── go.sum
├── Makefile                           # Common tasks
└── README.md
```

### Layer Dependencies

```
┌─────────────────────────────────────────────────────┐
│              main.go (Dependency Injection)         │
│  • Wires all components together                    │
│  • Starts HTTP server                               │
│  • Handles graceful shutdown                        │
└───────────────────────┬─────────────────────────────┘
                        │
        ┌───────────────▼────────────────┐
        │       Handler Layer            │
        │  (HTTP/WebSocket Interface)    │
        │  • Thin layer, no business logic│
        │  • Parse requests, return responses│
        └───────────────┬────────────────┘
                        │
        ┌───────────────▼────────────────┐
        │       Service Layer            │
        │    (Business Logic)            │
        │  • Independent of HTTP/WS      │
        │  • Orchestrates repositories   │
        │  • Enforces business rules     │
        └───────────────┬────────────────┘
                        │
        ┌───────────────▼────────────────┐
        │    Repository Layer            │
        │    (Data Access)               │
        │  • Abstracts storage           │
        │  • Interface-based (testable)  │
        └───────────────┬────────────────┘
                        │
        ┌───────────────▼────────────────┐
        │   Infrastructure Layer         │
        │  (External Systems)            │
        │  • Redis, Kafka, PostgreSQL    │
        │  • Swappable implementations   │
        └────────────────────────────────┘
```

**Key Principle:** Dependencies point inward. Inner layers don't know about outer layers.

---

## Kubernetes-Native Features

### 1. Health Checks

#### Liveness Probe

**Endpoint:** `GET /health`
**Purpose:** Is the application alive?
**Response:**

```json
{
  "status": "ok",
  "timestamp": "2025-09-30T12:34:56Z"
}
```

**K8s Config:**

```yaml
livenessProbe:
  httpGet:
    path: /health
    port: 8080
  initialDelaySeconds: 10
  periodSeconds: 10
  timeoutSeconds: 2
  failureThreshold: 3
```

#### Readiness Probe

**Endpoint:** `GET /ready`
**Purpose:** Can the application accept traffic?
**Checks:**

- Redis connection is healthy
- Pub/Sub subscriber is connected
- (Optional) Metrics endpoint is responding

**Response:**

```json
{
  "status": "ready",
  "checks": {
    "redis": "ok",
    "pubsub": "ok"
  }
}
```

**K8s Config:**

```yaml
readinessProbe:
  httpGet:
    path: /ready
    port: 8080
  initialDelaySeconds: 5
  periodSeconds: 5
  timeoutSeconds: 2
  failureThreshold: 2
```

### 2. Graceful Shutdown

**Problem:** When K8s terminates a pod, in-flight WebSocket connections should close cleanly.

**Implementation:**

1. Receive SIGTERM signal
2. Stop accepting new connections (remove from service)
3. Wait for existing connections to close (30s timeout)
4. Drain message queue
5. Close Redis connections
6. Exit

**Code Pattern:**

```go
func (s *Server) GracefulShutdown(ctx context.Context) error {
    // Stop accepting new connections
    s.httpServer.SetKeepAlivesEnabled(false)

    // Close all WebSocket connections with close frame
    s.melody.CloseWithMsg([]byte("Server shutting down"))

    // Wait for connections to close (with timeout)
    done := make(chan struct{})
    go func() {
        s.httpServer.Shutdown(ctx)
        close(done)
    }()

    select {
    case <-done:
        return nil
    case <-ctx.Done():
        return ctx.Err()
    }
}
```

**K8s Config:**

```yaml
terminationGracePeriodSeconds: 30
```

### 3. Horizontal Pod Autoscaling (HPA)

**Metric:** CPU utilization (later: custom metrics like WebSocket connections)

**Config:**

```yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: backend-hpa
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: backend
  minReplicas: 3
  maxReplicas: 10
  metrics:
    - type: Resource
      resource:
        name: cpu
        target:
          type: Utilization
          averageUtilization: 70
    - type: Resource
      resource:
        name: memory
        target:
          type: Utilization
          averageUtilization: 80
  behavior:
    scaleDown:
      stabilizationWindowSeconds: 300 # Don't scale down too fast
      policies:
        - type: Percent
          value: 50
          periodSeconds: 60
    scaleUp:
      stabilizationWindowSeconds: 0 # Scale up immediately
      policies:
        - type: Percent
          value: 100
          periodSeconds: 30
```

**Advanced (Phase 5):** Custom metrics HPA based on WebSocket connections

### 4. Resource Management

**Backend Pod Resources:**

```yaml
resources:
  requests:
    cpu: 100m # Minimum guaranteed
    memory: 64Mi
  limits:
    cpu: 1000m # Maximum allowed (1 core)
    memory: 256Mi
```

**Frontend Pod Resources:**

```yaml
resources:
  requests:
    cpu: 100m
    memory: 128Mi
  limits:
    cpu: 500m
    memory: 512Mi
```

**Redis Pod Resources:**

```yaml
resources:
  requests:
    cpu: 100m
    memory: 128Mi
  limits:
    cpu: 500m
    memory: 512Mi
```

### 5. ConfigMaps & Secrets

**ConfigMap (Non-sensitive config):**

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: backend-config
data:
  APP_ENV: "production"
  LOG_LEVEL: "info"
  REDIS_ADDR: "redis:6379"
  REDIS_DB: "0"
  MAX_MESSAGE_SIZE: "4096"
  RATE_LIMIT_PER_MINUTE: "100"
```

**Secret (Sensitive data):**

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: backend-secrets
type: Opaque
data:
  FIREBASE_SERVICE_ACCOUNT: <base64-encoded-json>
  REDIS_PASSWORD: <base64-encoded-password>
```

**Usage in Pod:**

```yaml
envFrom:
  - configMapRef:
      name: backend-config
  - secretRef:
      name: backend-secrets
```

### 6. Service Mesh Readiness (Optional)

**Why:** Future-proof for Istio/Linkerd if scaling beyond 1000 users

**Features:**

- mTLS between services
- Circuit breaking
- Retry policies
- Distributed tracing

**Not implementing now:** Adds complexity for limited current benefit

---
