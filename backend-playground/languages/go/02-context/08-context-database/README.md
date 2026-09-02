# Context + Database Cancellation

## Goal

This playground demonstrates what happens when a `context.Context` becomes done **while a database query is still running**.

The important flow is:

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
  |
  | ctx
  v
db.QueryContext(ctx, query)
  |
  | query is running
  |
  X context becomes done
  |
  v
QueryContext returns error
  |
  v
repository returns error
  |
  v
service returns error
  |
  v
main handles error
```

This is the important difference from a fake `time.After()` example: here the context is actually passed into `database/sql`.

---

## Core API

The important call is:

```go
rows, err := db.QueryContext(ctx, query)
```

The context belongs to the operation that started the query.

If the context is canceled while the query is executing, the database driver gets an opportunity to cancel the operation.

The exact cancellation behavior depends on the database driver.

---

## Example

```go
package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "modernc.org/sqlite"
)

func repository(ctx context.Context, db *sql.DB) error {
	fmt.Println("repository: query started")

	rows, err := db.QueryContext(ctx, `
		WITH RECURSIVE counter(n) AS (
			SELECT 1
			UNION ALL
			SELECT n + 1
			FROM counter
			WHERE n < 1000000000
		)
		SELECT sum(n)
		FROM counter;
	`)
	if err != nil {
		return fmt.Errorf("query failed: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var result int64

		if err := rows.Scan(&result); err != nil {
			return fmt.Errorf("scan failed: %w", err)
		}
	}

	return rows.Err()
}

func service(ctx context.Context, db *sql.DB) error {
	fmt.Println("service started")

	if err := repository(ctx, db); err != nil {
		return fmt.Errorf("service: %w", err)
	}

	fmt.Println("service finished")
	return nil
}

func main() {
	db, err := sql.Open(
		"sqlite",
		"file:context.db?mode=memory&cache=shared",
	)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(
		context.Background(),
		2*time.Second,
	)
	defer cancel()

	if err := service(ctx, db); err != nil {
		fmt.Println("request failed:", err)
		return
	}

	fmt.Println("main finished")
}
```

---

## What Are We Trying to Observe?

The query is intentionally expensive:

```sql
WITH RECURSIVE counter(n) AS (
    SELECT 1
    UNION ALL
    SELECT n + 1
    FROM counter
    WHERE n < 1000000000
)
SELECT sum(n)
FROM counter;
```

We give the context only:

```go
2 * time.Second
```

So the intended experiment is:

```text
Query starts
    |
    | query is still executing
    |
    | 2 seconds pass
    |
    v
context becomes DONE
    |
    v
database driver receives cancellation
    |
    v
QueryContext returns an error
```

---

## Layer-by-Layer Flow

### 1. `main`

`main` creates the context:

```go
ctx, cancel := context.WithTimeout(
	context.Background(),
	2*time.Second,
)
defer cancel()
```

This establishes the maximum lifetime of the operation.

Then:

```go
service(ctx, db)
```

passes the context downward.

---

### 2. `service`

The service does not create a new root context.

It receives:

```go
func service(ctx context.Context, db *sql.DB) error
```

and passes the same context to the repository:

```go
repository(ctx, db)
```

Flow:

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
```

---

### 3. `repository`

The repository receives the context:

```go
func repository(ctx context.Context, db *sql.DB) error
```

and uses it directly:

```go
db.QueryContext(ctx, query)
```

This is the critical part.

The repository does **not** do:

```go
context.Background()
```

because that would throw away the caller's cancellation/deadline.

---

### 4. Query Is Running

At this point:

```go
db.QueryContext(ctx, query)
```

is executing the SQL query.

Imagine:

```text
repository
    |
    v
QueryContext
    |
    v
SQLite
    |
    | calculating...
    | calculating...
    | calculating...
```

Meanwhile the context has its own timer:

```text
Context
   |
   | 2 seconds
   v
DONE
```

---

### 5. Context Becomes Done

After approximately 2 seconds:

```text
ctx.Done()
    ↓
CLOSED / READY
```

The context's error becomes:

```go
context.DeadlineExceeded
```

The database driver can observe the cancellation and stop the query.

Then `QueryContext` returns an error.

---

## What Does the Repository Receive?

The repository checks:

```go
rows, err := db.QueryContext(ctx, query)
if err != nil {
	return fmt.Errorf("query failed: %w", err)
}
```

So if the driver returns a cancellation-related error, the repository returns it upward.

Conceptually:

```text
database
   ↓
driver error
   ↓
QueryContext
   ↓
repository
   ↓
service
   ↓
main
```

---

## Error Propagation

The service wraps the repository error:

```go
if err := repository(ctx, db); err != nil {
	return fmt.Errorf("service: %w", err)
}
```

Then `main` receives the final error:

```go
if err := service(ctx, db); err != nil {
	fmt.Println("request failed:", err)
	return
}
```

The important concept is:

> The context cancellation happens deep inside the operation, but the error can propagate back through the application layers.

---

## Why `%w`?

The repository uses:

```go
fmt.Errorf("query failed: %w", err)
```

and the service uses:

```go
fmt.Errorf("service: %w", err)
```

`%w` preserves the wrapped error.

That means callers can inspect the underlying cause with:

```go
errors.Is(err, context.DeadlineExceeded)
```

For example:

```go
if errors.Is(err, context.DeadlineExceeded) {
	fmt.Println("operation exceeded its deadline")
}
```

---

## Important Driver Behavior

Do not assume:

> `QueryContext` always instantly kills the SQL query.

The actual behavior depends on the database driver.

`database/sql` provides the context-aware API, but the driver is responsible for implementing the cancellation behavior.

So the correct mental model is:

```text
Context becomes done
       |
       v
database/sql knows the context is done
       |
       v
driver is asked to stop/cancel
       |
       v
driver behavior determines how quickly operation stops
```

This distinction is important in production systems.

---

## Why This Is Better Than `time.After`

A previous learning example may have looked like:

```go
select {
case <-ctx.Done():
	return ctx.Err()

case <-time.After(5 * time.Second):
	// pretend database finished
}
```

That is useful for learning context cancellation.

But it is not an actual database operation.

Here we use:

```go
db.QueryContext(ctx, query)
```

So the context is connected to the real database API.

```text
Fake simulation:

ctx
 ↓
select
 ↓
time.After


Real database API:

ctx
 ↓
QueryContext
 ↓
database driver
 ↓
database
```

---

## `QueryContext` vs `Query`

Context-aware:

```go
db.QueryContext(ctx, query)
```

Non-context-aware:

```go
db.Query(query)
```

When the query belongs to a cancelable operation, prefer the context-aware method.

Other context-aware methods include:

```go
db.QueryRowContext(ctx, query)
```

and:

```go
db.ExecContext(ctx, query)
```

---

## Experiment

Run:

```bash
go mod tidy
```

Then:

```bash
go run .
```

You should see something similar to:

```text
service started
repository: query started
request failed: service: query failed: ...
```

The exact error text can depend on the SQLite driver implementation.

The important thing is that the query should be running long enough for the 2-second context deadline to become relevant.

---

## Experiment: Increase the Timeout

Change:

```go
2*time.Second
```

to:

```go
10*time.Second
```

Then run again.

You may observe that the query has more time to complete.

This demonstrates that the database operation is tied to the context's lifetime.

---

## Experiment: Make the Deadline Very Short

Try:

```go
ctx, cancel := context.WithTimeout(
	context.Background(),
	1*time.Nanosecond,
)
```

Then run the program.

The context will likely be done before the query gets a chance to complete.

This is useful for seeing how cancellation can happen before an operation really gets started.

---

## Experiment: Check `ctx.Err()`

You can inspect the context:

```go
if err := ctx.Err(); err != nil {
	fmt.Println("context error:", err)
}
```

Possible values include:

```go
context.Canceled
```

or:

```go
context.DeadlineExceeded
```

---

## Important Mental Model

There are **two different things** happening:

### Context

```text
"Your operation is no longer allowed to continue."
```

### Database driver

```text
"Okay, I'll attempt to stop the database operation."
```

So:

```text
Context
  |
  | cancellation signal
  v
database/sql
  |
  | cancellation request
  v
driver
  |
  v
database
```

This is why context is a mechanism for **coordinating cancellation**, not a magical force that kills arbitrary code.

---

## Production Pattern

A typical HTTP application may look like:

```text
HTTP Request
      |
      v
r.Context()
      |
      v
Handler
      |
      v
Service
      |
      v
Repository
      |
      v
db.QueryContext(ctx, query)
```

If the client disconnects or the request deadline expires:

```text
request canceled
      |
      v
request context DONE
      |
      v
QueryContext receives cancellation
      |
      v
driver attempts cancellation
      |
      v
repository returns error
      |
      v
service returns error
      |
      v
handler returns
```

This is one of the most important real-world uses of Go context.

---

## Key Takeaways

1. `QueryContext` connects a database query to a context.
2. Context cancellation can happen while a query is executing.
3. The database driver determines how cancellation is actually handled.
4. Context cancellation is cooperative.
5. Propagate the caller's context into the repository.
6. Do not replace it with `context.Background()`.
7. Use `%w` when wrapping context-related errors.
8. Use `errors.Is` when checking for `context.DeadlineExceeded` or `context.Canceled`.
9. `QueryContext`, `QueryRowContext`, and `ExecContext` are context-aware database APIs.
10. A context timeout does not guarantee an immediate physical termination of database work.
11. The important production pattern is:

```text
request
  ↓
context
  ↓
service
  ↓
repository
  ↓
database operation
```

12. The purpose of this playground is to observe what happens when the context becomes done **during** a real database operation.

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
        ├── 07-context-http/
        └── 08-context-database/
```

Next:

- Context-aware workers
- Common context mistakes
