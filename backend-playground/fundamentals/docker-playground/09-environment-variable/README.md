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
Application
```

## Structure

```text
09-environment-variables/
├── README.md
├── docker-compose.yml
├── .env
└── main.go
```

## `.env`

```env
APP_ENV=development
DB_HOST=postgres:5432
REDIS_HOST=redis:6379
```

The `.env` file provides values that Docker Compose can use for variable substitution.

## `docker-compose.yml`

```yaml
services:
  app:
    build: .
    environment:
      APP_ENV: ${APP_ENV}
      DB_HOST: ${DB_HOST}
      REDIS_HOST: ${REDIS_HOST}
```

Compose reads the values and passes them into the container environment.

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
    "os"
)

func main() {
    fmt.Println("APP_ENV:", os.Getenv("APP_ENV"))
    fmt.Println("DB_HOST:", os.Getenv("DB_HOST"))
    fmt.Println("REDIS_HOST:", os.Getenv("REDIS_HOST"))
}
```

## Run

```bash
docker compose up --build
```

The application receives:

```text
APP_ENV=development
DB_HOST=postgres:5432
REDIS_HOST=redis:6379
```

## Compose Variable Substitution

```yaml
environment:
  DB_HOST: ${DB_HOST}
```

`${DB_HOST}` is replaced by Compose using the value from `.env` or another supported environment source.

The resulting value becomes an environment variable inside the container:

```text
DB_HOST=postgres:5432
```

## `environment` vs `.env`

`.env` is commonly used as a source of values for Compose variable substitution.

```text
.env
 ↓
Compose
 ↓
environment:
 ↓
Container
 ↓
Application
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

For production, use an appropriate secrets/configuration management mechanism rather than committing secrets to Git.

## Inspect Container Environment

For a running Compose service:

```bash
docker compose exec app env
```

Or:

```bash
docker exec <container-name> env
```

## Key Takeaway

```text
Configuration
     ↓
Environment Variables
     ↓
Docker Compose
     ↓
Container
     ↓
Application
```

Keep configuration separate from application code and Docker images.
