# Select

This playground demonstrates Go's `select` statement for waiting on multiple channel operations.

## Goal

Understand how one goroutine can wait for multiple channel operations and continue when one becomes ready.

```text
              select
             /      \
            ↓        ↓
          ch1       ch2
            │        │
            ↓        ↓
         ready?    ready?
            │        │
            └───┬────┘
                ↓
           first ready
```

## Structure

```text
05-select/
├── README.md
├── go.mod
└── main.go
```

## Setup

This playground is isolated and has its own Go module.

```bash
go mod init select-playground
```

## Practice

Create `main.go`:

```go
package main

import (
    "fmt"
    "time"
)

func main() {
    ch1 := make(chan string)
    ch2 := make(chan string)

    go func() {
        time.Sleep(1 * time.Second)
        ch1 <- "message from ch1"
    }()

    go func() {
        time.Sleep(2 * time.Second)
        ch2 <- "message from ch2"
    }()

    select {
    case message := <-ch1:
        fmt.Println(message)

    case message := <-ch2:
        fmt.Println(message)
    }
}
```

Run:

```bash
go run .
```

Expected:

```text
message from ch1
```

## How `select` Works

Without `select`:

```go
message := <-ch1
```

the goroutine waits specifically for `ch1`.

With:

```go
select {
case message := <-ch1:
    fmt.Println(message)

case message := <-ch2:
    fmt.Println(message)
}
```

the goroutine waits for either `ch1` or `ch2`.

Whichever operation becomes ready first is selected.

## Multiple Ready Cases

If multiple cases are ready at the same time, `select` chooses one of the ready cases pseudo-randomly.

Do not rely on case order to create priority.

## Select With Timeout

A common pattern is:

```go
select {
case message := <-ch:
    fmt.Println(message)

case <-time.After(1 * time.Second):
    fmt.Println("timeout")
}
```

The goroutine waits for either a channel value or the timeout.

## Important: Select Does Not Wait for All Goroutines

When one case is selected, the `select` finishes.

If `main()` then returns:

```text
main returns
    ↓
program exits
    ↓
remaining goroutines stop
```

`select` is not equivalent to `sync.WaitGroup`.

Use `WaitGroup` when the goal is to wait for a group of goroutines to finish.

## `select` vs `WaitGroup`

`select`:

```text
Wait for one of multiple channel operations
```

`WaitGroup`:

```text
Wait for a group of goroutines to finish
```

They solve different problems.

## Key Takeaway

`select` allows a goroutine to wait on multiple channel operations:

```go
select {
case value := <-ch1:
    // ch1 ready

case value := <-ch2:
    // ch2 ready
}
```

It is especially useful for:

```text
multiple channels
timeouts
cancellation
non-blocking channel operations
```

Next:

```text
Channel + Goroutines
        ↓
Worker Pool
```
