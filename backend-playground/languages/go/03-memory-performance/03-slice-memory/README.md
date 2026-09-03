# Chapter 3 — Slice Memory

## Goal

Understand how Go slices interact with their backing arrays and why a **small slice can keep a much larger allocation alive**.

The main use case is a production-style HTTP service that temporarily loads a large buffer but returns only a small portion of it.

## What You Should Learn

After this chapter, you should understand:

- What a slice represents in Go.
- The relationship between a slice and its backing array.
- Why slicing does not automatically copy data.
- How a small slice can retain a large backing array.
- The difference between allocation size and retained/live memory.
- How slice retention can become a production memory problem.
- How to benchmark the trade-off between retaining and copying data.
- When copying a slice is worth the extra allocation.
- How to investigate slice-related memory behavior.

## Production Problem

Imagine a service that loads a large report:

```text
10 MiB input buffer
        ↓
only 100 bytes are needed
        ↓
return data[:100]
```

The returned slice has a length of only 100 bytes, but slicing does not copy the backing array.

Conceptually:

```text
10 MiB backing array
┌──────────────────────────────────────────────┐
│ first 100 bytes │ remaining ~10 MiB         │
└──────────────────────────────────────────────┘
        ↑
      slice
      len = 100
```

If that 100-byte slice remains reachable for a long time, the large backing array can also remain reachable.

This is a **memory retention** problem.

## Concept

### 1. A slice is not the array

A useful mental model is:

```text
slice
 ├── pointer → backing array
 ├── length
 └── capacity
```

For example:

```go
data := make([]byte, 10<<20)
small := data[:100]
```

`small` has length 100, but it still refers to the same backing array.

### 2. Slicing does not copy

This:

```go
small := data[:100]
```

creates a slice view over the existing backing array.

It does not normally mean:

```text
allocate 100 bytes
copy first 100 bytes
```

This is efficient when shared storage is what you want.

### 3. The retention problem

Suppose:

```go
data := make([]byte, 10<<20)
small := data[:100]
```

Then `data` becomes unreachable:

```text
data
 ↓
no longer referenced

small
 ↓
still references backing array
```

The backing array is still reachable through `small`.

So:

```text
slice length       = 100 bytes
backing allocation = 10 MiB
```

Those numbers can be very different.

## Example — Real HTTP Use Case

The runnable example intentionally creates a large buffer and returns only a small portion:

```go
func loadReport() []byte {
	data := make([]byte, 10<<20)

	for i := range data {
		data[i] = 'x'
	}

	return data[:100]
}
```

The handler calls it for every request:

```go
func reportHandler(w http.ResponseWriter, r *http.Request) {
	report := loadReport()

	w.Header().Set("Content-Type", "text/plain")
	_, _ = w.Write(report)
}
```

In a real service, the large buffer could come from a file, HTTP response, message payload, decompression step, or parser.

The important production pattern is:

```text
large temporary buffer
        ↓
small piece is kept
        ↓
small slice keeps backing array reachable
        ↓
memory retention
```

## Fix — Copy the Small Data

If the large backing array is no longer needed, copy the required portion:

```go
func loadReportFixed() []byte {
	data := make([]byte, 10<<20)

	for i := range data {
		data[i] = 'x'
	}

	result := make([]byte, 100)
	copy(result, data[:100])

	return result
}
```

Now the returned slice has its own 100-byte backing array.

The trade-off is:

```text
Without copy:
less CPU/allocation
but potentially more memory retained

With copy:
extra allocation + copy
but large backing array can become unreachable
```

## Benchmark

Run:

```bash
go test -bench=. -benchmem
```

Compare:

```text
BenchmarkLoadReport
BenchmarkLoadReportFixed
```

Look at:

```text
ns/op
B/op
allocs/op
```

The fixed version intentionally performs an additional allocation for the small result.

The benchmark therefore measures the cost of the fix, while the retention problem itself is about **how long the large backing array remains reachable**.

## Experiment

### Experiment 1 — Run the API

```bash
go run .
```

Then:

```bash
curl http://localhost:8080/report
```

The endpoint returns only 100 bytes.

### Experiment 2 — Run the benchmark

```bash
go test -bench=. -benchmem
```

