# Context: Propagation

## Goal

This playground demonstrates **context propagation** in Go.

The main idea is simple:

> A context is passed from an upper layer to lower layers so cancellation, deadlines, and request-scoped values can travel through the call chain.

Example flow:

```text
main
  |
  | ctx
  v
service(ctx)
  |
  | ctx
  v
repository(ctx)
```

The repository does not create a new root context. It receives the context from its caller.

## Example

```go
package main

import (
	"context"
	"fmt"
	"time"
)

func repository(ctx context.Context) {
	fmt.Println("repository started")

	select {
	case <-ctx.Done():
		fmt.Println("repository stopped:", ctx.Err())
		return

	case <-time.After(5 * time.Second):
		fmt.Println("repository finished")
	}
}

func service(ctx context.Context) {
	fmt.Println("service started")

	repository(ctx)

	fmt.Println("service finished")
}

func main() {
	ctx, cancel := context.WithTimeout(
		context.Background(),
		2*time.Second,
	)
	defer cancel()

	service(ctx)

	fmt.Println("main finished")
}
```

## What Is Context Propagation?

Propagation means **passing the same context down the call chain**.

```go
func main() {
	ctx, cancel := context.WithTimeout(...)
	defer cancel()

	service(ctx)
}
```

Then:

```go
func service(ctx context.Context) {
	repository(ctx)
}
```

Then:

```go
func repository(ctx context.Context) {
	...
}
```

Notice that `ctx` is passed through every layer.

```text
main
 │
 │ ctx
 ▼
service
 │
 │ ctx
 ▼
repository
```

The context created in `main` is therefore available to the repository.

---

## Why Does This Matter?

Imagine a real HTTP request:

```text
HTTP Request
     |
     v
Controller
     |
     v
Service
     |
     v
Repository
     |
     v
Database
```

Suppose the client disconnects or the request deadline expires.

You don't want the database operation to continue unnecessarily.

If the context is propagated:

```text
Request Context
      |
      +--> Controller
      |
      +--> Service
      |
      +--> Repository
      |
      +--> Database
```

The lower layers can observe the same cancellation signal.

---

## What Happens in This Example?

The main function creates:

```go
ctx, cancel := context.WithTimeout(
	context.Background(),
	2*time.Second,
)
```

So the context has a **2-second timeout**.

Then:

```go
service(ctx)
```

passes the context into the service.

The service passes it into:

```go
repository(ctx)
```

The repository simulates a 5-second operation:

```go
case <-time.After(5 * time.Second):
	fmt.Println("repository finished")
```

But the context only allows 2 seconds.

Therefore:

```text
0s
 |
 | main creates ctx
 |
 v
service(ctx)
 |
 v
repository(ctx)
 |
 | repository is working...
 |
 | 2 seconds
 v
ctx.Done()
 |
 v
repository stops
```

The repository prints:

```text
repository stopped: context deadline exceeded
```

The 5-second operation does not need to finish because the context's deadline has already expired.

---

## Important: The Context Is Not "Executing"

A common misunderstanding is thinking that the context itself performs the cancellation.

It is better to think of context as a **signal carrier**.

The context provides:

```go
ctx.Done()
```

The repository chooses to listen:

```go
select {
case <-ctx.Done():
	return
}
```

So:

```text
context
   |
   | cancellation signal
   v
repository
   |
   | observes signal
   v
return
```

The context does not forcibly kill the repository function.

---

## Why Not Create a New Context in Every Function?

Avoid doing this:

```go
func service(ctx context.Context) {
	ctx := context.Background()

	repository(ctx)
}
```

This replaces the incoming context with a new unrelated root context.

The original cancellation/deadline is lost.

For example:

```text
main
 |
 | ctx with 2s timeout
 v
service
 |
 | ❌ creates Background()
 v
repository
```

Now the repository no longer receives the original timeout.

Instead:

```go
func service(ctx context.Context) {
	repository(ctx)
}
```

Keep propagating the existing context.

---

## Context Is Usually the First Parameter

A common Go convention is:

