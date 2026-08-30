# Basic Unit Test

This playground demonstrates the fundamentals of writing and running a **Unit Test** in Go.

## Structure

```text
01-basic-unit-test/
├── README.md
├── calculator.go
└── calculator_test.go
```

## What is a Unit Test?

A Unit Test tests a small, isolated piece of code, usually a function or method.

Example:

```go
func Add(a, b int) int {
    return a + b
}
```

We test whether:

```text
Add(2, 3)
    ↓
    5
```

---

## Basic Test

```go
func TestAdd(t *testing.T) {
    got := Add(2, 3)

    if got != 5 {
        t.Errorf("got %d, want %d", got, 5)
    }
}
```

### Flow

```text
Call Function
     ↓
Get Actual Result
     ↓
Compare with Expected Result
     ↓
PASS / FAIL
```

---

## Test Naming

Go recognizes test functions using the `TestXxx` convention.

```go
func TestAdd(t *testing.T) {
    // test
}
```

The test file must end with:

```text
_test.go
```

Example:

```text
calculator_test.go
```

---

## PASS

If the actual result matches the expected result:

```text
Actual   = 5
Expected = 5
    ↓
PASS
```

Example output:

```text
=== RUN   TestAdd
--- PASS: TestAdd (0.00s)
PASS
```

---

## FAIL

If the actual result does not match the expected result:

```text
Actual   = 5
Expected = 6
    ↓
FAIL
```

Example:

```text
=== RUN   TestAdd
    calculator_test.go:...: got 5, want 6
--- FAIL: TestAdd
FAIL
```

---

## `testing` Package

Go provides the standard `testing` package:

```go
import "testing"
```

Common methods:

```go
t.Errorf(...)
t.Fatalf(...)
t.Error(...)
t.Fatal(...)
```

For basic tests, the important idea is:

```text
Actual Result
     ↓
Expected Result
     ↓
Assertion
```

---

## Test Pattern

A simple unit test commonly follows:

```text
Arrange
  ↓
Act
  ↓
Assert
```

### Arrange

Prepare the input:

```go
a := 2
b := 3
```

### Act

Run the function:

```go
got := Add(a, b)
```

### Assert

Check the result:

```go
if got != 5 {
    t.Errorf("got %d, want %d", got, 5)
}
```

---

## Why Unit Tests?

Unit tests help verify that individual pieces of logic behave correctly.

They should generally be:

```text
Unit Test
├── Small
├── Fast
├── Isolated
└── Repeatable
```

A unit test should normally avoid unnecessary dependencies such as:

```text
PostgreSQL
Redis
HTTP APIs
External Services
```

Those dependencies are covered more appropriately by integration tests.

---

## Run Tests

From the project root:

```bash
go test ./...
```

Run tests verbosely:

```bash
go test -v ./...
```

Run a specific test:

```bash
go test -run TestAdd
```

---

## Key Takeaway

```text
Unit Test
=
Test a small piece of logic
+
Isolated
+
Expected result
+
PASS / FAIL
```

Basic pattern:

```go
func TestXxx(t *testing.T) {
    // Arrange

    // Act

    // Assert
}
```
