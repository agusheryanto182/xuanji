# Context: WithTimeout

## Goal

This playground demonstrates how `context.WithTimeout` works in Go.

The main ideas are:

- A context can automatically become done after a specified duration.
- `ctx.Done()` provides a signal that the operation should stop.
- Cancellation does not forcibly kill a goroutine.
- The goroutine must cooperate by checking `ctx.Done()`.
- `defer cancel()` should generally be used to release resources associated with the child context.

## Example

```go
package main

import (
	"context"
	"fmt"
	"time"
)

func worker(ctx context.Context) {
	i := 0

	fmt.Println("worker started")

	for {
		select {
		case <-ctx.Done():
			fmt.Println("worker stopped by context")
			return

		default:
			i++
			time.Sleep(1 * time.Second)
			fmt.Println("worker working...", i)
		}
	}
}

func main() {
	ctx, cancel := context.WithTimeout(
		context.Background(),
		10*time.Second,
	)
	defer cancel()

	go worker(ctx)

	time.Sleep(11 * time.Second)

	fmt.Println("main finished")
}
```

## How It Works

### 1. Create a timeout context

```go
ctx, cancel := context.WithTimeout(
	context.Background(),
	10*time.Second,
)
```

This creates a child context that will automatically become done after 10 seconds.

Conceptually:

```text
Background Context
       |
       v
WithTimeout(10s)
       |
       v
      ctx
```

After approximately 10 seconds, the context is cancelled automatically.

---

### 2. Start the worker

```go
go worker(ctx)
```

The worker receives the context and continuously checks whether it has been cancelled.

---

### 3. Check `ctx.Done()`

```go
select {
case <-ctx.Done():
	fmt.Println("worker stopped by context")
	return

default:
	// continue working
}
```

When the timeout expires, `ctx.Done()` becomes ready.

The worker then enters:

```go
case <-ctx.Done():
```

and returns.

The context does **not** forcibly terminate the goroutine.

The worker chooses to stop because its code observes the cancellation signal.

---

### 4. Why does the worker keep working?

The `select` is inside an infinite loop:

```go
for {
	select {
	...
	}
}
```

So the worker repeatedly checks the context.

The simplified lifecycle is:

```text
worker starts
     |
     v
check ctx.Done()
     |
     +---- not done ----> do work
     |                      |
     |                      v
     |                 check again
     |
     +---- done ----------> return
```

---

## Important Detail: `default` + `Sleep`

This example uses:

```go
default:
	i++
	time.Sleep(1 * time.Second)
	fmt.Println("worker working...", i)
```

This is useful for learning, but it is not the most responsive cancellation pattern.

While the worker is inside:

```go
time.Sleep(1 * time.Second)
```

it is not checking `ctx.Done()`.

For example, if the timeout happens at `10.0s` while the worker is sleeping, the worker may only observe the cancellation after the sleep finishes.

So cancellation is cooperative and checked at points where the goroutine can observe it.

A more responsive pattern is to make the waiting itself part of the `select`.

Example:

```go
func worker(ctx context.Context) {
	i := 0

	fmt.Println("worker started")

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			fmt.Println("worker stopped by context")
			return

		case <-ticker.C:
			i++
			fmt.Println("worker working...", i)
		}
	}
}
```

Now the worker can react to either:

```text
ctx.Done()
   OR
ticker.C
```

whichever becomes ready first.

---

## Why `main` Sleeps for 11 Seconds

The timeout is:

```go
10 * time.Second
```

while `main` waits:

```go
time.Sleep(11 * time.Second)
```

This is intentionally longer than the timeout so the worker has enough time to observe the timeout and stop.

If `main` only waited 3 seconds:

```go
time.Sleep(3 * time.Second)
```

the program would exit before the 10-second timeout.

Remember:

> When `main` returns, the Go process exits and remaining goroutines are terminated with the process.

For production code, prefer synchronization such as `sync.WaitGroup`, a result channel, or another appropriate coordination mechanism instead of arbitrary `time.Sleep`.

---

## `defer cancel()`

Even though `WithTimeout` automatically cancels the context after 10 seconds, keep:

```go
defer cancel()
```

This is good practice.

If the operation finishes early, calling `cancel()` releases resources associated with the context instead of waiting until the timeout.

The common pattern is:

```go
ctx, cancel := context.WithTimeout(parent, timeout)
defer cancel()
```

---

## Mental Model

Think of `WithTimeout` as:

```text
Start operation
      |
      v
   10-second
    timer
      |
      v
context becomes Done
      |
      v
worker notices signal
      |
      v
worker stops itself
```

It is better to think:

> "The context tells the goroutine to stop."

Not:

> "The context kills the goroutine."

---

## Key Takeaways

1. `context.WithTimeout` automatically cancels a context after a duration.
2. `ctx.Done()` is the cancellation signal.
3. `ctx.Done()` returns a receive-only channel.
4. `<-ctx.Done()` waits until the context is done.
5. Context cancellation is cooperative.
6. A goroutine must check the context and decide to stop.
7. Putting `select` inside a loop allows continuous cancellation checks.
8. `time.Sleep` inside `default` can delay cancellation response.
9. `defer cancel()` is good practice even with `WithTimeout`.
10. `main` must stay alive long enough for the worker to observe the timeout.
11. In production, synchronization is preferable to arbitrary sleeps.

## Run

```bash
go run .
```

Expected behavior is approximately:

```text
worker started
worker working... 1
worker working... 2
worker working... 3
...
worker working... 10
worker stopped by context
main finished
```

The exact timing can vary slightly because scheduling and sleep timing are not guaranteed to be exact.

## Related Topics

This playground belongs to:

```text
languages/
└── go/
    └── 02-context/
        └── 04-with-timeout/
```

Next concepts:

- `WithDeadline`
- context propagation
- context with HTTP requests
- context with database operations
- context-aware workers
- common context mistakes
