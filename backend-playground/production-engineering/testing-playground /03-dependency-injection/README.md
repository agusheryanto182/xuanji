# Dependency Injection

This playground demonstrates **Dependency Injection (DI)** in Go and how it helps make code easier to test.

## Structure

```text
03-dependency-injection/
├── README.md
├── user.go
└── user_test.go
```

## What is Dependency Injection?

Dependency Injection means:

> A component receives its dependencies from the outside instead of creating them internally.

Without Dependency Injection:

```text
UserService
     ↓
creates Repository itself
     ↓
PostgreSQL
```

With Dependency Injection:

```text
Repository
     ↓
   inject
     ↓
UserService
```

The `UserService` does not need to know how the repository is created.

---

## Example

Define the dependency as an interface:

```go
type UserRepository interface {
    GetByID(id int) (User, error)
}
```

The service receives the dependency:

```go
type UserService struct {
    repo UserRepository
}
```

Inject it through the constructor:

```go
func NewUserService(repo UserRepository) *UserService {
    return &UserService{
        repo: repo,
    }
}
```

Usage:

```go
repo := NewUserRepository(db)

service := NewUserService(repo)
```

---

## Why Dependency Injection?

It makes dependencies replaceable.

### Production

```text
UserService
     ↓
PostgreSQL Repository
     ↓
PostgreSQL
```

### Unit Test

```text
UserService
     ↓
Fake / Mock Repository
     ↓
No PostgreSQL
```

This allows us to test `UserService` without requiring a real database.

---

## Dependency Flow

```text
                    ┌────────────────────┐
                    │    UserService     │
                    └─────────┬──────────┘
                              │
                              ▼
                    UserRepository
                       (interface)
                              │
                 ┌────────────┴────────────┐
                 ▼                         ▼
        PostgreSQL Repository       Fake Repository
             Production                 Testing
```

---

## Key Takeaway

```text
Dependency Injection
=
Give dependencies from outside
instead of creating them inside.
```

Main benefits:

```text
DI
├── Loose Coupling
├── Easier Testing
├── Replaceable Dependencies
└── Cleaner Architecture
```

For testing:

```text
Real Dependency
      ↓
replace with
      ↓
Fake / Mock / Stub
```

---

## Run Tests

From the project root:

```bash
go test ./...
```

Verbose:

```bash
go test -v ./...
```
