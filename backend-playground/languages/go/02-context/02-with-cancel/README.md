# 02 - WithCancel

## What is `context.WithCancel()`?

`context.WithCancel()` creates a child context that can be manually cancelled.

```go
ctx, cancel := context.WithCancel(context.Background())
```

It returns two things:

```text
ctx
 │
 └── context that carries the cancellation signal

cancel
 │
 └── function used to send the cancellation signal
```

---

## Basic Example

```go
ctx, cancel := context.WithCancel(context.Background())

go worker(ctx)

time.Sleep(2 * time.Second)

cancel()
```

The important relationship is:

```text
WithCancel()
     │
     ├── ctx
     │
     └── cancel()
```

`ctx` is passed to the operation or goroutine.

`cancel()` is called by the owner when the operation should stop.

---

## Listening for Cancellation

A goroutine can listen for cancellation through:

```go
ctx.Done()
```

Example:

```go
func worker(ctx context.Context) {
	<-ctx.Done()

	fmt.Println("worker stopped")
}
```

The worker waits until the context is cancelled.

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

---

## `WithCancel` + `select`

A common pattern is:

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

The worker can either continue working or stop when cancellation is signaled.

```text
              worker
                 │
                 ▼
              select
             /                  /               cancelled     continue
          │             │
          ▼             ▼
        return         work
                         │
                         └── loop
```

---

## Complete Example

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

Possible output:

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

## Important: `cancel()` Does Not Kill the Goroutine

Calling:

```go
cancel()
```

does **not** forcibly kill the goroutine.

It sends a cancellation signal.

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
   │
   ▼
goroutine finishes
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

## Why Is This Useful?

Consider a long-running worker:

```text
worker
   │
   ▼
working...
   │
   ▼
working...
   │
   ▼
working...
```

At some point, the application knows the worker is no longer needed.

Without cancellation:

```text
worker
   │
   ▼
working / waiting
   │
   ▼
never stops
```

With `WithCancel()`:

```text
worker
   │
   ▼
working...
   │
   │
cancel()
   │
   ▼
ctx.Done()
   │
   ▼
return
   │
   ▼
goroutine finished
```

This gives the goroutine a clear lifecycle.

---

## Connection to Goroutine Leaks

This connects directly to the previous topic.

A goroutine leak can happen when:

```text
goroutine
    ↓
waiting forever
    ↓
no exit condition
```

A context can provide the exit signal:

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

Therefore, cancellation is one of the tools that can help prevent goroutine leaks.

---

## `defer cancel()`

When creating a cancellable context, the code that creates the context is generally responsible for calling the cancellation function.

A common pattern is:

```go
ctx, cancel := context.WithCancel(context.Background())
defer cancel()
```

Example:

```go
func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go worker(ctx)

	// ...
}
```

Calling `cancel()` is safe even if the context has already been cancelled.

---

## Parent and Child

`WithCancel()` creates a child context from a parent:

```go
root := context.Background()

ctx, cancel := context.WithCancel(root)
```

The relationship is:

```text
Background
    │
    ▼
   ctx
    │
    └── cancellation
```

The child inherits cancellation from its parent.

This becomes important when multiple layers of an application share the same context.

---

## Multiple Children

One parent can have multiple children:

```go
root := context.Background()

ctx1, cancel1 := context.WithCancel(root)
ctx2, cancel2 := context.WithCancel(root)
```

Structure:

```text
             Background
              /                    /                    ctx1        ctx2
```

Cancelling `ctx1` does not cancel `ctx2`.

But cancelling the parent:

```text
Background
    │
    ├── ctx1
    └── ctx2
```

can propagate cancellation to both children.

---

## Key Takeaways

1. `context.WithCancel()` creates a cancellable child context.
2. It returns `(ctx, cancel)`.
3. `ctx` is passed to the operation that should be cancellable.
4. `cancel()` sends the cancellation signal.
5. `ctx.Done()` can be used to observe cancellation.
6. `cancel()` does not forcibly kill a goroutine.
7. The goroutine must cooperate and return.
8. `select` is commonly used to handle cancellation while doing work.
9. The creator of the context is generally responsible for calling `cancel()`.
10. Cancellation provides a clear exit path and can help prevent goroutine leaks.

## Mental Model

```text
context.Background()
        │
        ▼
   WithCancel()
        │
        ├──────────────► ctx
        │                  │
        │                  ▼
        │              ctx.Done()
        │
        └──────────────► cancel()
                           │
                           ▼
                      cancellation
                           │
                           ▼
                     worker returns
                           │
                           ▼
                    goroutine ends
```

The simplest definition to remember:

> **`context.WithCancel()` creates a context whose cancellation can be triggered manually by calling `cancel()`.**
