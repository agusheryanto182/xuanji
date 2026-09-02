# Context: Common Mistakes

## Overview

This material collects common mistakes when using `context.Context` in Go.

The goal is not to memorize rules blindly. The goal is to understand the design behind them:

- context represents cancellation, deadlines, and request-scoped values
- context should usually flow from caller to callee
- cancellation is cooperative
- derived contexts should be canceled when the caller is finished with them
- ordinary business data should not be hidden inside context values

---

## 1. Creating `context.Background()` in the Middle of a Request

A common mistake is:

```go
func service() {
    ctx := context.Background()
    repository(ctx)
}
```

This creates a completely new root context.

If the original request is canceled, the new `Background()` context does not know about it.

### Correct pattern

Pass the existing context downward:

```go
func service(ctx context.Context) error {
    return repository(ctx)
}
```

The flow should look like:

```text
HTTP request
    |
    v
handler(ctx)
    |
    v
service(ctx)
    |
    v
repository(ctx)
```

Do not break the chain unnecessarily.

---

## 2. Forgetting `cancel()`

When creating a derived context:

```go
ctx, cancel := context.WithTimeout(parent, time.Second)
```

you should normally call:

```go
defer cancel()
```

This releases resources associated with the derived context when the function is finished.

Typical pattern:

```go
ctx, cancel := context.WithTimeout(parent, time.Second)
defer cancel()
```

Even if the timeout will eventually happen automatically, calling `cancel()` when you are done is still the correct ownership pattern.

---

## 3. Thinking `cancel()` Kills a Goroutine

This is one of the most important mistakes.

Calling:

```go
cancel()
```

does **not** forcibly kill:

```go
go worker(ctx)
```

Instead, cancellation makes:

```go
ctx.Done()
```

ready.

The goroutine must cooperate:

```go
func worker(ctx context.Context) {
    for {
        select {
        case <-ctx.Done():
            return
        default:
            // continue work
        }
    }
}
```

Mental model:

```text
cancel()
   |
   v
context becomes done
   |
   v
goroutine notices signal
   |
   v
goroutine returns
```

So context cancellation is a **signal**, not a kill switch.

---

## 4. Ignoring `ctx.Done()`

A function can receive a context but still fail to respond to cancellation:

```go
func badWorker(ctx context.Context) {
    for {
        doWork()
    }
}
```

The function receives `ctx`, but never checks it.

A context is useful only when the operation actually observes it.

For blocking operations, prefer context-aware APIs when available.

Examples include:

```go
db.QueryContext(ctx, query)
```

and:

```go
http.NewRequestWithContext(ctx, ...)
```

---

## 5. Not Propagating Context

Bad:

```go
func handler(ctx context.Context) {
    service()
}

func service() {
    repository(context.Background())
}
```

The cancellation chain is lost.

Better:

```go
func handler(ctx context.Context) {
    service(ctx)
}

func service(ctx context.Context) {
    repository(ctx)
}
```

The same context can flow through multiple layers.

---

## 6. Replacing an Existing Context with `Background()`

Suppose a function receives:

```go
func service(ctx context.Context) error
```

and then does:

```go
ctx = context.Background()
```

This throws away the caller's cancellation and deadline.

Unless you are intentionally creating a new independent operation, do not replace the incoming context.

---

## 7. Misusing `context.WithValue`

Context values exist for request-scoped metadata.

Example:

```go
type requestIDKey struct{}

ctx := context.WithValue(
    context.Background(),
    requestIDKey{},
    "req-123",
)
```

Then:

```go
requestID := ctx.Value(requestIDKey{})
```

This can be useful for metadata such as:

- request ID
- trace information
- authentication-related request metadata

But do not use context as a general-purpose parameter bag.

Bad idea:

```go
ctx = context.WithValue(ctx, "user", user)
ctx = context.WithValue(ctx, "product", product)
ctx = context.WithValue(ctx, "price", price)
```

Normal business data should be explicit:

```go
func createOrder(ctx context.Context, user User, product Product) error
```

This makes dependencies visible in the function signature.

---

