# Chapter 1 — Allocation

## Goal

Understand memory allocation from a **production Go** perspective.

The main example is a high-traffic HTTP API. The goal is not to memorize that "allocation is bad", but to understand when allocation becomes worth investigating, how to measure it, and how it can affect CPU, garbage collection, and latency.

## What You Should Learn

After this chapter, you should understand:

- What memory allocation means in Go.
- Why having many variables does not automatically mean many allocations.
- Why allocation is normal and should not be avoided blindly.
- Why allocation on a hot path can matter at high traffic.
- How to read `ns/op`, `B/op`, and `allocs/op`.
- How to write a benchmark that does not easily get optimized away by the compiler.
- How allocation can contribute to GC pressure.
- When allocation is worth investigating in production.
- Why optimization should be validated with measurements rather than assumptions.

## Production Problem

Imagine a Go service with this endpoint:

```text
GET /users
```

The handler loads users from a database and returns them as JSON.

The simplified flow is:

```text
HTTP request
    ↓
query database
    ↓
create / populate User values
    ↓
append to slice
    ↓
JSON encoding
    ↓
HTTP response
```

A single request may only perform a small amount of allocation.

The problem appears when the endpoint becomes a **hot path**.

For example:

```text
1 request  → several KB of allocation
1,000 req/s
       ↓
allocation traffic becomes significant
       ↓
more garbage
       ↓
GC has more work to do
       ↓
CPU / latency may increase
```

This does not mean every allocation immediately increases RAM permanently.

The important questions are allocation rate, object lifetime, and the actual impact on the workload.

## Concept

### 1. What is allocation?

Allocation is the process of requesting memory to store data.

Examples:

```go
users := make([]User, 0, 100)
```

and:

```go
user := &User{
	ID:   1,
	Name: "Agus",
}
```

But do not assume:

```text
variable exists → allocation definitely happened
```

The Go compiler can use stack or registers, perform escape analysis, and apply optimizations.

### 2. Allocation is not automatically bad

Go applications naturally allocate memory.

Allocation becomes interesting when it is:

- happening very frequently,
- happening on a hot path,
- creating many temporary objects,
- producing a large amount of garbage,
- or proven to contribute to CPU, latency, or GC overhead.

The production mindset is:

```text
Don't ask:
"How do I eliminate all allocations?"

Ask:
"Are these allocations actually causing a problem?"
```

### 3. Allocation vs live memory

Suppose an operation produces:

```text
10 KB of allocation
```

That does not mean the process permanently uses an additional 10 KB.

Objects that are no longer reachable can become garbage and may later be processed by the garbage collector.

Therefore:

```text
allocation volume != live memory
```

### 4. Why hot paths matter

A function called once a day is rarely worth optimizing for allocation.

A function called:

```text
10,000 times/sec
```

is much more interesting.

A small allocation per operation can become significant when multiplied by request volume.

For example:

```text
2 KB/request × 10,000 requests/sec
= 20 MB/sec allocation traffic
```

This is **allocation traffic**, not a claim that RAM must grow by 20 MB every second.

## Example — Real HTTP Use Case

The `main.go` example simulates a simple production-style HTTP endpoint:

```go
func getUsers() []User {
	users := make([]User, 0, 3)

	for i := 1; i <= 3; i++ {
		users = append(users, User{
		ID:   i,
		Name: fmt.Sprintf("user-%d", i),
		})
	}

	return users
}
```

The handler calls this function for every HTTP request:

```go
func usersHandler(w http.ResponseWriter, r *http.Request) {
	users := getUsers()

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(users)
}
```

Run the application:

```bash
go run .
```

Then:

```bash
curl http://localhost:8080/users
```

Example response:

```json
[
  { "id": 1, "name": "user-1" },
  { "id": 2, "name": "user-2" },
  { "id": 3, "name": "user-3" }
]
```

In a real production service, `getUsers()` could instead process hundreds or thousands of database rows.

The important pattern is:

```text
request
  ↓
getUsers()
  ↓
slice + User values + string formatting
  ↓
JSON encoding
```

The same pattern can become expensive when executed at high request rates.

## Benchmark

The benchmark measures allocation behavior for `getUsers()`:

