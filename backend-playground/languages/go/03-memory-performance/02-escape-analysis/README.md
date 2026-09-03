# Chapter 2 — Escape Analysis

## Goal

Understand **escape analysis** as a production Go performance tool.

The main use case is an HTTP API that creates response data for every request. The goal is to understand why some values may need heap storage, how to inspect the compiler's decision, and how to verify whether an allocation actually matters.

Escape analysis is not a rule that says "pointers are bad" or "returning values is always faster."

The compiler makes the decision based on whether a value can safely remain on the stack or must outlive its stack frame.

## What You Should Learn

After this chapter, you should understand:

- What escape analysis does.
- The practical difference between stack and heap lifetime.
- Why returning a pointer can cause a value to escape.
- Why a pointer does not automatically mean heap allocation.
- How to inspect compiler escape decisions with `-gcflags=-m` and `-gcflags=-m=2`.
- How escape analysis can affect allocation count and GC work.
- Why compiler output is more reliable than guessing from source code.
- How to validate an optimization with benchmarks.
- When escape analysis is worth investigating in production.

## Production Problem

Imagine an API handler that creates a response object for every request:

```text
HTTP request
    ↓
create response object
    ↓
JSON encode
    ↓
HTTP response
```

Suppose the service receives thousands of requests per second.

If response objects repeatedly need heap storage, those allocations can add to allocation traffic and increase the amount of work the garbage collector has to handle.

The production question is not:

```text
"Can I remove every pointer?"
```

It is:

```text
"Which values are escaping, why are they escaping,
and does that allocation matter for this hot path?"
```

## Concept

### 1. Stack vs heap

At a practical level:

```text
Stack
- tied to function execution
- suitable for values whose lifetime does not need to outlive the frame

Heap
- used when data needs a longer lifetime
- managed by the garbage collector
- can contribute to allocation and GC work
```

These are useful mental models, but the compiler decides the actual placement.

### 2. What escape analysis does

Go's compiler performs escape analysis to determine which variables and implicit allocations can safely be allocated on the stack.

The compiler must preserve important lifetime and pointer-safety properties. It analyzes how values flow through the program rather than applying a simple "pointer means heap" rule.

### 3. A common example

```go
func buildUserPointer(id int) *User {
	user := User{
		ID:   id,
		Name: "Agus",
	}

	return &user
}
```

`user` is declared inside `buildUserPointer`, but a pointer to it is returned.

The caller can still access that `User` after `buildUserPointer` returns.

Conceptually:

```text
buildUserPointer()
        |
        | user lives here
        ↓
return &user
        |
        ↓
caller still uses *User
```

This is the kind of value flow escape analysis must reason about.

### 4. Returning a value is different

Compare:

```go
func buildUserValue(id int) User {
	return User{
		ID:   id,
		Name: "Agus",
	}
}
```

The compiler may keep the value in stack/register storage depending on the surrounding code and optimizations.

Do not turn this into:

```text
pointer → heap
value   → stack
```

That is too simplistic.

The compiler decides.

## Example — Real HTTP Use Case

The runnable example contains:

```go
func usersHandler(w http.ResponseWriter, r *http.Request) {
	user := buildUserPointer(1)

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(user)
}
```

Run it:

```bash
go run .
```

Then:

```bash
curl http://localhost:8080/users/1
```

Example response:

```json
{ "id": 1, "name": "Agus" }
```

In a real application, the object could contain fields loaded from a database or assembled from multiple services.

The important production pattern is:

```text
request
  ↓
create response object
  ↓
pass object to another component
  ↓
encode / serialize
  ↓
response
```

The more frequently this path runs, the more interesting allocation behavior becomes.

## Inspecting Escape Analysis

The Go compiler can print optimization decisions with:

```bash
go build -gcflags=-m .
```

For more detail:

```bash
go build -gcflags=-m=2 .
```

The compiler documentation identifies `-m` as the flag for printing optimization decisions, with higher values providing more detail. 

Look for messages containing:

```text
escapes to heap
```

or:

```text
does not escape
```

The exact output depends on the code and compiler version. Treat the compiler output as the source of truth rather than memorizing individual cases.

## Benchmark

The chapter includes:

```text
BenchmarkBuildUserPointer
BenchmarkBuildUserValue
```

Run:

```bash
go test -bench=. -benchmem
```

Compare:

```text
ns/op
B/op
allocs/op
```

The benchmark is intentionally small. Its purpose is to show how value flow can correlate with allocation behavior while reinforcing that the benchmark result—not a rule such as "pointers are bad"—should drive optimization decisions.