## 8. Using `nil` Context

Avoid:

```go
service(nil)
```

A context should normally be a real context:

```go
context.Background()
```

or:

```go
context.TODO()
```

`nil` contexts can cause panics when code tries to call methods such as:

```go
ctx.Done()
```

If a function needs a context, make it a required parameter.

---

## 9. `context.TODO()` Is Not a Magic Context

`context.TODO()` is useful when the correct context is not known yet.

Example:

```go
ctx := context.TODO()
```

It is mainly a placeholder.

It should not become a habit of replacing proper context propagation with `TODO()`.

If you know the operation belongs to an existing request, pass that request's context instead.

---

## 10. Confusing Timeout, Deadline, and Cancellation

These are related but different.

### Cancellation

```go
ctx, cancel := context.WithCancel(parent)
```

Someone explicitly calls:

```go
cancel()
```

### Timeout

```go
ctx, cancel := context.WithTimeout(parent, 5*time.Second)
```

The context becomes done automatically after the duration.

### Deadline

```go
ctx, cancel := context.WithDeadline(parent, deadline)
```

The context becomes done at a specific absolute time.

Mental model:

```text
WithCancel
    |
    +---- explicit cancellation

WithTimeout
    |
    +---- cancellation after duration

WithDeadline
    |
    +---- cancellation at absolute time
```

---

## 11. Forgetting That Cancellation Is Cooperative

Consider:

```go
func worker(ctx context.Context) {
    time.Sleep(10 * time.Second)
}
```

If the context is canceled after one second, the `Sleep` itself does not automatically stop.

The goroutine is still sleeping.

A more responsive pattern is to use a context-aware wait:

```go
select {
case <-ctx.Done():
    return
case <-time.After(10 * time.Second):
    // continue
}
```

For production code, use context-aware APIs whenever the underlying operation supports them.

---

## 12. Context Is Usually the First Parameter

The common Go convention is:

```go
func service(ctx context.Context, userID int) error
```

rather than:

```go
func service(userID int, ctx context.Context) error
```

This makes context propagation consistent across APIs.

---

## 13. The Big Picture

Most context mistakes come from misunderstanding what context is for.

Think of context as:

```text
                Context
                   |
        +----------+----------+
        |          |          |
   cancellation  timeout   deadline
        |
        v
   request-scoped
      metadata
```

It is **not**:

```text
Context
   |
   +-- database for all function arguments
   +-- global state
   +-- goroutine killer
   +-- replacement for normal parameters
```

---

## 14. Recommended Pattern

A typical service flow looks like:

```go
func handler(ctx context.Context) error {
    return service(ctx)
}

func service(ctx context.Context) error {
    return repository(ctx)
}

func repository(ctx context.Context) error {
    return doDatabaseOperation(ctx)
}
```

If a child timeout is needed:

```go
func service(ctx context.Context) error {
    ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
    defer cancel()

    return repository(ctx)
}
```

The service creates a child context while preserving the parent's cancellation chain.

---

## Key Takeaways

1. Do not create `context.Background()` in the middle of an existing request flow.
2. Propagate the incoming context downward.
3. Call `cancel()` for derived contexts when you are done with them.
4. `cancel()` signals cancellation; it does not kill goroutines.
5. Code must observe `ctx.Done()` or use context-aware APIs to react to cancellation.
6. Do not use context as a general-purpose data container.
7. Use `WithValue` only for appropriate request-scoped metadata.
8. Avoid passing `nil` as a context.
9. `context.TODO()` is a placeholder, not a replacement for proper propagation.
10. Understand the difference between cancellation, timeout, and deadline.
11. Context is normally the first parameter in a function.
12. Good context propagation preserves the cancellation/deadline chain from caller to lower layers.

---

## Suggested Practice

After studying this material, try intentionally introducing each mistake into a small program and observe what changes.

Especially compare:

```go
service(ctx)
```

with:

```go
service(context.Background())
```

and compare a worker that checks:

```go
<-ctx.Done()
```

with one that ignores the context completely.

These experiments make context behavior much easier to remember.