```go
func service(ctx context.Context, input Input) error
```

and:

```go
func repository(ctx context.Context, id int) error
```

The context is commonly placed first.

Example:

```go
func GetUser(ctx context.Context, id int) (*User, error)
```

This makes it obvious that the operation is context-aware.

---

## Context Should Usually Flow Downward

Think of the direction as:

```text
caller
  |
  v
function
  |
  v
lower-level function
```

For example:

```text
HTTP handler
     |
     v
service
     |
     v
repository
     |
     v
database
```

The context flows in the same direction:

```text
ctx
 ↓
handler
 ↓
service
 ↓
repository
 ↓
database
```

---

## Parent and Child Contexts

Propagation does not mean every function must use exactly the same context forever.

A function may create a **child context** when it needs a stricter timeout or cancellation scope.

Example:

```go
func service(ctx context.Context) {
	childCtx, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()

	repository(childCtx)
}
```

Now:

```text
parent ctx
    |
    v
service
    |
    | creates child
    v
child ctx
    |
    v
repository
```

The child still depends on the parent.

If the parent is cancelled, the child is also cancelled.

But the child can have a shorter deadline.

---

## Important Rule

A child context cannot extend its parent's lifetime.

Example:

```text
parent deadline: 10 seconds

child deadline: 30 seconds
```

The child effectively cannot live beyond the parent's 10-second deadline.

```text
Parent
|-------------------- 10s ------------------|
                     ↓
                   DONE
                     ↓
Child also DONE
```

Context cancellation flows downward.

---

## What If `repository` Ignores the Context?

For example:

```go
func repository(ctx context.Context) {
	time.Sleep(5 * time.Second)
	fmt.Println("repository finished")
}
```

The context still gets cancelled after 2 seconds.

But the repository does not check it.

Therefore the repository continues sleeping for 5 seconds.

This demonstrates an important principle:

> Passing a context is not enough. The operation must actually use the context.

---

## Real-World Example

Imagine an HTTP request:

```text
Client
  |
  | request
  v
HTTP Handler
  |
  | ctx
  v
Order Service
  |
  | ctx
  v
Order Repository
  |
  | ctx
  v
Database
```

If the request has a 5-second deadline:

```text
request deadline = 5s
```

and the database call receives the same context:

```go
db.QueryContext(ctx, query)
```

the database operation can stop when the context becomes done.

This prevents work from continuing after the request is no longer useful.

---

## Mental Model

Think of context as a **baton**.

```text
main
 🏃
  |
  | passes ctx
  v
service
 🏃
  |
  | passes ctx
  v
repository
 🏃
```

The context is passed downward.

If the context says:

```text
"STOP"
```

the lower-level operation can see the same signal.

```text
main
 |
 | ctx
 v
service
 |
 | ctx
 v
repository
      ↑
      |
   STOP signal
```

---

## Key Takeaways

1. Context propagation means passing a context through function calls.
2. Context commonly flows from higher layers to lower layers.
3. Do not replace an incoming context with `context.Background()`.
4. A context carries cancellation and deadline information.
5. Passing a context alone does not stop work.
6. Lower-level operations must listen to `ctx.Done()` or use context-aware APIs.
7. A child context can have a stricter deadline than its parent.
8. A child cannot outlive its parent.
9. Context is commonly the first parameter of a context-aware function.
10. Real applications often propagate context from HTTP/RPC handlers through services and repositories to external systems or databases.

## Run

```bash
go run .
```

Expected output is approximately:

```text
service started
repository started
repository stopped: context deadline exceeded
service finished
main finished
```

The repository's simulated 5-second operation is interrupted by the 2-second context deadline.

## Related Topics

```text
languages/
└── go/
    └── 02-context/
        ├── 01-background/
        ├── 02-with-cancel/
        ├── 03-done/
        ├── 04-with-timeout/
        ├── 05-with-deadline/
        └── 06-context-propagation/
```

Next:

- Context with HTTP requests
- Context with database operations
- Context-aware workers
- Common context mistakes
