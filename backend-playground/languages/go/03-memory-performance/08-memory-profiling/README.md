# Chapter 08 — Memory Profiling

## Goal

Learn how to use Go memory profiles to answer a production question:

> Where is this process allocating or retaining memory?

## What You Should Learn

- What memory profiling measures.
- Benchmark vs profiling.
- `inuse_space`, `inuse_objects`, `alloc_space`, and `alloc_objects`.
- Why memory profiling is sampled.
- How to generate a memory profile with `go test`.
- How to inspect a profile with `go tool pprof`.
- How to reason from a profile back to production code.

## Production Problem

Imagine an HTTP service that processes reports. During heavy traffic, memory usage becomes unexpectedly high.

A benchmark tells you:

```text
ns/op
B/op
allocs/op
```

A memory profile helps answer:

> Which functions are responsible for the memory?

The production workflow is:

```text
Memory symptom
      ↓
Collect profile
      ↓
Inspect allocation / live heap
      ↓
Find expensive call path
      ↓
Inspect source
      ↓
Form hypothesis
      ↓
Optimize
      ↓
Profile again
```

## Concept

### Benchmark vs Memory Profile

A benchmark asks:

> How expensive is this operation under a controlled workload?

A memory profile asks:

> Where is memory being allocated or currently retained?

They complement each other.

```text
Benchmark
  ├── ns/op
  ├── B/op
  └── allocs/op

Memory profile
  ├── where memory is allocated
  ├── how much is currently live
  └── which call paths contribute
```

### Four Important Metrics

#### `inuse_space`

Approximate bytes currently in use.

Useful for investigating retained/live heap memory.

Think:

```text
"What memory is still alive?"
```

#### `inuse_objects`

Approximate number of currently live objects.

Useful when many small objects are retained.

Think:

```text
"How many objects are still alive?"
```

#### `alloc_space`

Cumulative bytes allocated over the profiling period, including memory that may already have been garbage collected.

Useful for allocation churn.

Think:

```text
"How much memory did we allocate in total?"
```

#### `alloc_objects`

Cumulative allocation events represented by the profile.

Useful for object/allocation churn.

Think:

```text
"How many objects did we allocate?"
```

### Important Distinction

If a service allocates 100 MB and later the objects become unreachable:

```text
alloc_space  ≈ 100 MB
inuse_space  ≈ small
```

That does not automatically mean a leak.

If a large amount remains live:

```text
alloc_space  = 100 MB
inuse_space  = 90 MB
```

that deserves investigation.

## Sampling

Go's memory profiler is sampling-based. It does not necessarily record every allocation as an individual profile entry.

Therefore:

> Profile values are measurements with sampling behavior, not an exact accounting database of every allocation.

For controlled experiments, the sampling rate can be adjusted with `-memprofilerate`.

## Example

This package contains a small report-processing workload:

```go
func buildReport(rows int) []byte
```

It creates report buffers and temporary work.

The benchmark makes allocation behavior measurable.

The HTTP handlers provide a production-shaped workload for the next profiling chapter.

The important question is not:

> "Can I memorize this function?"

It is:

> "If memory usage becomes a production problem, how do I find where it comes from?"

## Experiment

Run:

```bash
go test -bench=. -benchmem
```

Generate a memory profile:

```bash
go test -bench=BenchmarkTemporaryReports -benchmem -memprofile=mem.out
```

Inspect it:

```bash
go tool pprof -top mem.out
```

Interactive view:

```bash
go tool pprof -http=:0 mem.out
```

Useful pprof commands include:

```text
top
top -cum
list buildReport
web
```

`top` focuses on direct contribution. `top -cum` includes cumulative call-path contribution. `list` connects profile data to source code.

## Benchmark

### Allocation churn

```go
func BenchmarkTemporaryReports(b *testing.B)
```

Repeatedly allocates report buffers while old results become unreachable.

This is useful for understanding:

```text
alloc_space
alloc_objects
```

### Live memory

```go
func BenchmarkRetainedReports(b *testing.B)
```

Keeps a bounded collection of report buffers reachable.

This is useful for understanding:

```text
inuse_space
inuse_objects
```

Mental model:

```text
allocated repeatedly
        ↓
alloc_space

still reachable
        ↓
inuse_space
```

## What To Observe

Ask:

1. Which function allocates the most?
2. Is the allocation cumulative or still live?
3. Is the allocation expected?
4. Does the object live longer than required?
5. Is the allocation on a hot request path?
6. Can allocation size or frequency be reduced?
7. Would reuse actually improve the system?
8. Does the profile agree with the benchmark?

Example reasoning:

```text
inuse_space
    ↓
buildReport
    ↓
make([]byte, ...)
    ↓
large report buffer
    ↓
why is it still reachable?
```

## Production Implication

### High `alloc_space`, low `inuse_space`

Often indicates allocation churn.

Possible consequences:

- CPU spent allocating
- memory bandwidth pressure
- more GC work
- latency impact

It is not necessarily a leak.

### High `inuse_space`

Indicates substantial memory is still live in the profile.

Possible causes:

- intentional cache
- large active workload
- long-lived data
- accidental retention
- oversized buffers
- unbounded collections

The profile tells you where to investigate; it does not automatically tell you the code is wrong.

## Common Mistakes

### 1. Treating `alloc_space` as current RAM

Wrong:

```text
alloc_space = 2 GB
therefore the process uses 2 GB RAM
```

Not necessarily. `alloc_space` is cumulative allocation volume.

### 2. Treating every allocation as a leak

Allocation is normal. A retention problem is about memory remaining reachable longer than intended.

### 3. Looking only at `inuse_space`

A system can have low live memory but enormous allocation churn.

Use the metric that matches the question.

### 4. Profiling without representative workload

A profile is meaningful relative to the workload that produced it.

### 5. Optimizing from one number

Validate:

```text
profile
  ↓
change
  ↓
benchmark/profile again
```

## Commands

Run the example:

```bash
go run .
```

Run benchmarks:

```bash
go test -bench=. -benchmem
```

Generate a memory profile:

```bash
go test -bench=BenchmarkTemporaryReports -benchmem -memprofile=mem.out
```

Inspect:

```bash
go tool pprof -top mem.out
```

Interactive UI:

```bash
go tool pprof -http=:0 mem.out
```

For a controlled experiment:

```bash
go test -bench=BenchmarkTemporaryReports -memprofile=mem.out -memprofilerate=1
```

Do not blindly use an extremely detailed sampling rate in production because profiling has overhead.

## Summary

```text
inuse_space
    = approximate live bytes

inuse_objects
    = approximate live object count

alloc_space
    = cumulative allocated bytes

alloc_objects
    = cumulative allocated object count
```

The most important mental model:

```text
Allocation
    ↓
may become garbage
    ↓
GC
    ↓
memory becomes reusable

Retention
    ↓
object stays reachable
    ↓
memory remains live
```

Use profiles to find where these behaviors originate.

Production workflow:

```text
Observe
  ↓
Profile
  ↓
Understand
  ↓
Hypothesize
  ↓
Optimize
  ↓
Measure again
```

## Next Chapter

**Chapter 09 — pprof**

Next, move from generating and understanding memory profiles to using Go's `pprof` tooling for a running service, including HTTP profiling and practical investigation workflows.
