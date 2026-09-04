# Dockerfile

This playground demonstrates how to create a custom Docker image using a `Dockerfile`.

## Goal

Understand the basic flow:

```text
Dockerfile
    ↓
docker build
    ↓
Custom Image
    ↓
docker run
    ↓
Container
```

## Structure

```text
03-dockerfile/
├── Dockerfile
└── index.html
```

## Practice

### 1. Create `index.html`

```html
<h1>Hello from my Docker image!</h1>
```

### 2. Create `Dockerfile`

```dockerfile
FROM nginx:alpine

COPY index.html /usr/share/nginx/html/index.html
```

## Dockerfile Instructions

### `FROM`

```dockerfile
FROM nginx:alpine
```

Defines the base image used to build the new image.

```text
nginx:alpine
      ↓
Base Image
```

### `COPY`

```dockerfile
COPY index.html /usr/share/nginx/html/index.html
```

Copies a file from the build context into the image.

```text
Host
 ↓
index.html
 ↓ COPY
Docker Image
```

## Build the Image

From the `03-dockerfile` directory:

```bash
docker build -t my-nginx .
```

Check:

```bash
docker images
```

You should see:

```text
my-nginx
```

## Run the Container

```bash
docker run -d --name my-nginx my-nginx
```

Check:

```bash
docker ps
```

## Verify the File

The `index.html` file should exist inside the container:

```bash
docker exec -it my-nginx cat /usr/share/nginx/html/index.html
```

Expected:

```text
<h1>Hello from my Docker image!</h1>
```

This confirms that the `COPY` instruction worked.

## Important Flow

```text
Dockerfile
     ↓
docker build
     ↓
my-nginx Image
     ↓
docker run
     ↓
my-nginx Container
```

## Cleanup

Remove the container:

```bash
docker rm -f my-nginx
```

Remove the image:

```bash
docker rmi my-nginx
```

## Key Takeaway

```text
FROM
→ Base image

COPY
→ Copy files into the image

docker build
→ Build an image

docker run
→ Create and run a container from the image
```

For now, focus on the basic Dockerfile instructions. More advanced instructions such as `CMD`, `ENTRYPOINT`, multi-stage builds, and image optimization can be learned later when needed.
