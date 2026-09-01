# Goroutine

This playground demonstrates the basic use of goroutines in Go.

## Goal

Understand the difference between sequential execution and concurrent execution.

```text
Without goroutine:

say hello
   ↓
finish
   ↓
say world
```

With goroutines:

```text
        ┌── say hello
main ───┤
        └── say world
```

## Setup

This playground is isolated and has its own Go module.

```bash
mkdir -p languages/go/01-concurrency
cd languages/go/01-concurrency

go mod init concurrency-playground
```

## Practice

Create `main.go`:

```go
package main

import (
    "fmt"
    "time"
)

func say(message string) {
    for i := 0; i < 3; i++ {
        fmt.Println(message, i)
        time.Sleep(100 * time.Millisecond)
    }
}

func main() {
    go say("hello")
    go say("world")

    time.Sleep(time.Second)
}
```

Run:

```bash
go run .
```

The output order can vary:

```text
hello 0
world 0
world 1
hello 1
hello 2
world 2
```

Another run may produce a different order.

## Sequential Version

Remove `go`:

```go
say("hello")
say("world")
```

Run:

```bash
go run .
```

The output will be sequential:

```text
hello 0
hello 1
hello 2
world 0
world 1
world 2
```

## Why Does the Order Change?

With:

```go
go say("hello")
go say("world")
```

both functions run as goroutines.

The Go runtime scheduler determines when each goroutine gets execution time, so the exact output order is not guaranteed.

## Important

This playground uses:

```go
time.Sleep(time.Second)
```

only to keep the main function alive long enough to observe the goroutines.

Do not use `time.Sleep` as a synchronization mechanism in production code.

Later, synchronization tools such as `WaitGroup`, channels, and context will be used for proper goroutine lifecycle management.

## Key Takeaway

A goroutine is started with the `go` keyword:

```go
go say("hello")
```

The important distinction is:

```text
Sequential
    ↓
function A finishes
    ↓
function B starts

Concurrent
    ↓
function A and B can make progress independently
```

This is the foundation for the next topics:

```text
Goroutine
   ↓
Race Condition
   ↓
Synchronization
   ↓
Mutex / Channel
```