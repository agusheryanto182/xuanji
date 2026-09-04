# Docker Volume

This playground demonstrates how Docker volumes provide **persistent data storage** outside the lifecycle of a container.

## Goal

Understand the difference between container storage and persistent volume storage.

```text
Without Volume

Container
   ↓
Container Filesystem
   ↓
docker rm
   ↓
Data is lost
```

With a volume:

```text
Container
   ↓
Volume
   ↓
Data persists
```

## Practice

### 1. Run a Container

```bash
docker run -d   --name volume-test   nginx
```

Enter the container:

```bash
docker exec -it volume-test bash
```

Create a file:

```bash
echo "hello docker volume" > /tmp/test.txt
```

Exit:

```bash
exit
```

Remove the container:

```bash
docker rm -f volume-test
```

Create a new container:

```bash
docker run -d   --name volume-test   nginx
```

Try to read the file:

```bash
docker exec volume-test cat /tmp/test.txt
```

The file is gone because it existed only inside the old container.

## 2. Create a Volume

```bash
docker volume create my-data
```

Check:

```bash
docker volume ls
```

## 3. Run a Container With a Volume

```bash
docker run -d   --name volume-test   -v my-data:/data   nginx
```

Enter the container:

```bash
docker exec -it volume-test bash
```

Create persistent data:

```bash
echo "hello persistent data" > /data/test.txt
```

Exit:

```bash
exit
```

Remove the container:

```bash
docker rm -f volume-test
```

Create another container using the same volume:

```bash
docker run -d   --name volume-test   -v my-data:/data   nginx
```

Read the file:

```bash
docker exec volume-test cat /data/test.txt
```

Expected:

```text
hello persistent data
```

The data still exists.

## Volume Mapping

The syntax is:

```bash
-v VOLUME_NAME:CONTAINER_PATH
```

Example:

```bash
-v my-data:/data
```

Flow:

```text
my-data Volume
      ↓
container:/data
```

## Mental Model

Without a volume:

```text
Container
   │
   └── /data
         ↓
    docker rm
         ↓
       Data Lost
```

With a volume:

```text
Container
   │
   └── /data
         │
         ↓
      Volume
         │
         ↓
    Data Persists
```

The container and volume have separate lifecycles.

```text
Container → temporary
Volume    → persistent
```

## Cleanup

Remove the container:

```bash
docker rm -f volume-test
```

Remove the volume:

```bash
docker volume rm my-data
```

## Key Takeaway

```text
-v volume-name:/container/path
```

Example:

```bash
-v my-data:/data
```

A Docker volume allows data to survive even when the container that uses it is removed.
