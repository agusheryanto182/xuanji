# Garbage Collector

## Goal

Understand how Go's garbage collector responds to allocation rate, and how `GOGC` changes the CPU-versus-memory trade-off in production.

## What You Should Learn

- Understand what the Go GC considers garbage.
- Understand why allocation rate matters even when objects are short-lived.
- Understand live heap versus allocation traffic.
- Understand what `GOGC` controls.
- Observe how different `GOGC` values can change GC frequency and memory behavior.
- Know why `runtime.GC()` is a diagnostic/experiment tool, not a normal production optimization.

## Production Problem

A web service can allocate many short-lived objects per request:

```text
request
  ↓
allocate
  ↓
use
  ↓
becomes unreachable
  ↓
GC eventually reclaims/reuses memory
```

For example, `100 KB/request × 10,000 requests/sec ≈ 1 GB/sec` of allocation traffic. This does **not** mean the process grows by 1 GB every second. Most objects may be short-lived.

The production question is how much CPU the GC spends keeping up with allocation, and how much memory the application uses between GC cycles.

## Concept

Go's standard GC is a concurrent, non-moving, tracing garbage collector. It finds live objects by tracing pointers from roots, marks live objects, and makes unreachable memory available for allocation.

Simplified:

```text
allocation
    ↓
heap grows
    ↓
GC starts
    ↓
mark live objects
    ↓
sweep/reuse unreachable memory
    ↓
next cycle
```

Allocation traffic and live memory are different:

```text
1 GB allocated over time
        ↓
50 MB remains live
```

The application generated 1 GB of allocation traffic, but its live heap may remain around 50 MB.

## GOGC

`GOGC` controls the target heap growth between GC cycles.

Conceptually:

```text
higher GOGC
    ↓
more heap growth allowed
    ↓
GC generally runs less frequently
    ↓
less GC CPU overhead
    ↓
potentially more memory usage
```

Lower `GOGC` generally makes GC run more frequently, trading CPU work for lower heap growth.

The default is `GOGC=100`.

Configure it with:

```bash
GOGC=200 go run .
```

## Example

A request handler may create temporary buffers:

```go
func handleRequest() {
    data := make([]byte, 256)
    process(data)
}
```

If this happens at high request volume, allocation traffic can become significant.

Before tuning `GOGC`, ask:

```text
Why are we allocating this much?
```

Removing unnecessary allocations can reduce both allocation traffic and GC work.

## Experiment

Run:

```bash
go run .
```

The program runs the same allocation workload with `GOGC=100` and `GOGC=200`, then prints:

- `HeapAlloc`
- `HeapObjects`
- `NumGC`
- `GCCPUFraction`

Exact values depend on your machine and Go version.

Pay attention to `NumGC`: a higher `GOGC` generally allows more heap growth before another cycle, so the same workload can require fewer GC cycles.

## Benchmark

Run:

```bash
go test -bench=. -benchmem
```

The benchmark compares the same allocation pattern under `GOGC=100` and `GOGC=200`.

Benchmark output can vary with CPU, runtime state, and GC scheduling. It is not a replacement for production profiling.

## What To Observe

### Allocation rate matters

Large amounts of short-lived garbage can still create significant GC work.

### Live heap is different from allocation traffic

High allocation rate does not automatically mean a memory leak.

### GOGC is a CPU/memory trade-off

Think:

```text
GOGC ↑ → memory headroom ↑, GC frequency ↓
GOGC ↓ → memory headroom ↓, GC frequency ↑
```

These are general tendencies, not guarantees for every tiny benchmark.

### GC is concurrent

Go's GC performs most of its work concurrently with the application, though GC activity still consumes CPU and can affect latency.

## Production Implication

When production shows high GC CPU or latency pressure:

```text
measure
  ↓
check allocation rate + heap profiles
  ↓
find allocation/retention hotspots
  ↓
reduce unnecessary allocations
  ↓
measure again
  ↓
then consider GOGC / memory-limit tuning
```

Do not treat high GC activity as the root problem automatically.

## `runtime.GC()` Warning

This experiment uses:

```go
runtime.GC()
```

to make observations easier.

Do not add manual GC calls to request handlers just because memory appears high. Manual GC forces collection at a time chosen by the application and is not a general production memory-management strategy.

## GOMEMLIMIT Preview

Go also supports a runtime memory limit through `GOMEMLIMIT`. It is especially relevant to containerized services with a fixed memory budget.

For now, remember:

```text
GOGC
→ controls the CPU/memory trade-off of GC frequency

GOMEMLIMIT
→ gives the runtime a target for total memory use
```

Memory-limit tuning is covered later.

## Common Mistakes

### Mistake 1: "GC is slow, so increase GOGC"

Not necessarily. First find the allocation and retention causes.

### Mistake 2: "High allocation rate means a memory leak"

No. Short-lived allocations can produce high allocation traffic without retained memory growth.

### Mistake 3: "GC immediately returns memory to the OS"

Not necessarily. Reclaimed memory can remain available to the Go runtime for reuse.

### Mistake 4: "Call runtime.GC() when memory is high"

Usually wrong. Use measurement and profiling first.

### Mistake 5: "Higher GOGC is always faster"

No. It can reduce GC frequency and CPU overhead while increasing memory usage.

## Commands

```bash
go run .
GOGC=50 go run .
go test -bench=. -benchmem
go test ./...
```

## Go 1.27 Note

This chapter targets Go 1.27. Exact GC behavior and benchmark numbers are version- and workload-dependent.

## Summary

- GC reclaims unreachable heap objects.
- High allocation traffic can create significant GC work.
- Live heap and allocation traffic are different.
- `GOGC` is a CPU-versus-memory trade-off.
- Reduce unnecessary allocations before tuning GC.
- Use production metrics and profiles to validate changes.

## Next Chapter

Next: **Memory Retention** — investigate objects that remain reachable longer than intended and distinguish retained memory from ordinary allocation pressure.
