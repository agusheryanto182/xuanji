# Docker Compose

This playground demonstrates how Docker Compose manages multiple containers and the resources they need.

## Goal

Combine the Docker fundamentals learned previously:

```text
Image
Container
Port
Volume
Network
```

using a single `docker-compose.yml`.

## Structure

```text
07-docker-compose/
├── docker-compose.yml
└── index.html
```

## `index.html`

```html
<h1>Hello from Docker Compose!</h1>
```

## `docker-compose.yml`

```yaml
services:
  nginx:
    image: nginx:alpine
    ports:
      - "8080:80"
    volumes:
      - ./index.html:/usr/share/nginx/html/index.html

  postgres:
    image: postgres:17
    environment:
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: postgres
      POSTGRES_DB: compose_test
```

## Run

Start all services:

```bash
docker compose up -d
```

Check:

```bash
docker compose ps
```

You should have:

```text
nginx
postgres
```

Test Nginx:

```bash
curl localhost:8080
```

Expected:

```html
<h1>Hello from Docker Compose!</h1>
```

## Automatic Network

Docker Compose automatically creates a network for the services.

You do not need to manually run:

```bash
docker network create
```

The services can communicate through their service names.

```text
nginx
   ↓
Compose Network
   ↓
postgres
```

From another container, PostgreSQL can be reached using:

```text
postgres:5432
```

not:

```text
localhost:5432
```

Remember that `localhost` inside a container refers to that same container.

## Inspect the Network

List networks:

```bash
docker network ls
```

Inspect the Compose network:

```bash
docker network inspect 07-docker-compose_default
```

The exact generated network name can vary depending on the project directory.

## Stop

Stop and remove the containers:

```bash
docker compose down
```

The images remain.

To also remove Compose-managed volumes:

```bash
docker compose down -v
```

## Docker Compose vs Manual Docker Commands

Without Compose:

```text
docker network create
docker run ...
docker run ...
docker ...
```

With Compose:

```text
docker-compose.yml
        ↓
docker compose up
        ↓
All services
        ↓
Network + Containers + Configuration
```

## Key Takeaway

Docker Compose lets you define and run a multi-container application from one configuration file.

```text
docker-compose.yml
       ↓
docker compose up
       ↓
┌─────────────────┐
│ Compose Network │
└────────┬────────┘
         │
    ┌────┴────┐
    ↓         ↓
  nginx    postgres
```

It is especially useful for local development and multi-container environments.
