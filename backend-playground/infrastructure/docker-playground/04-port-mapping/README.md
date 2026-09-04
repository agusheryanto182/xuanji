# Docker Port Mapping

This playground demonstrates how to expose a port from a Docker container to the host machine.

## Goal

Understand the flow:

```text
Host :8080
    ↓
Docker Port Mapping
    ↓
Container :80
    ↓
Nginx
```

## Structure

```text
04-port-mapping/
├── Dockerfile
└── index.html
```

## Practice

Build the image:

```bash
docker build -t my-nginx .
```

### Run Without Port Mapping

```bash
docker run -d --name my-nginx my-nginx
```

Check:

```bash
docker ps
```

Try:

```bash
curl localhost:8080
```

The request should fail because the container port has not been published to the host.

## Run With Port Mapping

Remove the previous container:

```bash
docker rm -f my-nginx
```

Run with port mapping:

```bash
docker run -d   --name my-nginx   -p 8080:80   my-nginx
```

Now:

```bash
curl localhost:8080
```

Expected:

```html
<h1>Hello from my Docker image!</h1>
```

## Port Mapping

The syntax is:

```bash
-p HOST_PORT:CONTAINER_PORT
```

Example:

```bash
-p 8080:80
```

Means:

```text
Host
localhost:8080
      ↓
    Docker
      ↓
Container:80
      ↓
    Nginx
```

The Nginx application is still listening on port `80` inside the container.

Docker forwards traffic from host port `8080` to container port `80`.

## Check Port Mapping

```bash
docker port my-nginx
```

Expected output:

```text
80/tcp -> 0.0.0.0:8080
```

## Important

Port mapping does **not** change the application's port inside the container.

```text
-p 8080:80

8080 → Host port
80   → Container port
```

## Cleanup

```bash
docker rm -f my-nginx
```

## Key Takeaway

```text
-p HOST:CONTAINER
```

Example:

```bash
-p 8080:80
```

Flow:

```text
Browser / curl
      ↓
localhost:8080
      ↓
Docker
      ↓
container:80
      ↓
Nginx
```
