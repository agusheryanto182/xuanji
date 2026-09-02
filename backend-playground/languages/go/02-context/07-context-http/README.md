# Context: HTTP Requests

## Goal

This playground demonstrates how `context.Context` is used with HTTP requests in Go.

The key idea is:

> An incoming HTTP request already has a context, and that context can be propagated to the work performed for that request.

Flow:

```text
HTTP Client
    |
    | request
    v
HTTP Handler
    |
    | r.Context()
    v
Business logic
    |
    v
Database / external service
```

## Example

```go
func handler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	fmt.Println("handler started")

	select {
	case <-ctx.Done():
		fmt.Println("request context done:", ctx.Err())
		return

	case <-time.After(5 * time.Second):
		fmt.Fprintln(w, "request finished")
		fmt.Println("handler finished")
	}
}
```

The important line is:

```go
ctx := r.Context()
```

Go's `http.Request` already provides a context associated with that request.

---

## Why Does an HTTP Request Have a Context?

An HTTP request has a lifecycle.

For example:

```text
client sends request
       |
       v
server handles request
       |
       v
handler does work
       |
       v
response
```

But the client can disappear before the server finishes.

For example:

```text
Client
  |
  | request
  v
Server
  |
  | expensive work...
  |
  X client disconnects
```

The server should not necessarily continue doing expensive work if that work is no longer useful.

The request context gives the handler a way to observe that lifecycle.

---

## `r.Context()`

Inside the handler:

```go
ctx := r.Context()
```

This retrieves the context associated with the incoming request.

You can then pass it downward:

```go
service(ctx)
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
	// use ctx
}
```

So the complete flow can be:

```text
HTTP request
     |
     v
r.Context()
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

This is context propagation in a real HTTP application.

---

## What Happens When the Request Is Canceled?

The handler is waiting for one of two events:

```go
select {
case <-ctx.Done():
	// request was canceled
case <-time.After(5 * time.Second):
	// work finished
}
```

Think of it as:

```text
                  ┌── ctx.Done()
                  │
handler → select ─┤
                  │
                  └── 5-second work
```

Whichever event happens first determines what the handler does.

### If the work finishes first

```text
5 seconds
    ↓
work finished
    ↓
send response
```

### If the request is canceled first

```text
client disconnects / request canceled
    ↓
ctx.Done()
    ↓
handler returns
```

---

## Important: Context Does Not Kill Your Work

This:

```go
case <-ctx.Done():
	return
```

means the handler **chooses to stop** after observing the cancellation signal.

The context does not forcibly kill a goroutine.

This is the same cooperative cancellation principle learned earlier.

```text
context
   |
   | cancellation signal
   v
handler
   |
   | observes signal
   v
return
```

---

## Running the Example

Start the server:

```bash
go run .
```

You should see:

```text
server listening on :8080
```

Then open:

```text
http://localhost:8080
```

The handler waits approximately 5 seconds and then returns:

```text
request finished
```

---

## Try Canceling the Request

A useful experiment is to start a request and cancel it before the 5 seconds finish.

For example, use:

```bash
curl http://localhost:8080
```

and interrupt the request before it completes.

The exact behavior can depend on how the client and connection are terminated, but the important concept is that the request context can become done when the request lifecycle is canceled.

The handler is designed to observe:

```go
<-ctx.Done()
```

---

## Propagating the Request Context

Imagine a more realistic application:

```text
                    HTTP Request
                         |
                         v
                    r.Context()
                         |
                         v
                      Handler
                         |
                         | ctx
                         v
                       Service
                         |
                         | ctx
                         v
                     Repository
                         |
                         | ctx
                         v
                      Database
```

The same request context can travel through the entire operation.

For example:

```go
func handler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	service(ctx)
}

func service(ctx context.Context) {
	repository(ctx)
}

func repository(ctx context.Context) {
	// database operation
}
```

This allows lower layers to respect the lifecycle of the original HTTP request.

---

## Context-Aware HTTP Client Requests

Context is also useful when your server calls another HTTP service.

For example:

```go
req, err := http.NewRequestWithContext(
	ctx,
	http.MethodGet,
	"https://example.com",
	nil,
)
if err != nil {
	return err
}

resp, err := http.DefaultClient.Do(req)
if err != nil {
	return err
}
defer resp.Body.Close()
```

Now the outgoing request is connected to the incoming request's context.

Conceptually:

```text
Incoming request
       |
       | ctx
       v
   Your server
       |
       | same ctx
       v
Outgoing HTTP request
       |
       v
External service
```

If the original request is canceled, the outgoing request can also be canceled.

---

## Context + Timeout

You can also derive a child context from the request context.

For example:

```go
func handler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(
		r.Context(),
		2*time.Second,
	)
	defer cancel()

	service(ctx)
}
```

This means:

> Use the incoming request's lifecycle, but impose an additional 2-second limit on this operation.

Remember:

```text
request context
       |
       v
child timeout context
       |
       v
service
```

The child cannot outlive the parent.

If the request is canceled first, the child is canceled.

If the 2-second timeout happens first, the child is canceled.

---

## Parent Cancellation Flows Downward

Suppose:

```text
HTTP request context
        |
        v
service context
        |
        v
repository context
```

If the HTTP request context becomes done:

```text
HTTP request
     |
     X canceled
     |
     v
service
     |
     v
repository
```

The derived contexts can observe the cancellation.

This is one of the most important reasons to propagate context instead of creating unrelated `Background()` contexts deep inside the application.

---

## Common Mistake

Bad:

```go
func service(ctx context.Context) {
	ctx = context.Background()

	repository(ctx)
}
```

The incoming request context has been replaced.

Better:

```go
func service(ctx context.Context) {
	repository(ctx)
}
```

Or derive a child only when there is a real reason:

```go
func service(ctx context.Context) {
	childCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	repository(childCtx)
}
```

---

## `context.Background()` vs `r.Context()`

Use:

```go
context.Background()
```

as a root context when starting a top-level operation.

Inside an HTTP handler, use:

```go
r.Context()
```

because the request already gives you the correct lifecycle context.

Do not casually replace:

```go
r.Context()
```

with:

```go
context.Background()
```

because doing so disconnects the operation from the HTTP request lifecycle.

---

## Mental Model

Think of the HTTP request as the owner of the operation:

```text
HTTP Request
     |
     | owns lifecycle
     v
Request Context
     |
     v
Handler
     |
     v
Service
     |
     v
Repository
```

If the request says:

```text
"STOP"
```

the lower layers can receive the same signal.

---

## Key Takeaways

1. `http.Request` already has a context.
2. Get it with `r.Context()`.
3. Pass the request context into service and repository layers.
4. Context propagation connects lower-level work to the request lifecycle.
5. `ctx.Done()` lets code observe cancellation.
6. Context cancellation is cooperative.
7. Do not replace request context with `context.Background()` inside the request flow.
8. Derive child contexts when you need a stricter timeout or cancellation scope.
9. An outgoing HTTP request can use `http.NewRequestWithContext`.
10. The child context cannot outlive its parent.

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
        ├── 06-context-propagation/
        └── 07-context-http/
```

Next:

- Context with database operations
- Context-aware workers
- Common context mistakes