## Experiment

### Experiment 1 — Run the API

```bash
go run .
```

Then:

```bash
curl http://localhost:8080/users/1
```

### Experiment 2 — Inspect escape analysis

```bash
go build -gcflags=-m .
```

Then:

```bash
go build -gcflags=-m=2 .
```

Search the output for:

```text
escapes to heap
```

### Experiment 3 — Benchmark both designs

```bash
go test -bench=. -benchmem
```

Compare:

```text
BenchmarkBuildUserPointer
BenchmarkBuildUserValue
```

Focus on:

```text
B/op
allocs/op
```

Do not assume the pointer version is always slower in a real application.

### Experiment 4 — Change the consumer

Change how the returned value is consumed and run escape analysis again.

The purpose is to observe that escape analysis depends on **how values flow through the program**, not just on one syntax choice.

## What To Observe

Focus on:

```text
source code
    ↓
value flow
    ↓
compiler escape analysis
    ↓
stack / heap decision
    ↓
allocation behavior
```

The same type can behave differently in different contexts.

A `*User` does not automatically prove a heap allocation.

A `User` value does not automatically prove stack allocation.

## Production Implication

Escape analysis is useful when you have evidence that a hot path is producing unnecessary allocations.

A practical workflow is:

```text
production symptom
       ↓
profile / benchmark
       ↓
identify allocation hotspot
       ↓
inspect escape analysis
       ↓
understand why value escapes
       ↓
make a small code change
       ↓
benchmark again
       ↓
verify production impact
```

For example, if a benchmark shows:

```text
Before
1 alloc/op
24 B/op
```

and a safe redesign produces:

```text
After
0 allocs/op
0 B/op
```

that is evidence the redesign changed allocation behavior.

But if the original allocation is tiny and the endpoint is dominated by database latency, changing it may provide no meaningful production benefit.

**Do not optimize escape behavior in isolation from the actual workload.**

## Common Mistakes

### 1. "Pointers always allocate on the heap"

Wrong.

Escape analysis considers how the pointer is used. The compiler can optimize many cases.

### 2. "Values always stay on the stack"

Wrong.

A value can still escape depending on how it is used.

### 3. "If the compiler says escape, the code is bad"

Wrong.

Heap allocation can be completely reasonable.

The compiler is solving a lifetime problem, not judging code quality.

### 4. Manually fighting the compiler

Avoid rewriting code just to make an escape-analysis message disappear.

First determine whether the allocation actually affects the workload.

### 5. Benchmarking source-code appearance

Do not reason:

```text
"&User" = bad
"User"  = good
```

Measure the generated behavior.

## Commands

Run the application:

```bash
go run .
```

Run benchmarks:

```bash
go test -bench=. -benchmem
```

Inspect basic compiler optimization decisions:

```bash
go build -gcflags=-m .
```

Inspect detailed optimization decisions:

```bash
go build -gcflags=-m=2 .
```

Run tests:

```bash
go test ./...
```

## Production Connection to Chapter 1

Chapter 1 taught:

```text
allocation can matter on hot paths
```

Chapter 2 adds:

```text
why some values become heap allocations
```

Together:

```text
allocation hotspot
      ↓
measure B/op + allocs/op
      ↓
inspect escape analysis
      ↓
understand why values escape
      ↓
make a targeted change
      ↓
benchmark again
```

This is the beginning of using compiler diagnostics as part of production performance work.

## Go 1.27 Note

This chapter targets Go 1.27.

Compiler optimizations change between Go releases. Go 1.26 added cases where slice backing storage can be allocated on the stack, illustrating why source-level assumptions about allocation can become outdated as the compiler improves.

Go 1.27 continues to evolve compiler and runtime optimizations, so treat escape-analysis output and benchmark results as observations for the specific Go version and build configuration you are investigating.

## Summary

- Escape analysis determines whether values can safely remain stack-oriented or need a longer lifetime.
- Returning a pointer is a common situation that can cause a value to escape.
- Pointers do not automatically mean heap allocation.
- Values do not automatically mean stack allocation.
- Use `go build -gcflags=-m` or `-m=2` to inspect compiler decisions.
- Use benchmarks to determine whether allocation behavior actually matters.
- Escape analysis is a diagnostic tool, not a checklist for rewriting code.
- Production optimization should follow evidence: measure → inspect → change → measure.

## Next Chapter

**Chapter 3 — Slice Memory**

Now that we understand why values can move into heap-managed lifetimes, the next production problem is:

> **Why can a small slice keep a much larger backing array alive?**
