# 03 - Done

## What is `ctx.Done()`?

`ctx.Done()` returns a channel that is used as a **cancellation signal**.

```go
ctx.Done()
```

The returned channel has the type:

```go
<-chan struct{}
```

This means it can be received from, but not sent to.

A common pattern is:

```go
<-ctx.Done()
```

This means:

> Wait until the context is cancelled or otherwise done.

---

## Basic Example

```go
func worker(ctx context.Context) {
	fmt.Println("worker waiting...")

	<-ctx.Done()

	fmt.Println("worker stopped")
}
```

The worker blocks at:

```go
<-ctx.Done()
```

until the context becomes done.

Flow:

```text
worker
  │
  ▼
<-ctx.Done()
  │
  │ BLOCKED
  │
  │ waiting for cancellation
  │
cancel()
  │
  ▼
Done channel becomes ready
  │
  ▼
<-ctx.Done() finishes
  │
  ▼
worker stopped
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
	fmt.Println("worker waiting...")

	<-ctx.Done()

	fmt.Println("worker stopped")
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())

	go worker(ctx)

	time.Sleep(2 * time.Second)

	fmt.Println("cancelling...")
	cancel()

	time.Sleep(1 * time.Second)
}
```

The important sequence is:

```text
WithCancel()
     │
     ▼
    ctx
     │
     ▼
worker(ctx)
     │
     ▼
<-ctx.Done()
     │
     │ BLOCKED
     │
cancel()
     │
     ▼
Done becomes ready
     │
     ▼
worker continues
     │
     ▼
return
```

---

## `Done()` Is a Channel

This is an important connection with the previous concurrency topics.

You already learned:

```go
ch := make(chan int)
```

and:

```go
value := <-ch
```

`ctx.Done()` also gives you a channel:

```go
ctx.Done()
```

Therefore:

```go
<-ctx.Done()
```

is a channel receive.

That is why it can block.

---

## Why Does It Block?

Before cancellation:

```text
context
   │
   ▼
Done channel
   │
   └── no cancellation yet
            │
            ▼
         BLOCKED
```

When:

```go
cancel()
```

is called:

```text
cancel()
   │
   ▼
context becomes cancelled
   │
   ▼
Done channel is closed
   │
   ▼
<-ctx.Done()
   │
   ▼
receive completes
```

The goroutine can then continue executing the next line.

---

## `cancel()` Does Not Send a Normal Value

With a normal channel, you might do:

```go
ch <- 10
```

But you do not send values manually through `ctx.Done()`.

Do not do:

```go
ctx.Done() <- something // ❌
```

The context implementation controls the `Done()` channel.

When cancellation happens, the channel is closed/ready for receiving.

Therefore:

```go
<-ctx.Done()
```

becomes unblocked.

---

## `Done()` Is a Signal

Do not think:

```go
<-ctx.Done()
```

means:

> "Receive a cancellation value."

A better mental model is:

> **"Wait until the context tells me that it is done."**

The channel is primarily used as a signal mechanism.

```text
Context
   │
   ▼
Done()
   │
   ▼
signal channel
   │
   ├── not done → receive blocks
   │
   └── done     → receive continues
```

---

## Using `Done()` with `select`

One of the most common patterns in Go is:

```go
select {
case <-ctx.Done():
	return

case job := <-jobs:
	fmt.Println("processing:", job)
}
```

Now the worker waits for two possible events:

```text
                 worker
                   │
                   ▼
                select
               /                    /                     ▼          ▼
         ctx.Done()     jobs
             │            │
             ▼            ▼
          cancel        new job
             │            │
             ▼            ▼
          return        process
```

This allows a worker to stop while it is also waiting for work.

---

## `Done()` and Goroutine Leaks

This connects directly to the previous goroutine leak topic.

Without a cancellation path:

```text
goroutine
    │
    ▼
waiting forever
    │
    ▼
BLOCKED
```

With `ctx.Done()`:

```text
goroutine
    │
    ▼
waiting
    │
    ▼
ctx.Done()
    │
    │ cancellation
    ▼
return
    │
    ▼
goroutine finishes
```

Context therefore provides a possible exit path for a goroutine.

---

## Important: `Done()` Does Not Kill the Goroutine

`ctx.Done()` does not kill a goroutine.

It only provides a signal.

For example:

```go
func worker(ctx context.Context) {
	<-ctx.Done()

	fmt.Println("worker stopped")

	time.Sleep(10 * time.Second)

	fmt.Println("really finished")
}
```

After:

```go
cancel()
```

the worker continues:

```text
cancel()
   ↓
<-ctx.Done() unblocks
   ↓
worker stopped
   ↓
sleep 10 seconds
   ↓
really finished
   ↓
return
```

The goroutine itself decides what to do after receiving the signal.

---

## A Goroutine Can Ignore `Done()`

Consider:

```go
func worker(ctx context.Context) {
	for {
		fmt.Println("working...")
		time.Sleep(time.Second)
	}
}
```

Even if:

```go
cancel()
```

is called, the worker keeps running because it never checks:

```go
ctx.Done()
```

The context cancellation signal exists, but the worker does not cooperate with it.

```text
cancel()
   │
   ▼
context cancelled
   │
   ▼
worker
   │
   └── ignores ctx.Done()
          │
          ▼
       keeps running
```

---

## Key Takeaways

1. `ctx.Done()` returns a receive-only channel.
2. `<-ctx.Done()` is a channel receive.
3. That receive can block while the context is not done.
4. Calling `cancel()` makes the context done.
5. The `Done()` channel is closed when the context is cancelled.
6. Closing the channel makes receives from it complete.
7. `Done()` is primarily a signal, not a data channel.
8. `cancel()` does not forcibly kill a goroutine.
9. The goroutine must cooperate with cancellation.
10. `ctx.Done()` is commonly used inside `select`.

## Mental Model

```text
                 context
                    │
                    ▼
                 Done()
                    │
                    ▼
              signal channel
                    │
          ┌─────────┴─────────┐
          │                   │
     not cancelled        cancelled
          │                   │
          ▼                   ▼
       BLOCKED             channel closed
                              │
                              ▼
                       receive unblocks
                              │
                              ▼
                           return
```

The simplest definition to remember:

> **`ctx.Done()` is a signal channel that becomes ready when the context is done, allowing blocked goroutines to continue and handle cancellation.**
