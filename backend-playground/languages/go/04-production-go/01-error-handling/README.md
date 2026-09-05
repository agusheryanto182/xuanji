# Chapter 01 — Production Error Handling

## Goal

Learn how to design and handle errors in Go production code so that failures remain understandable, classifiable, and useful to callers and operators.

## What You Should Learn

- An error is part of a function's contract, not just a message.
- Preserve the original cause when adding context.
- Distinguish errors by meaning, not by string matching.
- Understand sentinel errors and typed errors.
- Use `errors.Is` for identity/meaning checks.
- Use `errors.As` when callers need structured error data.
- Wrap errors with `%w` when the underlying cause should remain discoverable.
- Avoid leaking low-level implementation details across inappropriate boundaries.

## Production Problem

Imagine an HTTP handler that asks a repository for a user:

```text
HTTP handler
    ↓
service
    ↓
repository
    ↓
database
```

The database can fail for many reasons. The handler should not need to understand every database-specific error.

A production system needs to answer questions such as:

- Is the requested user missing?
- Is this an internal failure?
- Can the caller retry?
- Should the HTTP layer return 404, 409, or 500?
- Can logs preserve the underlying cause?

The important design problem is **error semantics across boundaries**.

## Concept

### 1. Error values carry meaning

This is weak:

```go
return errors.New("user not found")
```

and later:

```go
if err.Error() == "user not found" {
    // ...
}
```

String matching is fragile because the string is presentation, not a stable contract.

Prefer a stable sentinel when the caller needs to recognize a known condition:

```go
var ErrUserNotFound = errors.New("user not found")
```

Then:

```go
if errors.Is(err, ErrUserNotFound) {
    // ...
}
```

### 2. Wrap when adding context

Suppose the repository returns:

```go
return ErrUserNotFound
```

The service can add useful context:

```go
return fmt.Errorf("get user %d: %w", id, err)
```

The resulting error contains both:

```text
context: get user 42
cause:   ErrUserNotFound
```

The important part is that `errors.Is` can still discover the original meaning.

### 3. `%v` versus `%w`

Compare:

```go
fmt.Errorf("get user %d: %v", id, err)
```

with:

```go
fmt.Errorf("get user %d: %w", id, err)
```

`%v` formats the error but does not preserve it as a wrapped cause for the standard error-chain operations.

`%w` wraps the error and preserves the chain.

Use `%w` when callers should be able to inspect the underlying cause.

### 4. Sentinel errors

A sentinel represents a stable, recognizable condition:

```go
var ErrUserNotFound = errors.New("user not found")
var ErrConflict = errors.New("conflict")
```

Use this when the important information is the **category/identity of the failure**, rather than additional structured fields.

### 5. Typed errors

A typed error is useful when callers need structured information:

```go
type ValidationError struct {
    Field string
    Reason string
}

func (e *ValidationError) Error() string {
    return e.Field + ": " + e.Reason
}
```

A caller can retrieve the type:

```go
var validationErr *ValidationError

if errors.As(err, &validationErr) {
    fmt.Println(validationErr.Field)
}
```

Mental model:

```text
errors.Is
    ↓
"Is this error meaning/category X?"

errors.As
    ↓
"Does this error contain structured type X?"
```

## Example

A small production-style flow:

```go
package main

import (
    "errors"
    "fmt"
)

var ErrUserNotFound = errors.New("user not found")

type Repository struct{}

func (Repository) FindUser(id int) error {
    if id == 42 {
        return ErrUserNotFound
    }

    return nil
}

type Service struct {
    repo Repository
}

func (s Service) GetUser(id int) error {
    if err := s.repo.FindUser(id); err != nil {
        return fmt.Errorf("get user %d: %w", id, err)
    }

    return nil
}

func main() {
    service := Service{repo: Repository{}}

    err := service.GetUser(42)
    if err != nil {
        fmt.Println(err)

        if errors.Is(err, ErrUserNotFound) {
            fmt.Println("caller can treat this as not found")
        }
    }
}
```

Expected output is conceptually:

```text
get user 42: user not found
caller can treat this as not found
```

The service adds context without destroying the error's meaning.

## Experiment

Run:

```bash
go run .
```

Then compare the behavior of these two versions:

```go
fmt.Errorf("get user %d: %v", id, err)
```

and:

```go
fmt.Errorf("get user %d: %w", id, err)
```

The printed message may look similar.

The important difference appears when using:

```go
errors.Is(...)
```

Ask yourself:

> Can the caller still recognize the original error after the service adds context?

That is the production reason `%w` matters.

## Benchmark

This chapter does not include a benchmark by default.

Error handling is primarily about **correct semantics and failure boundaries**, not making error creation artificially fast.

Do not optimize error paths before you have evidence that error handling is a measurable production bottleneck.

## What To Observe

Focus on these questions:

1. Can the caller distinguish "not found" from "database failure"?
2. Does adding context preserve the original cause?
3. What information belongs to the lower layer?
4. What information should cross the service boundary?
5. Is a sentinel enough, or does the caller need structured fields?

## Production Implication

A good production error chain should let each layer add context without destroying information:

```text
database failure
      ↓
repository context
      ↓
service context
      ↓
HTTP/application decision
```

For example:

```text
handle get user
    ↓
get user 42
    ↓
query user
    ↓
database error
```

The upper layer should be able to inspect the chain and make a meaningful decision.

This becomes especially important when mapping application failures to HTTP responses, retries, metrics, and logs.

## Common Mistakes

### 1. Comparing error strings

Avoid:

```go
if err.Error() == "user not found" {
}
```

Use:

```go
if errors.Is(err, ErrUserNotFound) {
}
```

### 2. Destroying the error chain

Avoid `%v` when the cause must remain discoverable:

```go
return fmt.Errorf("get user: %v", err)
```

Prefer:

```go
return fmt.Errorf("get user: %w", err)
```

### 3. Creating an interface for every error

Do not introduce abstractions just because they are possible.

Start with the simplest error contract that communicates the required meaning.

### 4. Exposing implementation details

A PostgreSQL-specific error does not necessarily belong in an HTTP API contract.

Separate:

```text
internal cause
    ↓
application meaning
    ↓
external response
```

## Commands

Run the example:

```bash
go run .
```

Run tests:

```bash
go test ./...
```

## Summary

Production error handling is about preserving **meaning and context** across boundaries.

Remember:

```text
errors.Is
    → recognize a known error meaning

errors.As
    → extract structured error information

%w
    → wrap while preserving the error chain

error string
    → human-readable context, not a stable API contract
```

The goal is not "more errors" or "more abstractions".

The goal is:

> **Make failures understandable to the next layer without losing their original meaning.**

## Next Chapter

**Chapter 02 — Package Design**

We will look at how package boundaries and dependency direction affect coupling, maintainability, and production architecture.
