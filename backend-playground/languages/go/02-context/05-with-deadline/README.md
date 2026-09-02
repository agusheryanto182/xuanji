# Context: WithDeadline

## Goal

This playground demonstrates how `context.WithDeadline` works in Go.

The main ideas are:

- `context.WithDeadline` stops a context at a specific point in time.
- `ctx.Done()` provides the cancellation signal.
- `ctx.Err()` tells us why the context ended.
- Context cancellation is cooperative.
- `WithDeadline` is similar to `WithTimeout`, but the API uses an absolute time instead of a duration.

## Example

```go
package main

import (
	"context"
	"fmt"
	"time"
)

func worker(ctx context.Context) {
	fmt.Println("worker started")

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	i := 0

	for {
		select {
		case <-ctx.Done():
			fmt.Println("worker stopped by context:", ctx.Err())
			return

		case <-ticker.C:
			i++
			fmt.Println("worker working...", i)
		}
	}
}

func main() {
	deadline := time.Now().Add(10 * time.Second)

	ctx, cancel := context.WithDeadline(
		context.Background(),
		deadline,
	)
	defer cancel()

	go worker(ctx)

	time.Sleep(11 * time.Second)

	fmt.Println("main finished")
}
```

## `WithTimeout` vs `WithDeadline`

This is the most important concept in this material.

### WithTimeout

```go
ctx, cancel := context.WithTimeout(
	context.Background(),
	10*time.Second,
)
```

Meaning:

> Stop this context after 10 seconds from now.

You provide a **duration**.

```text
now
 |
 |-------- 10 seconds --------|
                              ↓
                         context done
```

### WithDeadline

```go
deadline := time.Now().Add(10 * time.Second)

ctx, cancel := context.WithDeadline(
	context.Background(),
	deadline,
)
```

Meaning:

> Stop this context at this specific time.

You provide an **absolute point in time**.

```text
now
 |
 |---------------------------> deadline
                                ↓
                           context done
```

So:

```text
WithTimeout  → "how long?"
WithDeadline → "until when?"
```

---

## What Is a Deadline?

A deadline is a specific point in time.

For example:

```go
deadline := time.Now().Add(10 * time.Second)
```

If the current time is:

```text
10:00:00
```

the deadline becomes approximately:

```text
10:00:10
```

The context automatically becomes done at that deadline.

---

## What Happens Inside `worker`?

The worker waits for two possible events:

```go
select {
case <-ctx.Done():
	...
case <-ticker.C:
	...
}
```

Think of it as:

```text
                 ┌── ctx.Done()
                 │
worker → select ──┤
                 │
                 └── ticker.C
```

Every second, the ticker becomes ready:

```text
1s → ticker.C → working 1
2s → ticker.C → working 2
3s → ticker.C → working 3
...
```

When the deadline is reached:

```text
10s → ctx.Done() → worker stops
```

Because the `select` is inside:

```go
for {
	...
}
```

the worker keeps waiting for the next event until the context is done.

---

## `ctx.Err()`

The worker prints:

```go
ctx.Err()
```

after receiving:

```go
<-ctx.Done()
```

For a deadline or timeout, the result is normally:

```text
context deadline exceeded
```

So you may see:

```text
worker stopped by context: context deadline exceeded
```

This is useful when code needs to distinguish why an operation stopped.

Common context errors are:

```go
context.Canceled
```

and:

```go
context.DeadlineExceeded
```

You can check them explicitly:

```go
if ctx.Err() == context.DeadlineExceeded {
	fmt.Println("deadline exceeded")
}
```

---

## Why `defer cancel()` Still Exists

Even though the deadline automatically ends the context:

```go
ctx, cancel := context.WithDeadline(...)
defer cancel()
```

is still a good pattern.

If the operation finishes before the deadline, calling `cancel()` releases resources associated with the context rather than waiting for the deadline.

Common pattern:

```go
ctx, cancel := context.WithDeadline(parent, deadline)
defer cancel()
```

---

## Important: Deadline Is Absolute

Compare these:

```go
context.WithTimeout(parent, 10*time.Second)
```

and:

```go
deadline := time.Now().Add(10 * time.Second)

context.WithDeadline(parent, deadline)
```

For this simple example, they behave almost the same.

The difference becomes useful when a deadline is already known.

For example:

```go
deadline := requestDeadline
ctx, cancel := context.WithDeadline(parent, deadline)
```

Now multiple operations can share the same final deadline.

Example:

```text
Request deadline: 10:00:10

Database operation ──┐
                     ├── must finish before 10:00:10
HTTP request ────────┤
                     │
Worker operation ────┘
```

This is one reason deadlines are useful in real applications.

---

## Parent Deadline Matters

A child context cannot extend its parent's deadline.

For example:

```go
parent, cancel := context.WithTimeout(
	context.Background(),
	5*time.Second,
)
defer cancel()

child, cancel := context.WithTimeout(
	parent,
	30*time.Second,
)
defer cancel()
```

The child does **not** get 30 seconds.

The parent ends after 5 seconds, so the child also becomes done.

Conceptually:

```text
Parent
  |
  | 5 seconds
  ↓
DONE
  |
  ↓
Child also DONE
```

The effective deadline is the earlier one.

---

## `WithTimeout` and `WithDeadline` Relationship

Conceptually:

```text
WithTimeout(parent, duration)
             |
             | equivalent idea
             v
WithDeadline(parent, time.Now().Add(duration))
```

`WithTimeout` is convenient when you think in terms of duration.

`WithDeadline` is convenient when you already have a specific time.

---

## Mental Model

Think:

```text
WithTimeout
    ↓
"Give this operation 10 seconds."

WithDeadline
    ↓
"This operation must stop at 10:00:10."
```

Both eventually produce:

```text
ctx.Done()
     ↓
worker notices cancellation
     ↓
worker returns
```

---

## Key Takeaways

1. `context.WithDeadline` creates a context that ends at a specific time.
2. A deadline is an absolute point in time.
3. `WithTimeout` uses a duration.
4. `WithDeadline` uses a `time.Time`.
5. `ctx.Done()` signals that the context is finished.
6. `ctx.Err()` tells you why it finished.
7. A deadline can result in `context.DeadlineExceeded`.
8. Context cancellation does not forcibly kill goroutines.
9. A child cannot outlive its parent's deadline.
10. The effective deadline is the earliest applicable deadline.
11. `defer cancel()` is still good practice.
12. Deadlines are especially useful when multiple operations must share one final time limit.

## Run

```bash
go run .
```

Expected output is approximately:

```text
worker started
worker working... 1
worker working... 2
worker working... 3
...
worker working... 10
worker stopped by context: context deadline exceeded
main finished
```

Exact timing can vary slightly because goroutine scheduling and timers are not guaranteed to happen at an exact wall-clock instant.

## Related Topics

```text
languages/
└── go/
    └── 02-context/
        ├── 01-background/
        ├── 02-with-cancel/
        ├── 03-done/
        ├── 04-with-timeout/
        └── 05-with-deadline/
```

Next:

- Context propagation
- Context with HTTP requests
- Context with database operations
- Context-aware workers
- Common context mistakes