```bash
go test -bench=. -benchmem
```

Pay attention to:

```text
ns/op
B/op
allocs/op
```

### `ns/op`

Approximate time per benchmark operation.

### `B/op`

Bytes allocated per benchmark operation.

### `allocs/op`

Number of allocation events per benchmark operation.

For example:

```text
100 ns/op
1000 B/op
5 allocs/op
```

means one benchmark operation averages about 100 ns and causes approximately 1000 bytes of allocation across 5 allocation events.

## Why Does the Benchmark Use a Sink?

The benchmark contains:

```go
var sinkUsers []User
```

and:

```go
sinkUsers = getUsers()
```

The sink makes the benchmark result observable outside the local loop, reducing the chance that the compiler can eliminate the work as dead code.

This is a **benchmarking technique**, not an application pattern you should copy into production code.

## Experiment

### Experiment 1 — Run the application

```bash
go run .
```

Then:

```bash
curl http://localhost:8080/users
```

Every request executes `getUsers()`.

### Experiment 2 — Run the benchmark

```bash
go test -bench=. -benchmem
```

Record:

```text
ns/op
B/op
allocs/op
```

### Experiment 3 — Think in production scale

Imagine the function is called at:

```text
100 requests/sec
1,000 requests/sec
10,000 requests/sec
```

A small amount of allocation per request can become significant at high request rates.

Do not conclude that allocation is automatically the bottleneck. Use the numbers as a reason to investigate further.

## What To Observe

Focus on three things:

1. `B/op` tells you the allocation volume per benchmark operation.
2. `allocs/op` tells you how many allocation events occur per operation.
3. Reducing allocations does not automatically mean the entire service becomes faster.

Performance optimization involves trade-offs.

For example, object pooling can reduce allocations but introduce additional complexity and may affect memory retention or contention.

## Production Implication

When a production service shows symptoms such as:

```text
CPU usage is high
GC activity is high
latency is increasing
```

allocation is one area worth investigating.

A useful workflow is:

```text
production symptom
       ↓
measure
       ↓
profile
       ↓
find allocation hotspot
       ↓
benchmark candidate change
       ↓
optimize
       ↓
benchmark again
       ↓
measure production impact
```

Do not optimize simply because a function uses `make`, `append`, or pointers.

Look for **hot paths and evidence** first.

## Common Mistakes

### 1. Assuming every allocation is bad

Wrong.

Allocation is a normal part of Go programs.

### 2. Assuming many variables mean many allocations

Wrong.

The compiler can use stack/register storage and perform other optimizations.

### 3. Treating `B/op` as leaked memory

Wrong.

`B/op` represents allocation volume measured by the benchmark. It is not permanent live memory.

### 4. Optimizing before profiling

Do not.

Start with:

```text
measure → identify → optimize → measure
```

### 5. Writing a benchmark that the compiler can optimize away

If the result of an operation is never used, the compiler may remove or simplify work that you intended to measure.

That is why this benchmark uses a sink.

## Commands

Run the application:

```bash
go run .
```

Run the benchmark:

```bash
go test -bench=. -benchmem
```

Run tests:

```bash
go test ./...
```

## Go 1.27 Note

This chapter targets Go 1.27.

Go 1.27 introduced size-specialized allocation routines for some small allocations. As a result, the exact cost of allocation can change between Go versions.

Do not memorize benchmark numbers from this chapter as universal numbers.

The important production skill is:

```text
measure your workload
```

## Summary

- Allocation means requesting memory.
- Allocation is normal.
- Many variables do not automatically mean many allocations.
- Allocation on a hot path can matter when traffic is high.
- `B/op` and `allocs/op` help measure allocation in benchmarks.
- Allocation volume is not the same as live memory.
- Garbage can later be processed by the garbage collector.
- Benchmarks must be designed so the work being measured actually happens.
- Allocation optimization should start from measurement and profiling.
- The production question is not "How do I eliminate all allocations?"
- The production question is "Which allocations are proven to matter?"

## Next Chapter

**Chapter 2 — Escape Analysis**

Now that we understand why allocation can matter, the next question is:

> **Why does a value need to live longer, and how does the compiler decide whether it can remain on the stack or needs to escape to the heap?**
