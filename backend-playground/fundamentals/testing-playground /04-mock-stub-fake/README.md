# Mock / Stub / Fake

This playground demonstrates the basic differences between **Stub**, **Mock**, and **Fake** when testing a service with dependencies.

## Structure

```text
04-mock-stub-fake/
├── README.md
├── user.go
└── user_test.go
```

## Dependency Flow

### Production

```text
UserService
     ↓
UserRepository
     ↓
Database
```

### Testing

```text
UserService
     ↓
Test Double
     ↓
No real Database
```

---

## Stub

A **Stub** provides predefined data or behavior to the code under test.

**Focus:** Return value.

```text
UserService
     ↓
Stub
     ↓
Return predefined data
```

---

## Mock

A **Mock** is used to verify how a dependency was interacted with.

**Focus:** Interaction.

```text
UserService
     ↓
Mock
     ↓
Record interaction
     ↓
Assert interaction
```

---

## Fake

A **Fake** is a simplified but working implementation of a dependency.

**Focus:** Simple working implementation.

```text
UserService
     ↓
Fake Repository
     ↓
In-memory data
```

---

## Quick Comparison

| Type | Main Purpose                            |
| ---- | --------------------------------------- |
| Stub | Provide predefined data                 |
| Mock | Verify interaction                      |
| Fake | Provide a simple working implementation |

### Mental Model

```text
Stub
→ "Kasih gue data."

Mock
→ "Gue mau cek lo manggil dependency dengan benar."

Fake
→ "Gue kasih implementasi sederhana yang beneran jalan."
```

---

## Test Double

Stub, Mock, and Fake are types of **Test Double**.

```text
Test Double
├── Stub
├── Mock
└── Fake
```

A Test Double replaces the real dependency during testing.

### Production

```text
UserService
     ↓
PostgreSQL Repository
```

### Test

```text
UserService
     ↓
Stub / Mock / Fake
```

---

## Why?

Without a Test Double:

```text
Unit Test
    ↓
UserService
    ↓
PostgreSQL
```

With a Test Double:

```text
Unit Test
    ↓
UserService
    ↓
Test Double
```

This allows the unit test to run without requiring a real database or external dependency.

---

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
Stub → return data
Mock → verify interaction
Fake → simple working implementation
```

Choose the simplest Test Double that fits the behavior you need to test.
