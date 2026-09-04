# Usecase / Service Testing

This playground demonstrates how to test business logic in a service/usecase layer without using a real database.

## Structure

```text
05-usecase-testing/
├── README.md
├── user.go
└── user_test.go
```

## Flow

```text
Handler
   ↓
UserService
   ↓
UserRepository
   ↓
Database
```

During unit testing:

```text
UserService
   ↓
Fake Repository
   ↓
In-memory behavior
```

## Business Rule

A user cannot be created if the email is already registered.

```text
CreateUser()
     ↓
FindByEmail()
     ↓
┌──────────────────────┐
│ Email already exists │
└──────────┬───────────┘
           ↓
         Error

If email does not exist:
     ↓
Create()
     ↓
Success
```

## Test Cases

### Email is not registered

```text
FindByEmail()
     ↓
User not found
     ↓
Create()
     ↓
Success
```

### Email is already registered

```text
FindByEmail()
     ↓
User found
     ↓
Return error
     ↓
Create() is NOT called
```

## What This Tests

The test focuses on **business logic**, not PostgreSQL.

```text
Input
  ↓
Business Rule
  ↓
Repository Dependency
  ↓
Expected Behavior
```

The repository is replaced with a fake implementation.

## Run Tests

```bash
go test ./...
```

Verbose:

```bash
go test -v ./...
```

## Key Takeaway

```text
Usecase / Service Test
=
Test business logic
+
Replace external dependencies
+
Verify expected behavior
```

The database itself should be tested separately through repository/integration tests.
