# Docker Image vs Container

This playground demonstrates the difference between a Docker **Image** and a **Container**.

## Goal

Understand the basic relationship:

```text
Image
  ↓
docker run
  ↓
Container
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

### 2. Create the First Container

```bash
docker run -d --name nginx-1 nginx
```

Check:

```bash
docker ps
```

### 3. Create the Second Container

```bash
docker run -d --name nginx-2 nginx
```

Check:

```bash
docker ps
```

You should see:

```text
nginx-1
nginx-2
```

Both containers were created from the same `nginx` image.

## Mental Model

```text
             nginx IMAGE
              /       \
             ↓         ↓
         nginx-1    nginx-2
        CONTAINER   CONTAINER
```

One image can be used to create multiple containers.

## Inspect the Containers

```bash
docker inspect nginx-1
```

```bash
docker inspect nginx-2
```

Both containers use the same image.

## Stop One Container

```bash
docker stop nginx-1
```

Check:

```bash
docker ps
```

`nginx-2` continues running.

This demonstrates that:

```text
Image
  ↓
independent containers
```

Stopping one container does not stop the other containers created from the same image.

## Cleanup

Remove both containers:

```bash
docker rm -f nginx-1 nginx-2
```

Check that the image still exists:

```bash
docker images
```

The `nginx` image remains even after its containers are removed.

## Image vs Container

### Image

```text
Image
→ Blueprint / Template
→ Read-only artifact
→ Used to create containers
```

### Container

```text
Container
→ Instance created from an image
→ Has its own runtime state
→ Can be started, stopped, and removed
```

## Key Takeaway

```text
IMAGE
  │
  ├──→ Container A
  ├──→ Container B
  └──→ Container C
```

Remember:

```text
Container ≠ Image
```

An image is the template. A container is a running or stopped instance created from that template.
