# asocial - Original Architecture (Deprecated)

## Technology Stack

### Backend (Go 1.22.1)

- **HTTP/WebSocket:** Gin + Melody
- **Message Broker:** Apache Kafka + Zookeeper
- **Pub/Sub:** Watermill + Watermill-Kafka
- **Config:** Viper
- **DI:** Google Wire

### Frontend (Next.js 14)

- **Framework:** Next.js 14 (App Router) + React 18 + TypeScript
- **Styling:** TailwindCSS
- **Canvas:** react-zoom-pan-pinch
- **IDs:** uniqid

### Infrastructure

- **Reverse Proxy:** Traefik v2.11
- **Containers:** Docker Compose v3
- **Message Broker:** Kafka + Zookeeper (Confluent images)

---

## Architecture Overview

```
┌──────────────────── Docker Compose ─────────────────────┐
│                                                         │
│  Traefik (:80, :8080)                                   │
│    ├─ /* → Frontend (Next.js :3000)                     │
│    └─ /api/chat → Backend (Go :3001)                    │
│                         │                               │
│                    ┌────▼─────┐                         │
│                    │  Kafka   │  Topic: rc.msg.pub      │
│                    │  :9092   │  Pub/Sub fanout         │
│                    └────┬─────┘                         │
│                         │                               │
│                    ┌────▼─────┐                         │
│                    │Zookeeper │  Kafka coordination     │
│                    │  :2181   │                         │
│                    └──────────┘                         │
│                                                         │
└─────────────────────────────────────────────────────────┘

Access: http://localhost/  (Frontend)
        http://localhost:8080/  (Traefik Dashboard)
```

### Directory Structure

```
asocial/
├── cmd/                            # CLI (chat server command)
├── pkg/
│   ├── chat/                       # WebSocket handlers, message service
│   ├── common/                     # Server abstraction
│   ├── config/                     # Config management
│   └── infra/                      # Kafka setup
├── wire/                           # Dependency injection (Wire)
├── frontend/
│   ├── src/
│   │   ├── app/                    # Pages (/, /chat)
│   │   └── components/             # Messages, LocalMessage, ExternalMessage
│   └── dev.Dockerfile
├── docker-compose.yml
├── go.Dockerfile
├── .asocial.yaml                   # App config
└── go.mod
```

---

## Key Components

### Backend (Go)

**WebSocket Handler** (`pkg/chat/http_api.go`)

- `GET /api/chat?uid=<user_id>` - Upgrades to WebSocket
- On message: publishes to Kafka topic `rc.msg.pub`
- Session stores: `channelID` (default: "default"), `userID`

**Message Service** (`pkg/chat/service.go`)

- `MessageService`: Publishes messages to Kafka
- `MessageSubscriber`: Consumes from Kafka, broadcasts via WebSocket
- Filtering: Same channel, exclude sender

**Kafka Setup** (`pkg/infra/kafka.go`)

- Watermill pub/sub with retry middleware
- Topic: `rc.msg.pub`

**Config** (`.asocial.yaml`)

```yaml
chat:
  http:
    server:
      port: "3001"
  message:
    maxSizeByte: 4096
kafka:
  addrs: kafka:9092
```

**Dependency Injection** (`wire/`)

- Google Wire auto-generates dependency graph

### Frontend (Next.js)

**Chat Canvas** (`src/components/Messages.tsx`)

- WebSocket: `ws://<host>/api/chat?uid=<random_id>`
- Click canvas → Creates local message input
- On keystroke → Sends to backend
- On receive → Renders external message
- Auto-remove after 5 seconds
- Max 5 messages per user

**Message Format:**

```json
{
  "user_id": "abc123",
  "message_id": "msg456",
  "payload": "Hello",
  "position": { "x": 100, "y": 200 },
  "channel_id": "default"
}
```

**Components:**

- `LocalMessage`: Editable input (user's messages)
- `ExternalMessage`: Read-only div (others' messages)

## Running Locally

### Prerequisites

- Docker & Docker Compose installed
- Ports available: 80, 3000, 3001, 8080, 9092, 29092
- 4GB+ RAM available for containers

### Option 1: Docker Compose (Recommended)

**Start all services:**

```bash
cd /path/to/asocial
docker-compose up --build
```

**What happens:**

1. Zookeeper starts (Kafka dependency)
2. Kafka starts and waits for Zookeeper
3. Backend builds and starts, connects to Kafka
4. Frontend builds and starts
5. Traefik starts and routes traffic

**Access points:**

- Frontend: http://localhost/
- Traefik Dashboard: http://localhost:8080/
- Backend WebSocket: ws://localhost/api/chat

**View logs:**

```bash
# All services
docker-compose logs -f

# Specific service
docker-compose logs -f backend
docker-compose logs -f frontend
docker-compose logs -f kafka
```

### Option 2: Manual (Backend Only)

If you only want to test the backend:

**Terminal 1 - Kafka:**

```bash
docker run -d --name zookeeper -p 2181:2181 confluentinc/cp-zookeeper:latest \
  -e ZOOKEEPER_CLIENT_PORT=2181

docker run -d --name kafka -p 29092:29092 -p 9092:9092 \
  --link zookeeper \
  confluentinc/cp-kafka:latest \
  -e KAFKA_BROKER_ID=1 \
  -e KAFKA_ZOOKEEPER_CONNECT=zookeeper:2181 \
  -e KAFKA_ADVERTISED_LISTENERS=PLAINTEXT://localhost:29092 \
  -e KAFKA_OFFSETS_TOPIC_REPLICATION_FACTOR=1
```

**Terminal 2 - Backend:**

```bash
cd /path/to/asocial
export KAFKA_ADDRS=localhost:29092
go run main.go chat
```

Backend will be available at: http://localhost:3001/api/chat

### Option 3: Manual (Frontend Only)

**Terminal 1:**

```bash
cd /path/to/asocial/frontend
npm install
npm run dev
```

Frontend will be available at: http://localhost:3000

**Note:** WebSocket will fail to connect without backend running.

---

## Differences from New Architecture

This table shows what will change in the refactor:

| Aspect             | Current (Old)                    | Future (New)                             |
| ------------------ | -------------------------------- | ---------------------------------------- |
| **Message Broker** | Kafka + Zookeeper (3 containers) | Redis (1 container)                      |
| **Resource Usage** | ~1.5 GB memory                   | ~50 MB memory                            |
| **Startup Time**   | 60+ seconds                      | <5 seconds                               |
| **Architecture**   | Mixed concerns                   | Clean Architecture (layers)              |
| **Orchestration**  | Docker Compose                   | Kubernetes                               |
| **Scaling**        | Manual                           | Horizontal Pod Autoscaler                |
| **Health Checks**  | None                             | Liveness + Readiness probes              |
| **Monitoring**     | None                             | Prometheus + Grafana                     |
| **Logging**        | Unstructured                     | Structured (JSON + trace IDs)            |
| **Testing**        | None                             | Unit + Integration + E2E                 |
| **CI/CD**          | Manual builds                    | GitHub Actions                           |
| **Authentication** | Partial/broken                   | To be decided (likely removed initially) |
| **Rate Limiting**  | None                             | Redis-based sliding window               |
| **Config**         | YAML file                        | ConfigMaps + Secrets                     |
| **Deployment**     | docker-compose up                | kubectl apply / helm install             |

---
