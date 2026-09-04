# Integration Testing

This playground demonstrates how to test multiple application components working together with a **real PostgreSQL database**.

## Structure

```text
08-integration-testing/
├── README.md
├── docker-compose.yml
├── migrations/
│   └── 001_create_users.sql
├── handler.go
├── service.go
├── repository.go
└── integration_test.go
```

## Environment

This playground has its own PostgreSQL environment.

```text
Host     : localhost
Port     : 5434
Database : integration_testing
User     : postgres
Password : postgres
```

Start PostgreSQL:

```bash
docker compose up -d
```

Stop PostgreSQL:

```bash
docker compose down
```

Reset the database:

```bash
docker compose down -v
docker compose up -d
```

## What is Integration Testing?

Integration testing verifies that multiple components work correctly together.

```text
Handler
   ↓
Service
   ↓
Repository
   ↓
PostgreSQL
```

Unlike a unit test, the dependencies are not replaced with mocks or fakes.

## Test Flow

The main flow in this playground is:

```text
POST /users
    ↓
HTTP Handler
    ↓
User Service
    ↓
User Repository
    ↓
PostgreSQL
    ↓
201 Created
```

The test sends a real HTTP request to the handler and uses the real service, repository, and PostgreSQL database.

## Unit vs Repository vs Integration

### Unit Test

```text
Service
   ↓
Fake / Mock Repository
   ↓
No Database
```

Focus:

```text
Business Logic
```

### Repository Test

```text
Repository
   ↓
Real PostgreSQL
```

Focus:

```text
SQL + Database Interaction
```

### Integration Test

```text
HTTP
 ↓
Handler
 ↓
Service
 ↓
Repository
 ↓
PostgreSQL
```

Focus:

```text
Components Working Together
```

## What This Test Verifies

```text
HTTP Request
     ↓
Request Parsing
     ↓
Handler
     ↓
Service
     ↓
Repository
     ↓
SQL INSERT
     ↓
PostgreSQL
     ↓
HTTP 201 Response
```

This catches integration problems that isolated unit tests may not detect.

## Database Isolation

The PostgreSQL instance is dedicated to this playground.

It does not depend on:

```text
- 07-repository-testing
- Other playgrounds
- Production
- Staging
```

The database uses port `5434` so it does not conflict with the PostgreSQL environment in the repository-testing playground.

## Run Tests

From the project root:

```bash
go test ./08-integration-testing -v
```

Or run all playground tests:

```bash
go test ./...
```

## Key Takeaway

```text
Integration Test
=
Multiple real components
+
Real dependency
+
Verify they work together
```

The goal is not to test every detail. Focus on important flows that cross application boundaries.
