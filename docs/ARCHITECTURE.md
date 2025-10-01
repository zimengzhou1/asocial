# Asocial - System Architecture

> A real-time collaborative canvas chat application

---

## 🚀 Current Implementation Status

**Phase 1 Complete** (Backend Refactor with Redis + User Presence)

✅ **Implemented:**

- Clean architecture with domain/service/handler layers
- Redis pub/sub for real-time messaging (replaced Kafka)
- **User presence tracking with Redis (TTL-based with heartbeat)**
- **Frontend state management with Zustand**
- WebSocket handler with Melody
- Health check endpoints (`/health`, `/ready`)
- Structured logging with Go's slog
- Manual dependency injection
- Docker Compose with Traefik, Redis, Backend, Frontend
- **Multi-message type support (chat, presence events, user sync)**

📋 **Next Phases:**

- Phase 2: Kubernetes deployment (manifests, ConfigMaps, Services)
- Phase 3: Re-add authentication (Firebase/NextAuth)
- Phase 4: Observability (Prometheus, Grafana)
- Phase 5: Production hardening

---

## Table of Contents

- [Current Implementation Status](#-current-implementation-status)
- [System Architecture](#system-architecture)
- [Technology Stack](#technology-stack)
- [Architecture Decisions](#architecture-decisions)
- [Current Implementation](#current-implementation)
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
│                                   │   │ • Presence (TTL + Sets)  │ │   │
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

### User Presence Flow (join)

```
User A                 Frontend                  Backend                   Redis                     Backend                 Frontend (User B)
  │                        │                         │                         │                         │                         │
  │ 1. Open App            │                         │                         │                         │                         │
  ├───────────────────────>│                         │                         │                         │                         │
  │                        │                         │                         │                         │                         │
  │                        │ 2. WebSocket Connect    │                         │                         │                         │
  │                        ├────────────────────────>│                         │                         │                         │
  │                        │                         │                         │                         │                         │
  │                        │                         │ 3. Add User A to Set &  │                         │                         │
  │                        │                         │    Set 5-min Expiry     │                         │                         │
  │                        │                         │ (SADD, SETEX)           │                         │                         │
  │                        │                         ├────────────────────────>│                         │                         │
  │                        │                         │                         │                         │                         │
  │                        │                         │ 4. Get All Users in Set │                         │                         │
  │                        │                         │ (SMEMBERS)              │                         │                         │
  │                        │                         ├────────────────────────>│                         │                         │
  │                        │                         │                         │ 5. Returns              │                         |
  │                        │                         │<────────────────────────┤    [User A, User B]│    │                         │
  │                        │                         │                         │                         │                         │
  │                        │ 6. Send ONLY to User A: │                         │                         │                         │
  │                        │    user_sync message    │                         │                         │                         │
  │                        │    {users: [A, B]}      │                         │                         │                         │
  │                        │<────────────────────────┤                         │                         │                         │
  │                        │                         │                         │                         │                         │
  │                        │                         │ 7. Publish to Channel:  │                         │                         │
  │                        │                         │    user_joined(A) event │                         │                         │
  │                        │                         ├────────────────────────>│                         │                         │
  │                        │                         │                         │                         │                         │
  │                        │                         │                         │ 8. Receives Message &   │                         │
  │                        │                         │                         │    Broadcast to ALL     │                         │
  │                        │                         │                         │<────────────────────────┤                         │
  │                        │                         │                         │   (Except User A)       │                         │
  │                        │                         │                         │                         ├────────────────────────>│ 9. Receives user_joined(A)
  │                        │                         │                         │                         │                         │    & updates user list
  │                        │                         │                         │                         │                         │
  │  [Heartbeat Loop]      │                         │ 10. Every 60s, Refresh  │                         │                         │
  │                        │                         │    User A's 5-min Expiry│                         │                         │
  │                        │                         │    (EXPIRE)             │                         │                         │
  │                        │                         ├────────────────────────>│                         │                         │
  │                        │                         │                         │                         │                         │
```

### User Presence Flow (leave)

```
User A                 Frontend                  Backend                   Redis                     Backend                 Frontend (User B)
  │                        │                         │                         │                         │                         │
  │ 1. Close Browser       │                         │                         │                         │                         │
  ├───────────────────────>│                         │                         │                         │                         │
  │                        │                         │                         │                         │                         │
  │                        │ 2. WebSocket Connection │                         │                         │                         │
  │                        │    Closes               │                         │                         │                         │
  │                        ├────────────────────────>│                         │                         │                         │
  │                        │                         │                         │                         │                         │
  │                        │                         │ 3. Remove User A &      │                         │                         │
  │                        │                         │    Delete Expiry Key    │                         │                         │
  │                        │                         │ (SREM, DEL)             │                         │                         │
  │                        │                         ├────────────────────────>│                         │                         │
  │                        │                         │                         │                         │                         │
  │                        │                         │ 4. Publish to Channel:  │                         │                         │
  │                        │                         │    user_left(A) event   │                         │                         │
  │                        │                         ├────────────────────────>│                         │                         │
  │                        │                         │                         │                         │                         │
  │                        │                         │                         │ 5. Receives Message &   │                         │
  │                        │                         │                         │    Broadcast to ALL     │                         │
  │                        │                         │                         │<────────────────────────┤                         │
  │                        │                         │                         │                         │                         │
  │                        │                         │                         │                         ├────────────────────────>│ 6. Receives user_left(A)
  │                        │                         │                         │                         │                         │    & updates user list
  │                        │                         │                         │                         │                         │

```

**Key Features:**

- **TTL-based cleanup**: Users auto-removed after 5 minutes of inactivity
- **Heartbeat**: Every 60 seconds, backend refreshes user TTL
- **Initial sync**: New users receive complete user list immediately
- **Real-time events**: Join/leave events broadcasted to all users in channel
- **Per-user goroutines**: Each connection has dedicated heartbeat goroutine

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

| Component                | Technology               | Purpose                              | Version | Status     |
| ------------------------ | ------------------------ | ------------------------------------ | ------- | ---------- |
| **HTTP Framework**       | Gin                      | Fast HTTP router, middleware support | v1.9+   | ✅         |
| **WebSocket**            | Melody                   | WebSocket library built on Gorilla   | v1.2+   | ✅         |
| **Pub/Sub**              | go-redis/redis/v9        | Redis client with pub/sub support    | v9.14+  | ✅         |
| **Config**               | Viper                    | Configuration management             | v1.18+  | ✅         |
| **Logging**              | slog                     | Structured logging (Go 1.21+)        | stdlib  | ✅         |
| **Dependency Injection** | Manual                   | Constructor-based DI (removed Wire)  | -       | ✅         |
| **UUID**                 | google/uuid              | Unique ID generation                 | v1.6+   | ✅         |
| **Metrics**              | prometheus/client_golang | Prometheus metrics exporter          | v1.19+  | 📋 Phase 4 |
| **Testing**              | testify                  | Assertions and mocking               | v1.9+   | 📋 Future  |
| **Mocking**              | gomock                   | Mock generation                      | v1.6+   | 📋 Future  |
| **Validation**           | go-playground/validator  | Struct validation                    | v10.19+ | 📋 Future  |

### Frontend (Next.js)

| Component        | Technology           | Purpose                         | Version | Status |
| ---------------- | -------------------- | ------------------------------- | ------- | ------ |
| **Framework**    | Next.js              | React framework with SSR        | v14.1+  | ✅     |
| **Language**     | TypeScript           | Type safety                     | v5.0+   | ✅     |
| **UI Library**   | React                | UI components                   | v18.0+  | ✅     |
| **State**        | Zustand              | Lightweight state management    | v4.5+   | ✅     |
| **Styling**      | TailwindCSS          | Utility-first CSS               | v3.3+   | ✅     |
| **Canvas**       | @use-gesture/react   | Pan/zoom/pinch gestures         | v10.3+  | ✅     |
| **Auth**         | NextAuth.js          | Authentication                  | v5.0+   | 📋     |
| **Firebase**     | Firebase SDK         | User authentication             | v10.11+ | 📋     |
| **WebSocket**    | native WebSocket API | Real-time communication         | -       | ✅     |
| **Custom Hooks** | useWebSocket         | WebSocket connection management | -       | ✅     |

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

| Component       | Technology    | Purpose                         |
| --------------- | ------------- | ------------------------------- |
| **Pub/Sub**     | Redis Pub/Sub | Real-time message broadcasting  |
| **Presence**    | Redis Sets    | User presence tracking with TTL |
| **Cache**       | Redis         | Session storage, rate limiting  |
| **Persistence** | Redis RDB/AOF | Optional message persistence    |

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

## Current Implementation

### Actual Directory Structure (Phase 1)

```
asocial/
├── cmd/
│   └── server/
│       └── main.go                    # ✅ Application entry point, manual DI
│
├── internal/                          # ✅ Private application code
│   │
│   ├── domain/                        # ✅ Business entities
│   │   ├── message.go                 # Message entity with Position
│   │   └── errors.go                  # Domain-specific errors
│   │
│   ├── service/                       # ✅ Business logic layer
│   │   └── message_service.go         # Message publishing & broadcasting
│   │
│   ├── pubsub/                        # ✅ Pub/Sub abstraction
│   │   └── redis_pubsub.go            # Redis pub/sub implementation
│   │
│   ├── handler/                       # ✅ HTTP/WebSocket handlers
│   │   ├── websocket.go               # WebSocket upgrade & message handling
│   │   └── health.go                  # Health check endpoints
│   │
│   └── config/                        # ✅ Configuration management
│       └── config.go                  # Viper-based config loader
│
├── frontend/                          # ✅ Next.js application
│   ├── src/
│   │   ├── app/                       # App router pages
│   │   ├── components/                # React components
│   │   └── lib/                       # Utilities
│   ├── public/
│   ├── package.json
│   └── tsconfig.json
│
├── docs/                              # ✅ Documentation
│   ├── ARCHITECTURE.md                # This file
│   ├── OLD_ARCHITECTURE.md            # Legacy system docs
│   └── PLANNING.md                    # Implementation phases
│
├── config.yaml                        # ✅ Application configuration
├── docker-compose.yml                 # ✅ Local development stack
├── go.Dockerfile                      # ✅ Backend container
├── go.mod                             # ✅ Go dependencies
├── go.sum
├── main.go                            # ✅ Entry point
└── README.md
```

### Implemented Components

**Backend (`internal/`):**

1. **Domain Layer** (`internal/domain/`)

   - `Message` struct with Type, MessageID, ChannelID, UserID, Payload, Position, Users, Timestamp
   - `MessageType` enum: `chat`, `user_joined`, `user_left`, `user_sync`
   - `Position` struct with X, Y coordinates (float64 for sub-pixel precision)
   - Helper functions: `NewMessage`, `NewUserJoinedMessage`, `NewUserLeftMessage`, `NewUserSyncMessage`
   - JSON encoding/decoding methods
   - Domain errors

2. **Service Layer** (`internal/service/`)

   - `MessageService`: Publishes messages, starts subscriber, broadcasts to WebSocket clients
   - `PubSubClient` interface with presence operations
   - Filters broadcasts by channel and message type (excludes sender for chat, includes all for presence)
   - UUID generation for message IDs

3. **PubSub Layer** (`internal/pubsub/`)

   - `RedisPubSub`: Redis client wrapper with pub/sub and presence operations
   - Presence tracking: `AddUserToChannel`, `RemoveUserFromChannel`, `RefreshUserPresence`, `GetChannelUsers`
   - TTL-based cleanup (5-minute expiry)
   - Connection health checking
   - Structured logging

4. **Handler Layer** (`internal/handler/`)

   - `WebSocketHandler`: Manages WebSocket connections, handles messages, presence lifecycle
   - User connect: Add to Redis set, send user sync, publish join event, start heartbeat
   - User disconnect: Remove from Redis set, publish leave event
   - Heartbeat goroutine (60s interval) refreshes user presence TTL
   - `HealthHandler`: Liveness (`/health`) and readiness (`/ready`) probes

5. **Config Layer** (`internal/config/`)
   - Viper-based configuration loading from YAML
   - Environment variable overrides
   - Validation and defaults

**Frontend (`frontend/src/`):**

1. **State Management** (`src/stores/`)

   - `chatStore` (Zustand): Manages messages, users, viewport state
   - Actions: `addMessage`, `updateMessage`, `removeMessage`, `fadeOutMessage`, `addUser`, `removeUser`, `setViewport`
   - Auto-generates user colors from user IDs
   - Automatic message fade-out after 5 seconds

2. **Custom Hooks** (`src/hooks/`)

   - `useWebSocket`: WebSocket connection management with reconnection logic
   - Handles multiple message types: chat, user_joined, user_left, user_sync
   - Uses refs to prevent unnecessary re-renders
   - Connection state tracking

3. **Components** (`src/components/`)
   - `Canvas/CanvasViewport`: Pan/zoom/pinch gestures with proper coordinate transformation
   - `Messages`: Renders chat messages at canvas positions
   - `Layout/ChatLayout`: App layout with user count display

**Routes:**

- `GET /health` - Liveness probe (always returns 200)
- `GET /ready` - Readiness probe (checks Redis connection)
- `GET /api/chat?uid={user_id}` - WebSocket upgrade endpoint with user ID

### Planned Directory Structure (Future Phases)

```
asocial/
├── internal/
│   ├── middleware/                    # 📋 Phase 3: Auth, logging, metrics
│   ├── repository/                    # 📋 Future: If we need persistent storage
│   └── metrics/                       # 📋 Phase 4: Prometheus metrics
│
├── pkg/                               # 📋 Future: Reusable libraries
│   ├── logger/
│   └── validator/
│
├── tests/                             # 📋 Future: Comprehensive test suite
│   ├── unit/
│   ├── integration/
│   └── e2e/
│
├── k8s/                               # 📋 Phase 2: Kubernetes manifests
│   ├── base/
│   │   ├── deployment.yaml
│   │   ├── service.yaml
│   │   ├── configmap.yaml
│   │   └── ingress.yaml
│   └── overlays/
│       ├── dev/
│       └── prod/
│
└── scripts/                           # 📋 Future: Automation scripts
    ├── deploy.sh
    └── test.sh
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
