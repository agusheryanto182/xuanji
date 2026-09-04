# Repository Testing

This playground demonstrates how to test a Go repository against a **real PostgreSQL database**.

## Structure

```text
07-repository-testing/
├── README.md
├── docker-compose.yml
├── migrations/
│   └── 001_create_users.sql
├── repository.go
└── repository_test.go
```

## Environment

This playground is self-contained and uses its own PostgreSQL container.

```text
Host     : localhost
Port     : 5433
Database : testing
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

## Repository Flow

```text
Test
  ↓
Repository
  ↓
database/sql
  ↓
PostgreSQL
```

Unlike a unit test, the repository test uses a **real PostgreSQL database**.

```text
Unit Test
    ↓
Fake / Mock
    ↓
No Database

Repository Test
    ↓
Real Repository
    ↓
Real PostgreSQL
```

## What We Test

The playground focuses on the most important repository operations.

### Create

```text
Create User
    ↓
INSERT
    ↓
PostgreSQL
    ↓
Return ID
```

### Get

```text
GetByID
    ↓
SELECT
    ↓
PostgreSQL
    ↓
Return User
```

### Update

```text
Update User
    ↓
UPDATE
    ↓
PostgreSQL
    ↓
GetByID
    ↓
Verify Updated Data
```

## Test Flow

```text
Create
  ↓
GetByID
  ↓
Modify User
  ↓
Update
  ↓
GetByID
  ↓
Verify
```

## Why Repository Testing?

Repository tests verify that the actual database interaction works correctly.

They can catch problems such as:

```text
- Incorrect SQL
- Incorrect column names
- Incorrect parameter order
- Scan errors
- Database constraints
- PostgreSQL-specific behavior
```

A mock repository cannot detect these database-specific problems.

## Test Database Isolation

The PostgreSQL instance is dedicated to this playground.

It does not depend on:

```text
- PostgreSQL playground
- Other testing playgrounds
- Production database
- Staging database
```

The database is exposed on port `5433` to avoid conflicts with PostgreSQL instances using the default `5432`.

## Run Tests

From the project root:

```bash
go test ./07-repository-testing -v
```

Or run all playground tests:

```bash
go test ./...
```

## Key Takeaway

```text
Repository Test
=
Real Repository
+
Real PostgreSQL
+
Real SQL
```

Use repository/integration tests when you need to verify that your application actually works with the database.
