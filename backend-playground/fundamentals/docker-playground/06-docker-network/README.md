# Docker Network

This playground demonstrates how Docker containers communicate with each other through a Docker network.

## Goal

Understand the basic flow:

```text
Container A
    ↓
Docker Network
    ↓
Container B
```

## Practice

### 1. Create a Network

```bash
docker network create app-network
```

Check:

```bash
docker network ls
```

### 2. Run the First Container

```bash
docker run -d   --name nginx-server   --network app-network   nginx
```

### 3. Run the Second Container

```bash
docker run -d   --name client   --network app-network   alpine   sleep 3600
```

The network now contains:

```text
app-network
     │
     ├── nginx-server
     │
     └── client
```

### 4. Test Container-to-Container Communication

Enter the client container:

```bash
docker exec -it client sh
```

Install curl:

```bash
apk add --no-cache curl
```

Call the Nginx container:

```bash
curl http://nginx-server
```

You should receive a response from Nginx.

Notice that we use:

```text
nginx-server
```

instead of an IP address.

## Docker Internal DNS

Containers connected to the same Docker network can communicate using container names.

```text
Container Name
      ↓
Docker DNS
      ↓
Container IP
```

Example:

```text
client
  │
  │ http://nginx-server
  ↓
Docker DNS
  ↓
nginx-server
  ↓
Nginx :80
```

## Important: `localhost`

Inside a container:

```text
localhost
```

means **that container itself**.

It does not mean another container.

For example, from `client`:

```text
localhost:80
```

means:

```text
client container :80
```

While:

```text
nginx-server:80
```

means:

```text
nginx-server container :80
```

This distinction becomes important when running application, PostgreSQL, and Redis containers together.

## Cleanup

Exit the client container:

```bash
exit
```

Remove the containers:

```bash
docker rm -f nginx-server client
```

Remove the network:

```bash
docker network rm app-network
```

## Key Takeaway

Create a network:

```bash
docker network create app-network
```

Connect containers to it:

```bash
docker run --network app-network ...
```

Then containers can communicate using their container/service names.

```text
Container
    ↓
Docker Network
    ↓
Container Name
    ↓
Container-to-Container Communication
```
