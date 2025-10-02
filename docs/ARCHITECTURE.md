# asocial - System Architecture

> A real-time collaborative canvas chat application

## Deployment Architectures

### Kubernetes (Minikube)

```
┌───────────────────────────────────────────────────────────┐
│                    Minikube Cluster                       │
│                                                           │
│  ┌─────────────────────────────────────────────────────┐  │
│  │          Ingress-NGINX Controller                   │  │
│  │  • Routes /api/* → backend                          │  │
│  │  • Routes /* → frontend                             │  │
│  │  • WebSocket support (3600s timeout)                │  │
│  └───────┬─────────────────────────┬───────────────────┘  │
│          │                         │                      │
│  ┌───────▼─────────┐      ┌────────▼──────────┐           │
│  │ Frontend Svc    │      │  Backend Svc      │           │
│  │ ClusterIP:3000  │      │  ClusterIP:3001   │           │
│  └───────┬─────────┘      └────────┬──────────┘           │
│          │                         │                      │
│  ┌───────▼─────────┐      ┌────────▼──────────┐           │
│  │ Frontend Pods   │      │  Backend Pods     │           │
│  │ • Next.js 15    │      │  • Go/Gin/Melody  │           │
│  │ • 2 replicas    │      │  • 3 replicas     │           │
│  │ • Port 3000     │      │  • Port 3001      │           │
│  │                 │      │  • Health probes  │           │
│  └─────────────────┘      └────────┬──────────┘           │
│                                    │                      │
│                           ┌────────▼──────────┐           │
│                           │  Redis Service    │           │
│                           │  ClusterIP:6379   │           │
│                           └────────┬──────────┘           │
│                                    │                      │
│                           ┌────────▼──────────┐           │
│                           │ Redis StatefulSet │           │
│                           │ • Redis 7-alpine  │           │
│                           │ • 1 replica       │           │
│                           │ • Persistent (1Gi)│           │
│                           │ • Pub/Sub enabled │           │
│                           └───────────────────┘           │
│                                                           │
└───────────────────────────────────────────────────────────┘
                          ▲
                          │ minikube tunnel
                          │ (maps to localhost:80)
                          │
                    ┌─────┴──────┐
                    │   Browser  │
                    └────────────┘
```

### Docker Compose

```
┌────────────────────────────────────────────────────────────┐
│                    Docker Network                          │
│                                                            │
│  ┌─────────────────────────────────────────────────────┐   │
│  │              Traefik Reverse Proxy                  │   │
│  │  • Routes /api/* → backend:3001                     │   │
│  │  • Routes /* → frontend:3000                        │   │
│  │  • Port 80 (host)                                   │   │
│  │  • Dashboard: :8080                                 │   │
│  └───────┬─────────────────────────┬───────────────────┘   │
│          │                         │                       │
│  ┌───────▼─────────┐      ┌────────▼──────────┐            │
│  │  Frontend       │      │   Backend         │            │
│  │  • Next.js 15   │      │   • Go/Gin/Melody │            │
│  │  • Port 3000    │      │   • Port 3001     │            │
│  └─────────────────┘      └────────┬──────────┘            │
│                                    │                       │
│                           ┌────────▼──────────┐            │
│                           │      Redis        │            │
│                           │  • Port 6379      │            │
│                           │  • Pub/Sub        │            │
│                           │  • Volume mount   │            │
│                           └───────────────────┘            │
│                                                            │
└────────────────────────────────────────────────────────────┘
                          ▲
                          │ localhost:80
                          │
                    ┌─────┴──────┐
                    │   Browser  │
                    └────────────┘
```

## Message Flow

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

## Component Details

**Backend (Go):**

- **HTTP Layer (Gin)**: Routes WebSocket upgrades, health checks, API endpoints
- **WebSocket Handler (Melody)**: Manages WebSocket connections, broadcasts messages
- **Message Service**: Validates messages, coordinates pub/sub, manages user presence
- **Redis Pub/Sub**: Publishes messages to channels, subscribes for broadcasts
- **Health Probes**: `/health` (liveness), `/ready` (readiness - checks Redis)

**Frontend (Next.js 15):**

- **State Management (Zustand)**: Manages messages, users, WebSocket connection state
- **WebSocket Client**: Connects to `/api/chat`, sends/receives messages
- **Canvas Rendering**: Displays messages at user-specified positions with fade effects

**Redis:**

- **Pub/Sub**: Broadcasts messages across backend replicas
- **Presence Tracking**: Stores active users per channel with TTL expiry
- **Persistence**: AOF enabled for data durability

**Networking:**

- **K8s Ingress**: NGINX controller routes by path, supports WebSocket upgrades
- **Services**: ClusterIP for internal load balancing
- **Minikube Tunnel**: Exposes Ingress to localhost

---