Compare the two implementations.

### Experiment 3 — Think about a cache

Imagine:

```go
cache[key] = loadReport()
```

If `loadReport()` returns a small slice backed by a huge array, the cache entry can keep that large array reachable.

The important question becomes:

```text
How much memory does this cache entry actually retain?
```

Not simply:

```text
How long is the slice?
```

### Experiment 4 — Production investigation

If a production service has unexpectedly high heap usage, use heap profiling to investigate retained objects.

A useful workflow is:

```text
high heap usage
      ↓
heap profile
      ↓
find large retained objects
      ↓
inspect slice / buffer ownership
      ↓
check backing array size
      ↓
copy only when appropriate
      ↓
measure again
```

## What To Observe

The key relationship is:

```text
slice
  ↓
backing array
  ↓
reachability
  ↓
memory retention
```

Do not confuse:

```text
len(slice)
```

with:

```text
size of the backing allocation
```

Also distinguish:

```text
allocation
```

from:

```text
retained/live memory
```

A large temporary allocation is not automatically a problem if it becomes unreachable promptly.

The problem here is when a small long-lived reference accidentally keeps a large backing array alive.

## Production Implication

This pattern can appear in:

- caches,
- long-lived structs,
- queues,
- background jobs,
- HTTP response processing,
- file processing,
- message consumers.

Example:

```go
cache[key] = largeData[:100]
```

The cache entry may retain the entire backing array.

A common fix is:

```go
cache[key] = append([]byte(nil), largeData[:100]...)
```

But copying is not automatically better.

You trade:

```text
retention
```

for:

```text
new allocation + copy cost
```

The right choice depends on:

- backing-array size,
- retained portion size,
- lifetime of the returned data,
- request rate,
- cache size,
- memory pressure,
- profiling evidence.

## Common Mistakes

### 1. Assuming `data[:100]` creates a new 100-byte allocation

Wrong.

It normally creates a slice view over the same backing array.

### 2. Looking only at `len(slice)`

A slice length of 100 does not tell you the size of its backing allocation.

### 3. Copying every slice

Wrong.

Copying adds allocation and CPU cost.

Use it when the small data needs to outlive the large buffer and retention matters.

### 4. Calling this a memory leak

Usually this is better described as **memory retention**.

The memory is still reachable, so the GC is correctly keeping it alive.

### 5. Assuming GC can fix the problem

GC can only reclaim objects that are no longer reachable.

If a live slice still points into the backing array, the backing array remains reachable.

## Commands

Run the application:

```bash
go run .
```

Run benchmarks:

```bash
go test -bench=. -benchmem
```

Run tests:

```bash
go test ./...
```

For production heap investigation, use Go heap profiles and profiling tools.

## Production Connection to Previous Chapters

Chapter 1 taught:

```text
allocation can matter
```

Chapter 2 taught:

```text
escape analysis can explain heap allocation
```

Chapter 3 adds:

```text
a small live slice can retain a large backing allocation
```

The production workflow becomes:

```text
high memory usage
       ↓
profile
       ↓
find retained object
       ↓
inspect slice ownership
       ↓
check backing array
       ↓
copy if the lifetime requires it
       ↓
measure memory + CPU impact
```

## Go 1.27 Note

This chapter targets Go 1.27.

Compiler and runtime optimizations evolve over time, so do not rely on assumptions about exactly where every allocation occurs.

For slice-retention problems, the important production question remains reachability: if a live slice references a backing array, that backing array remains reachable.

## Summary

- A slice is a descriptor pointing to an underlying array.
- Slicing normally does not copy the underlying data.
- A small slice can therefore keep a large backing array reachable.
- `len(slice)` is not the same as backing-array size.
- This can create memory-retention problems in long-lived data structures such as caches.
- Copying the required data can allow the large backing array to become unreachable.
- Copying introduces allocation and CPU cost.
- Do not copy blindly; consider object lifetime and profiling evidence.
- When investigating memory problems, think about **reachability and retention**, not just allocation size.

## Next Chapter

**Chapter 4 — Map Memory**

Next we will look at another common source of unexpected memory usage:

> **Why can a Go map retain substantial memory even after many entries have been deleted?**
