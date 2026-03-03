# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

A ride-sharing microservices application in Go (v1.23.0) with a Next.js 15 frontend. Services communicate via gRPC (synchronous) and RabbitMQ (asynchronous/event-driven).

## Commands

### Local Development
```bash
tilt up                  # Start all services in local Kubernetes (hot-reload)
kubectl get pods         # Monitor pod status
minikube dashboard       # Visual dashboard
```

### Build (Linux/amd64 binaries for containers)
```bash
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o build/api-gateway ./services/api-gateway
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o build/trip-service ./services/trip-service/cmd
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o build/driver-service ./services/driver-service
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o build/payment-service ./services/payment-service/cmd
```

### Testing
```bash
go test ./...                              # All tests
go test ./services/trip-service/...        # Single service
go test -v -run TestFunctionName ./...     # Single test
```

### Proto Code Generation
```bash
make generate-proto   # Regenerate Go code from proto/ definitions
```

### Frontend
```bash
cd web && npm run dev     # Dev server (Next.js + Turbopack)
cd web && npm run lint    # ESLint
cd web && npm run build   # Production build
```

## Architecture

### Services

| Service | Port | Protocol | Description |
|---------|------|----------|-------------|
| api-gateway | 8081 | HTTP/WS | External entry point; REST + WebSocket |
| trip-service | 9093 | gRPC | Trip booking, route planning (OSRM), fare estimation |
| driver-service | 9092 | gRPC | Driver registration, location tracking (geohash) |
| payment-service | 9004 | gRPC | Stripe integration (partially implemented) |

### Communication Patterns

**gRPC (synchronous)**: API Gateway calls trip-service and driver-service via generated clients in `shared/proto/`.

**RabbitMQ (async)**: Topic exchange named `"trip"`. Routing key constants are in `shared/contracts/amqp.go`. Key flows:
- Trip created → `trip.event.created` → driver-service finds available driver
- Driver accepts → `driver.cmd.trip_accept` → trip-service updates status → `trip.event.driver_assigned`
- WebSocket clients subscribed to `/ws/riders` and `/ws/drivers` receive live updates via RabbitMQ events

**HTTP REST** (external, via api-gateway):
- `POST /trip/preview` — returns route + fare options
- `POST /trip/start` — books trip with selected fare

**WebSocket**: `/ws/riders` and `/ws/drivers` for real-time updates to frontend clients.

### Shared Libraries (`shared/`)

- `contracts/` — AMQP routing key constants (single source of truth for message names)
- `messaging/` — RabbitMQ client wrapper (connect, publish, consume)
- `env/` — Environment variable helpers with defaults
- `proto/` — Generated protobuf/gRPC code (do not edit manually)
- `types/`, `retry/`, `util/` — Common utilities

### Code Architecture Patterns

**Trip Service and Payment Service** follow Clean Architecture:
```
domain/       → interfaces (TripRepository, TripService)
service/      → business logic implementations
infrastructure/
  grpc/       → gRPC server handlers
  repository/ → data access (currently in-memory, MongoDB driver ready)
  events/     → RabbitMQ publishers & consumers
```

**Driver Service** uses a flat structure — all logic in the main package (`service.go`, `grpc_handler.go`, `trip_consumer.go`).

**API Gateway** is a single `main.go` — no layering, direct orchestration.

### Environment Variables

Defined via `shared/env` with defaults:
```
HTTP_ADDR          (default: ":8081")
GRPC_ADDR          (default: varies per service)
RABBITMQ_URI       (default: "amqp://guest:guest@rabbitmq:5672/")
STRIPE_SECRET_KEY  (required for payment-service)
APP_URL            (default: "http://localhost:3000")
```

### Infrastructure

- **Dev**: `infra/development/k8s/` — K8s manifests, orchestrated by `Tiltfile`
- **Prod**: `infra/production/k8s/` — GKE-ready manifests (GCP deployment)
- **Dockerfiles**: Alpine-based images; binaries built locally then copied in
- **RabbitMQ**: StatefulSet with 1Gi persistent volume; 5 named queues with topic bindings
- **Observability**: OpenTelemetry + Jaeger tracing already configured

### Key Domain Details

- **Fare types**: `suv` ($3 base), `sedan` ($2), `van` ($4), `luxury` ($10) — dynamic pricing adds distance/time cost
- **Location tracking**: Geohash-based (`github.com/mmcloughlin/geohash`)
- **Routing**: OSRM public API (`router.project-osrm.org`) for route/distance/duration
- **Storage**: In-memory with `sync.RWMutex`; MongoDB driver imported and ready in trip-service
- **No tests exist yet** — domain interfaces are designed for easy mocking via dependency injection
