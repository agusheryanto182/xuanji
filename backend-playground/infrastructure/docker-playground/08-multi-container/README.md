# Docker Multi-Container Application

This playground combines Docker fundamentals into a simple multi-container application using Go, PostgreSQL, Redis, and Docker Compose.

## Goal

```text
Docker Network
      │
  ┌───┼────────┐
  ↓   ↓        ↓
 Go  PostgreSQL Redis
```

## Structure

```text
08-multi-container/
├── README.md
├── docker-compose.yml
├── Dockerfile
├── go.mod
└── main.go
```

## Go Version

This playground uses Go `1.27.0`.

```text
Local Go  → 1.27.0
Docker Go → 1.27.0
go.mod    → 1.27
```

## Dockerfile

```dockerfile
FROM golang:1.27.0-alpine

WORKDIR /app

COPY go.mod .
COPY main.go .

RUN go build -o app .

CMD ["./app"]
```

## go.mod

```go
module multi-container

go 1.27
```

## docker-compose.yml

```yaml
services:
  app:
    build: .
    ports:
      - "8080:8080"
    environment:
      POSTGRES_HOST: postgres:5432
      REDIS_HOST: redis:6379
    depends_on:
      - postgres
      - redis

  postgres:
    image: postgres:17
    environment:
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: postgres
      POSTGRES_DB: app

  redis:
    image: redis:8-alpine
```

## Run

```bash
docker compose up -d --build
```

Check:

```bash
docker compose ps
```

You should have:

```text
app
postgres
redis
```

## Test

The Go application is exposed on port `8080`:

```bash
curl localhost:8080
```

Expected:

```text
PostgreSQL: postgres:5432
Redis: redis:6379
```

## Container-to-Container Communication

The application uses service names:

```text
postgres:5432
redis:6379
```

It does not use:

```text
localhost:5432
localhost:6379
```

Inside the `app` container, `localhost` means the `app` container itself.

Docker Compose provides internal DNS for service names.

```text
app
 │
 ├──→ postgres:5432
 │
 └──→ redis:6379
```

## `depends_on`

```yaml
depends_on:
  - postgres
  - redis
```

This controls startup order, but does not guarantee that the dependencies are ready to accept connections.

For real applications, readiness checks and retry strategies may be required.

## Docker Fundamentals Combined

```text
Dockerfile
    ↓
Image
    ↓
Container
    ↓
Port Mapping
    ↓
Network
    ↓
Docker Compose
    ↓
Multi-Container Application
```

Architecture:

```text
Host
 │
 └── localhost:8080
          ↓
       ┌─────┐
       │ app │
       └──┬──┘
          │
   Compose Network
      ┌───┴───┐
      ↓       ↓
 postgres   redis
  :5432      :6379
```

## Stop

```bash
docker compose down
```

Remove containers and volumes:

```bash
docker compose down -v
```

## Key Takeaway

Docker Compose lets multiple containers work together through an internal network.

```text
Go App
  ↓
Docker Network
  ├── PostgreSQL
  └── Redis
```

The application communicates with other containers using their service names instead of `localhost`.