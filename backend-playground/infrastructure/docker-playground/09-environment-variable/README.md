# Docker Environment Variables

This playground demonstrates how to pass configuration into Docker containers using environment variables.

## Goal

```text
.env / Host
    ↓
Docker Compose
    ↓
Container Environment
    ↓
Go Application
```

## Structure

```text
09-environment-variables/
├── README.md
├── Dockerfile
├── docker-compose.yml
├── .env
├── .gitignore
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

## `.env`

```env
APP_ENV=development
DB_HOST=postgres:5432
REDIS_HOST=redis:6379
```

The `.env` file is read by Docker Compose for variable substitution.

## `docker-compose.yml`

```yaml
services:
  app:
    build: .
    ports:
      - "8080:8080"
    environment:
      APP_ENV: ${APP_ENV}
      DB_HOST: ${DB_HOST}
      REDIS_HOST: ${REDIS_HOST}
```

Compose replaces `${APP_ENV}`, `${DB_HOST}`, and `${REDIS_HOST}` with values from `.env`.

## Go Application

Read environment variables with:

```go
os.Getenv("APP_ENV")
os.Getenv("DB_HOST")
os.Getenv("REDIS_HOST")
```

Example:

```go
package main

import (
    "fmt"
    "net/http"
    "os"
)

func main() {
    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        fmt.Println("watashiwa ningen desu!")

        fmt.Fprintln(w, "watashiwa ningen desu!")

        fmt.Fprintf(
            w,
            "PostgreSQL: %s\nRedis: %s\n",
            os.Getenv("DB_HOST"),
            os.Getenv("REDIS_HOST"),
        )
    })

    fmt.Println("server running on :8080")

    http.ListenAndServe(":8080", nil)
}
```

`fmt.Println()` writes to the application's stdout/terminal.

`fmt.Fprintln(w, ...)` and `fmt.Fprintf(w, ...)` write to the HTTP response, so `curl` can display them.

## Run

Build and start:

```bash
docker compose up --build
```

Run in detached mode:

```bash
docker compose up -d --build
```

Check:

```bash
docker compose ps
```

## Test the HTTP Response

The application is exposed on host port `8080`.

```bash
curl localhost:8080
```

Expected:

```text
watashiwa ningen desu!
PostgreSQL: postgres:5432
Redis: redis:6379
```

The message:

```text
watashiwa ningen desu!
```

must be written to the HTTP response using `w`.

## Verify Container Environment

```bash
docker compose exec app env
```

You should find:

```text
APP_ENV=development
DB_HOST=postgres:5432
REDIS_HOST=redis:6379
```

## Debugging Port 8080

If `curl localhost:8080` does not show the latest application changes, make sure the request is reaching the correct application/container.

Check:

```bash
docker compose ps
```

Check which process is listening on port `8080`:

```bash
sudo ss -ltnp | grep :8080
```

Or:

```bash
sudo lsof -i :8080
```

If another process or container is using port `8080`, `curl localhost:8080` may not be reaching the application you expect.

After changing Go source code, rebuild the image:

```bash
docker compose up -d --build
```

## `.env` vs Container Environment

Important distinction:

```text
.env
 ↓
Docker Compose variable substitution
 ↓
environment:
 ↓
Container environment
 ↓
Go os.Getenv()
```

A value existing in `.env` does not by itself mean the application container receives it.

For example:

```env
DB_HOST=postgres:5432
```

must be referenced by Compose:

```yaml
environment:
  DB_HOST: ${DB_HOST}
```

Then the container receives:

```text
DB_HOST=postgres:5432
```

## Secrets

Do not commit sensitive values such as:

```text
DATABASE_PASSWORD
JWT_SECRET
API_KEY
PRIVATE_KEY
```

A common local-development practice is:

```gitignore
.env
```

For production, use an appropriate secrets/configuration management mechanism instead of committing secrets to Git.

## Key Takeaway

```text
.env
  ↓
Docker Compose
  ↓
Container Environment
  ↓
Go Application
```

Environment variables keep configuration separate from application code and Docker images.

Also remember:

```text
fmt.Println()
    ↓
Terminal / stdout

fmt.Fprintln(w, ...)
    ↓
HTTP Response
    ↓
curl / browser
```
