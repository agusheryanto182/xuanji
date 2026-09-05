# Chapter 10 — Production Memory Tuning

## Goal

Learn how to make Go memory decisions in production using measurements and profiles instead of guessing.

## What You Should Learn

- Allocation rate vs live heap.
- GC pressure.
- GOGC trade-offs.
- GOMEMLIMIT as a soft runtime memory limit.
- Bounded caches and object lifetime.
- How to validate tuning with benchmarks and profiles.

## Production Problem

A service has high memory or GC activity. Possible fixes include reducing allocations, changing GOGC, setting GOMEMLIMIT, bounding caches, or increasing the memory limit.

The correct workflow is:

```text
Measure
  ↓
Profile
  ↓
Understand allocation + retention
  ↓
Identify the constraint
  ↓
Choose a trade-off
  ↓
Validate
```

## Concept

Three different questions:

### Allocation traffic

How much memory is allocated over time?

High allocation traffic can create GC work even when live memory is small.

### Live heap

How much allocated memory is still reachable?

This is more relevant to retention and live-memory pressure.

### Process/container memory

How much memory does the environment consume?

It can include Go heap, goroutine stacks, runtime metadata, executable/libraries, mappings, and other process memory.

Therefore:

```text
heap profile != complete container memory accounting
```

## Allocation Rate and GC Pressure

```text
Allocate
   ↓
Objects become unreachable
   ↓
GC discovers/reclaims them
   ↓
Application keeps allocating
```

A small live heap does not automatically mean low memory-management cost.

Reducing unnecessary allocation often reduces allocation traffic and GC pressure.

## GOGC

Simplified model:

```text
Lower GOGC
    ↓
Less heap growth between GC cycles
    ↓
Potentially lower heap usage
    ↓
Potentially more GC CPU

Higher GOGC
    ↓
More heap growth allowed
    ↓
Potentially less GC CPU
    ↓
Potentially higher memory usage
```

These are tendencies, not guarantees.

Do not assume one GOGC value is universally best.

## GOMEMLIMIT

`GOMEMLIMIT` gives the Go runtime a **soft memory limit** to consider when managing memory.

Example:

```bash
GOMEMLIMIT=700MiB
```

It is not a replacement for a container or OS memory limit.

```text
Container memory limit
        |
        | hard operational boundary
        v
+-----------------------------+
| Go runtime memory budget    |
|       GOMEMLIMIT             |
+-----------------------------+
```

Leave headroom for non-Go-heap memory and operational variance.

## GOGC + GOMEMLIMIT

Think of them as different controls:

```text
GOGC
  -> GC heap-growth target

GOMEMLIMIT
  -> runtime memory budget
```

Validate both under realistic workloads.

## Experiment

The example allocates temporary buffers:

```go
func allocateWork(size int) []byte {
    data := make([]byte, size)

    for i := range data {
        data[i] = byte(i)
    }

    return data
}
```

Run:

```bash
go test -bench=. -benchmem
```

Compare:

```bash
GOGC=50 go test -bench=BenchmarkAllocateWork -benchmem
GOGC=100 go test -bench=BenchmarkAllocateWork -benchmem
GOGC=200 go test -bench=BenchmarkAllocateWork -benchmem
```

The exact numbers depend on workload, machine, and Go version.

## Runtime Metrics

Useful `runtime.MemStats` fields include:

```text
HeapAlloc
HeapSys
HeapInuse
HeapObjects
NumGC
PauseTotalNs
GCCPUFraction
```

Use several metrics together.

## Bounded Caches

An unbounded cache can create retention:

```go
cache[key] = value
```

without a meaningful lifecycle.

Production caches commonly need:

```text
maximum size
TTL
eviction
expiration
explicit invalidation
```

The question is not simply "does this cache use memory?"

Ask:

> Is the data retained longer or in greater quantity than the business requirement needs?

## Object Lifetime

A large object can be correctly allocated but incorrectly retained:

```text
Request
   ↓
allocate large object
   ↓
store in long-lived cache
   ↓
request ends
   ↓
object remains reachable
```

Ask:

> Who owns this data, and how long does it actually need to live?

## Reuse vs Allocation

