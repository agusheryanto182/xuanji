# Mutex

This playground demonstrates how `sync.Mutex` protects shared mutable data from a race condition.

## Goal

Fix the race condition from the previous exercise.

```text
Without Mutex:

Goroutine A ──→ counter ←── Goroutine B
                    ↑
               unsafe access
```

With a mutex:

```text
Goroutine
    ↓
  Lock
    ↓
counter++
    ↓
 Unlock
```

## Structure

```text
03-mutex/
├── README.md
├── go.mod
└── main.go
```

## Setup

This playground is isolated and has its own Go module.

```bash
go mod init mutex-playground
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
    var mu sync.Mutex
    var wg sync.WaitGroup

    for i := 0; i < 1000; i++ {
        wg.Add(1)

        go func() {
            defer wg.Done()

            mu.Lock()
            counter++
            mu.Unlock()
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

Expected:

```text
counter: 1000
```

Run with the race detector:

```bash
go run -race .
```

There should be no:

```text
WARNING: DATA RACE
```

## What Does the Mutex Do?

The critical section is:

```go
mu.Lock()
counter++
mu.Unlock()
```

`Lock()` prevents another goroutine from entering the protected section at the same time.

```text
Goroutine A
    ↓
  Lock 🔒
    ↓
counter++
    ↓
 Unlock 🔓
```

If another goroutine tries to lock while A owns the mutex:

```text
Goroutine B
    ↓
  Lock 🔒
    ↓
   WAIT
```

After A unlocks:

```text
A → Unlock
      ↓
B → Lock
      ↓
B → counter++
```

This protects the shared variable.

## Why the Race Disappears

Before:

```text
Goroutine A ──┐
Goroutine B ──┼──→ counter
Goroutine C ──┘
```

All goroutines could modify `counter` at the same time.

With a mutex:

```text
Goroutine A
    ↓
   Lock
    ↓
 counter++
    ↓
  Unlock
```

Only one goroutine can modify the protected state at a time.

## Defer Unlock

A common pattern is:

```go
mu.Lock()
defer mu.Unlock()

counter++
```

This ensures the mutex is unlocked when the surrounding function returns.

For this exercise, explicit `Lock()` / `Unlock()` is used first because it makes the synchronization flow easier to see.

## Practice: Reproduce the Race

Temporarily remove:

```go
mu.Lock()
mu.Unlock()
```

and leave:

```go
counter++
```

Then run:

```bash
go run -race .
```

The race detector should report a data race again.

Restore the mutex and run:

```bash
go run -race .
```

The warning should disappear.

## Key Takeaway

```text
Shared mutable data
        +
Multiple goroutines
        +
No synchronization
        ↓
Data Race
```

A mutex protects the critical section:

```text
Lock
 ↓
Critical Section
 ↓
Unlock
```

Use `sync.Mutex` when multiple goroutines need safe access to shared mutable state.

Next:

```text
Mutex
  ↓
Channel
```
