# Race Condition

This playground demonstrates a **data race** caused by multiple goroutines accessing shared data concurrently without synchronization.

## Goal

Create a race condition and detect it using Go's race detector.

```text
Goroutine 1 ──┐
Goroutine 2 ──┤
Goroutine 3 ──┼──→ shared counter
Goroutine 4 ──┤
Goroutine 5 ──┘
```

## Structure

```text
02-race-condition/
├── README.md
├── go.mod
└── main.go
```

## Setup

This playground is isolated and has its own Go module.

```bash
go mod init race-condition-playground
```

## Practice

Create `main.go`:

```go
package main

import (
    "fmt"
    "sync"
)

func main() {
    var counter int
    var wg sync.WaitGroup

    for i := 0; i < 1000; i++ {
        wg.Add(1)

        go func() {
            defer wg.Done()
            counter++
        }()
    }

    wg.Wait()

    fmt.Println("counter:", counter)
}
```

## Run

```bash
go run .
```

The result may sometimes be:

```text
counter: 1000
```

Do not assume this means the code is safe.

The problem is concurrent access to the same variable:

```go
counter++
```

## Detect the Race

Run:

```bash
go run -race .
```

You should see a warning similar to:

```text
WARNING: DATA RACE
```

The race detector reports concurrent accesses to shared memory that are not properly synchronized.

You can also use it with tests:

```bash
go test -race ./...
```

## Why Does `counter++` Race?

Conceptually, this:

```go
counter++
```

involves multiple steps:

```text
READ counter
     ↓
ADD 1
     ↓
WRITE counter
```

Two goroutines can interleave these operations.

Example:

```text
counter = 10

Goroutine A → read 10
Goroutine B → read 10

Goroutine A → write 11
Goroutine B → write 11

Expected: 12
Actual:   11
```

The exact result is not guaranteed.

## The Problem

Many goroutines access the same variable:

```text
Goroutine 1 ──┐
Goroutine 2 ──┤
Goroutine 3 ──┼──→ counter
Goroutine 4 ──┤
Goroutine 5 ──┘
```

There is no synchronization protecting the shared state.

## Important

Do not fix this playground yet.

The purpose of this exercise is to:

1. Create a race condition.
2. Run the program.
3. Use `-race`.
4. Observe the `DATA RACE` warning.
5. Understand why concurrent access is unsafe.

The next exercise will fix this problem using a mutex.

## Key Takeaway

```text
Concurrent goroutines
        +
Shared mutable data
        +
No synchronization
        ↓
Data Race
```

The important command:

```bash
go run -race .
```

The race detector is a development tool for finding unsafe concurrent memory access.

Next:

```text
Race Condition
      ↓
Mutex
```
