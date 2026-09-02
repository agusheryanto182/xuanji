# 02 - Context

## What is Context?

In Go, `context.Context` is used to carry cancellation signals, deadlines, and request-scoped values across API boundaries and between goroutines.

The most important use case to understand first is **cancellation**.

> Context gives a running operation a way to know that it should stop.

This is especially useful for:

- HTTP requests
- database queries
- external API calls
- background goroutines
- worker pools
- timeouts and deadlines

---

## Why Do We Need Context?

Imagine an HTTP request starts several operations:

```text
HTTP Request
     │
     ▼
  Handler
     │
     ├── DB Query
     ├── External API Call
     └── Goroutine
```

Then the client cancels the request.

Without cancellation:

```text
request cancelled
       │
       ├── DB query still running
       ├── API call still running
       └── goroutine still running
```

Those operations may continue even though the original request is no longer needed.

Context provides a cancellation signal:

```text
cancel
  │
  ▼
context.Done()
  │
  ├── DB query can stop
  ├── API call can stop
  └── goroutine can stop
```

This helps prevent wasted work and goroutine leaks.

---

# `context.Background()`

The simplest context is:

```go
ctx := context.Background()
```

`context.Background()` returns an empty root context.

It does not:

- have a deadline
- have a cancellation signal
- contain values

It is commonly used as the starting point for creating other contexts.

Example:

```go
package main

import (
	"context"
	"fmt"
)

func main() {
	ctx := context.Background()

	fmt.Println(ctx)
}
```

Think of it as:

```text
context.Background()
        │
        ▼
   root context
        │
        ├── WithCancel
        ├── WithTimeout
        └── WithDeadline
```

---

# `context.WithCancel()`

`WithCancel()` creates a child context that can be cancelled manually.

```go
ctx, cancel := context.WithCancel(context.Background())
```

It returns two things:

```text
ctx
 │
 └── context

cancel
 │
 └── function used to send cancellation
```

Example:

```go
ctx, cancel := context.WithCancel(context.Background())

go worker(ctx)

time.Sleep(2 * time.Second)

cancel()
```

Calling:

```go
cancel()
```

signals every operation using that context that it should stop.

---

# `ctx.Done()`

The cancellation signal can be observed through:

```go
ctx.Done()
```

`Done()` returns a channel.

A goroutine can wait for cancellation using:

```go
<-ctx.Done()
```

Example:

```go
func worker(ctx context.Context) {
	<-ctx.Done()

	fmt.Println("worker stopped")
}
```

Flow:

```text
worker
  │
  ▼
<-ctx.Done()
  │
  │ waiting
  │
cancel()
  │
  ▼
ctx.Done()
  │
  ▼
worker continues
  │
  ▼
return
```

This is closely related to the goroutine leak topic.

---

# Context + `select`

A very common pattern is:

```go
func worker(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			fmt.Println("worker stopped")
			return

		default:
			fmt.Println("worker working...")
			time.Sleep(500 * time.Millisecond)
		}
	}
}
```

The worker has two possible paths:

```text
              worker
                 │
                 ▼
              select
             /      \
            /        \
     cancelled       not cancelled
        │                  │
        ▼                  ▼
      return             work
                           │
                           └── loop
```

When `cancel()` is called:

```go
cancel()
```

the worker receives the cancellation signal:

```text
cancel()
   ↓
ctx.Done()
   ↓
select
   ↓
worker returns
```

This gives the goroutine a clear lifecycle.

---

# Complete Example

```go
package main

import (
	"context"
	"fmt"
	"time"
)

func worker(ctx context.Context) {
	fmt.Println("worker started")

	for {
		select {
		case <-ctx.Done():
			fmt.Println("worker stopped")
			return

		default:
			fmt.Println("worker working...")
			time.Sleep(500 * time.Millisecond)
		}
	}
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())

	go worker(ctx)

	time.Sleep(2 * time.Second)

	fmt.Println("cancelling worker...")
	cancel()

	time.Sleep(1 * time.Second)

	fmt.Println("main finished")
}
```

Expected flow:

```text
worker started
worker working...
worker working...
worker working...
worker working...
cancelling worker...
worker stopped
main finished
```

---

# Context Does Not Kill a Goroutine

This is an important concept.

Calling:

```go
cancel()
```

does **not** forcibly kill the goroutine.

Instead, it sends a cancellation signal.

The goroutine must cooperate:

```go
select {
case <-ctx.Done():
	return
}
```

Think of it as:

```text
cancel()
   │
   ▼
"Please stop"
   │
   ▼
goroutine checks signal
   │
   ▼
return
```

Not:

```text
cancel()
   │
   ▼
FORCE KILL GOROUTINE
```

Go does not provide a general mechanism to forcibly kill an arbitrary goroutine.

---

# Context Propagation

Context is usually passed down through function calls.

Example:

```go
func handler(ctx context.Context) {
	service(ctx)
}

func service(ctx context.Context) {
	repository(ctx)
}

func repository(ctx context.Context) {
	// database operation
}
```

The flow becomes:

```text
HTTP Handler
     │
     │ ctx
     ▼
  Service
     │
     │ ctx
     ▼
Repository
     │
     ▼
Database
```

This allows cancellation to propagate through the call chain.

---

# Do Not Create Random Background Contexts

Suppose a function already receives a context:

```go
func service(ctx context.Context) {
	repository(ctx)
}
```

Do not unnecessarily replace it with:

```go
func service(ctx context.Context) {
	newCtx := context.Background()

	repository(newCtx)
}
```

Doing this breaks the cancellation chain.

Prefer:

```go
func service(ctx context.Context) {
	repository(ctx)
}
```

The existing context should normally be propagated downward.

---

# Always Call `cancel()`

When creating a cancellable context:

```go
ctx, cancel := context.WithCancel(context.Background())
```

the caller is generally responsible for calling:

```go
defer cancel()
```

when appropriate.

Example:

```go
func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go worker(ctx)

	// ...
}
```

Calling `cancel()` releases resources associated with the derived context and signals its children.

---

# Context and Goroutine Leak

These two topics are strongly connected.

A goroutine leak can happen when:

```text
goroutine
    ↓
waiting forever
    ↓
no exit condition
```

Context can provide that exit signal:

```text
goroutine
    ↓
select
    ↓
<-ctx.Done()
    ↓
return
    ↓
goroutine finished
```

So one useful mental model is:

```text
Goroutine
    │
    ├── Work
    │
    └── Cancellation
           │
           ▼
        ctx.Done()
           │
           ▼
         return
```

---

# Key Takeaways

Remember these points:

1. `context.Context` helps control the lifecycle of operations.
2. `context.Background()` is commonly used as a root context.
3. `context.WithCancel()` creates a manually cancellable context.
4. `ctx.Done()` provides the cancellation signal.
5. `cancel()` sends the cancellation signal.
6. Context does not forcibly kill goroutines.
7. Goroutines must cooperate with cancellation.
8. Context should normally be propagated through function calls.
9. Cancellation is an important tool for preventing goroutine leaks.
10. Always think about how and when an operation should stop.

## Mental Model

```text
CREATE
  │
  ▼
context.Background()
  │
  ▼
WithCancel()
  │
  ├───────────────► ctx
  │
  └───────────────► cancel()
                       │
                       ▼
                    ctx.Done()
                       │
                       ▼
                  goroutine stops
```

The key question when working with goroutines is:

> **"How will this operation know that it should stop?"**

Context is one of Go's primary answers to that question.