Reuse can reduce allocation traffic, but adds trade-offs:

- synchronization
- ownership complexity
- stale data risks
- larger retained buffers
- accidental sharing

Reuse is an optimization, not a default rule.

## Production Workflow

```text
Production symptom
        ↓
Measure memory + GC
        ↓
Profile allocations / live heap
        ↓
Identify allocation or retention source
        ↓
Form hypothesis
        ↓
Benchmark the change
        ↓
Test realistic workload
        ↓
Deploy gradually
        ↓
Validate memory + CPU + latency
```

Never optimize only one metric.

A change that reduces memory but causes a large latency regression may be a bad trade.

## What To Observe

Compare:

```text
ns/op
B/op
allocs/op
HeapAlloc
HeapObjects
NumGC
GCCPUFraction
latency
CPU
RSS / container memory
```

Look for relationships:

```text
allocation ↑
    → GC work may ↑

live heap ↑
    → memory pressure ↑

GOGC ↑
    → heap may grow more
    → GC work may decrease

GOMEMLIMIT ↓
    → runtime may work harder to stay within budget
```

These are workload-dependent tendencies.

## Benchmark

Run:

```bash
go test -bench=BenchmarkAllocateWork -benchmem
```

Then:

```bash
GOGC=50 go test -bench=BenchmarkAllocateWork -benchmem
GOGC=100 go test -bench=BenchmarkAllocateWork -benchmem
GOGC=200 go test -bench=BenchmarkAllocateWork -benchmem
```

Repeat noisy benchmarks:

```bash
go test -bench=BenchmarkAllocateWork -benchmem -count=10
```

Keep machine, Go version, workload, and command structure controlled.

## Production Implication

Good memory tuning usually combines:

```text
Better allocation behavior
        +
Correct object lifetime
        +
Bounded caches
        +
Appropriate GC settings
        +
Realistic memory limits
```

Do not expect one environment variable to fix an architectural allocation problem.

If the application allocates unnecessarily, reducing the allocation itself is often more fundamental than changing when GC runs.

## Common Mistakes

### 1. Tuning GOGC before profiling

The real problem may be retention or excessive allocation.

### 2. Treating GOMEMLIMIT as a hard limit

It is a soft runtime limit, not a replacement for the container/OS limit.

### 3. Setting GOMEMLIMIT equal to the container limit

Leave headroom for non-Go-heap memory.

### 4. Optimizing for minimum heap

The goal is a memory/CPU/latency/throughput trade-off, not minimum memory at any cost.

### 5. Reusing everything

Reuse can create more complexity than the allocation it avoids.

### 6. Calling runtime.GC() in normal request paths

Useful for experiments/diagnostics, not a normal optimization strategy.

### 7. Looking only at HeapAlloc

Use heap profiles and process/container metrics together.

## Commands

```bash
go test -bench=. -benchmem

GOGC=50 go test -bench=BenchmarkAllocateWork -benchmem
GOGC=100 go test -bench=BenchmarkAllocateWork -benchmem
GOGC=200 go test -bench=BenchmarkAllocateWork -benchmem

GOMEMLIMIT=700MiB go test -bench=BenchmarkAllocateWork -benchmem

go test -bench=BenchmarkAllocateWork -benchmem -memprofile=mem.out
go tool pprof -top mem.out
```

## Summary

Production memory tuning is a trade-off problem.

1. Allocation traffic and live heap are different.
2. High allocation rate can create GC pressure even with a small live heap.
3. Lower GOGC generally favors lower heap growth at the cost of potentially more GC work.
4. Higher GOGC generally allows more heap growth and may reduce GC work.
5. GOMEMLIMIT is a soft runtime memory target, not a hard container limit.
6. Leave operational headroom outside the Go heap.
7. Bounded caches and correct object lifetimes are often more important than GC tuning.
8. Reduce unnecessary allocations before reaching for runtime tuning.
9. Validate memory, CPU, latency, and throughput together.
10. Production tuning should follow measurement and profiling.

## Next Chapter

**Chapter 11 — PGO**

Next we move from runtime memory behavior into Profile-Guided Optimization: using real profile data to influence compiler optimization decisions.
