# Channel

This playground demonstrates how Go channels allow goroutines to communicate by sending and receiving values.

## Goal

```text
Goroutine
    ↓
  channel
    ↓
Goroutine / main
```

## Structure

```text
04-channel/
├── README.md
├── go.mod
└── main.go
```

## Setup

This playground is isolated and has its own Go module.

```bash
go mod init channel-playground
```

## Practice

Create `main.go`:

```go
package main

import "fmt"

func main() {
    ch := make(chan string)

    go func() {
        ch <- "hello from goroutine"
    }()

    message := <-ch

    fmt.Println(message)
}
```

Run:

```bash
go run .
```

Expected:

```text
hello from goroutine
```

## Send and Receive

Create a channel:

```go
ch := make(chan string)
```

Send a value:

```go
ch <- "hello"
```

Receive a value:

```go
message := <-ch
```

```text
ch <- value
    ↓
send

value := <-ch
    ↓
receive
```

## Communication Flow

```text
Goroutine
    │
    │ ch <- "hello"
    ↓
┌─────────┐
│ channel │
└─────────┘
    │
    │ <-ch
    ↓
  main()
```

## Integer Example

```go
package main

import "fmt"

func main() {
    ch := make(chan int)

    go func() {
        ch <- 42
    }()

    number := <-ch

    fmt.Println(number)
}
```

Expected:

```text
42
```

## Unbuffered Channel

This:

```go
ch := make(chan string)
```

creates an unbuffered channel.

A send blocks until another goroutine is ready to receive.

```text
Sender
  │
  │ ch <- value
  ↓
 waits
  │
  ↓
Receiver
  │
  │ <-ch
  ↓
continues
```

This synchronization behavior is an important property of channels.

## Deadlock Example

This code can deadlock:

```go
ch := make(chan string)

ch <- "hello"

fmt.Println(<-ch)
```

There is no separate receiver ready when the send occurs.

The runtime may report:

```text
fatal error: all goroutines are asleep - deadlock!
```

## Mutex vs Channel

Mutex:

```text
Shared State
    ↓
  Mutex
    ↓
Protect Access
```

Channel:

```text
Goroutine A
    ↓
  Channel
    ↓
Goroutine B
```

Use a mutex when the main problem is protecting shared mutable state.

Use a channel when goroutines need to communicate or transfer data.

## Key Takeaway

A channel provides communication between goroutines.

```text
Send:
ch <- value

Receive:
value := <-ch
```

Basic pattern:

```text
Goroutine
    ↓
  Channel
    ↓
Goroutine
```

Next:

```text
Channel
  ↓
Buffered Channel / Select
```
