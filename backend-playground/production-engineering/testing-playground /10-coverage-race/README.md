# Coverage + Race Detector

This playground demonstrates two important Go testing tools:

- **Coverage** — checks which parts of the code are executed by tests.
- **Race Detector** — detects data races caused by unsafe concurrent access.

## Coverage

Run:

```bash
go test ./... -cover
```

Example:

```text
ok  github.com/.../05-usecase-testing  80.0% coverage
```

Coverage answers:

> "How much of the code was executed by the tests?"

### Coverage Profile

Generate a coverage profile:

```bash
go test ./... -coverprofile=coverage.out
```

View coverage by function:

```bash
go tool cover -func=coverage.out
```

Open the HTML coverage report:

```bash
go tool cover -html=coverage.out
```

Flow:

```text
Code
 ↓
go test
 ↓
Coverage Profile
 ↓
See which code was executed
```

## Important

High coverage does **not** mean the code has no bugs.

```text
100% Coverage
    ≠
100% Bug Free
```

Coverage tells us whether code was executed, not whether the test assertions are good.

---

## Race Detector

Go has a built-in race detector.

Run:

```bash
go test -race ./...
```

The race detector looks for unsafe concurrent access to shared memory.

Example:

```go
func TestRace(t *testing.T) {
    var counter int

    done := make(chan bool)

    go func() {
        counter++
        done <- true
    }()

    go func() {
        counter++
        done <- true
    }()

    <-done
    <-done
}
```

Two goroutines modify the same variable:

```text
Goroutine A
     ↓
   counter
     ↑
Goroutine B
```

There is no synchronization between the accesses.

Running:

```bash
go test -race
```

allows Go to detect this race.

---

## Key Difference

```text
Coverage
→ Was this code executed by a test?

Race Detector
→ Is concurrent access to shared data unsafe?
```

## Important Commands

### Coverage

```bash
go test ./... -cover
```

### Coverage Profile

```bash
go test ./... -coverprofile=coverage.out
```

### Coverage Details

```bash
go tool cover -func=coverage.out
```

### Coverage HTML

```bash
go tool cover -html=coverage.out
```

### Race Detector

```bash
go test -race ./...
```

### Both

```bash
go test -race ./... -cover
```

## Key Takeaway

```text
Coverage
=
Measure tested code execution

Race Detector
=
Detect unsafe concurrent access
```

Use coverage to find areas that may need tests, and use the race detector when working with concurrent Go code.
