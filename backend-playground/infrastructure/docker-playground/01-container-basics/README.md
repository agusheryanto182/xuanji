# Docker Container Basics

This playground demonstrates the fundamentals of Docker containers.

## Goal

Understand the basic flow:

```text
Docker Image
     ↓
docker run
     ↓
Container
     ↓
Process
```

## Practice

### 1. Pull an Image

```bash
docker pull nginx
```

Check images:

```bash
docker images
```

### 2. Run a Container

```bash
docker run -d --name docker-basic-nginx nginx
```

Check:

```bash
docker ps
```

### 3. List Containers

```bash
docker ps
docker ps -a
```

```text
docker ps
→ running containers

docker ps -a
→ all containers
```

### 4. Execute a Command Inside a Container

```bash
docker exec -it docker-basic-nginx bash
```

Inside:

```bash
ls
```

Exit:

```bash
exit
```

### 5. Stop

```bash
docker stop docker-basic-nginx
```

The container is stopped but still exists:

```bash
docker ps -a
```

### 6. Start Again

```bash
docker start docker-basic-nginx
```

Important:

```text
stop
  ↓
container still exists

start
  ↓
container runs again
```

### 7. Remove

```bash
docker rm docker-basic-nginx
```

The container is removed, but the `nginx` image remains.

## Image vs Container

```text
Image
→ Blueprint / Template

Container
→ Running instance created from an Image
```

One image can create multiple containers:

```text
nginx Image
   ├── Container #1
   ├── Container #2
   └── Container #3
```

## Important Commands

```bash
docker pull
docker run
docker ps
docker ps -a
docker exec
docker stop
docker start
docker rm
```

## Key Takeaway

```text
Image
  ↓ docker run
Container
  ↓
Process
```

A container can be stopped and started again without recreating the image.
