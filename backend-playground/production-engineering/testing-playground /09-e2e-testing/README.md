# E2E Testing

This playground demonstrates **End-to-End (E2E) testing** for a Go HTTP API.

The test behaves like an external client and verifies a complete application flow through the real HTTP API and PostgreSQL database.

## Structure

```text
09-e2e-testing/
├── README.md
├── docker-compose.yml
├── migrations/
│   └── 001_create_users.sql
├── server.go
└── e2e_test.go
```

## Environment

This playground has its own PostgreSQL environment.

```text
Host     : localhost
Port     : 5435
Database : e2e_testing
User     : postgres
Password : postgres
```

Start PostgreSQL:

```bash
docker compose up -d
```

Check the container:

```bash
docker compose ps
```

The PostgreSQL port should be mapped as:

```text
0.0.0.0:5435->5432/tcp
```

## What is E2E Testing?

E2E testing verifies a complete user-facing flow from outside the application.

```text
E2E Test
    ↓
HTTP Client
    ↓
Application
    ↓
Handler
    ↓
Service
    ↓
Repository
    ↓
PostgreSQL
```

The test communicates with the application through its external HTTP interface.

## E2E Flow

The main flow is:

```text
POST /users
    ↓
Create User
    ↓
201 Created
    ↓
Get returned ID
    ↓
GET /users/:id
    ↓
200 OK
    ↓
Verify User
```

## Application Lifecycle

The E2E test should be **self-contained**.

You do not need to manually run:

```bash
go run server.go
```

The test is responsible for starting the application and using it through HTTP.

The intended flow is:

```text
docker compose up -d
        ↓
go test -v
        ↓
Start Application
        ↓
HTTP Client
        ↓
Running API
        ↓
PostgreSQL
        ↓
Test Complete
        ↓
Application Stops
```

This makes the test easier to reproduce locally and in CI/CD.

## E2E vs Integration Test

### Integration Test

```text
Test
 ↓
Handler
 ↓
Service
 ↓
Repository
 ↓
PostgreSQL
```

The test may use `httptest` and directly invoke the application handler.

### E2E Test

```text
Test
 ↓
HTTP Client
 ↓
Running Application
 ↓
Handler
 ↓
Service
 ↓
Repository
 ↓
PostgreSQL
```

The test communicates with the application through its external HTTP interface.

## Why E2E Testing?

E2E tests verify that an important user-facing flow works from beginning to end.

They can catch problems such as:

```text
- Routing problems
- Request/response problems
- Handler problems
- Service integration problems
- Repository problems
- Database integration problems
```

## Keep E2E Tests Focused

E2E tests should focus on important critical flows.

```text
Critical User Flow
        ↓
      E2E Test
```

Detailed business rules should generally be covered by unit and integration tests.

## Database Isolation

The PostgreSQL instance is dedicated to this playground.

It does not depend on:

```text
- 07-repository-testing
- 08-integration-testing
- Other playgrounds
- Production
- Staging
```

Port `5435` is used to avoid conflicts with other PostgreSQL playgrounds.

## Run Tests

Start PostgreSQL:

```bash
docker compose up -d
```

Then run the E2E test:

```bash
go test -v
```

Or from the project root:

```bash
go test ./09-e2e-testing -v
```

You should **not** need to manually start the application.

## Key Takeaway

```text
E2E Test
=
External Client
+
Running Application
+
Real Dependencies
+
Complete User Flow
```

For this playground:

```text
HTTP Client
    ↓
API
    ↓
PostgreSQL
```

The goal is to verify that the application works correctly from the perspective of a real client.
