# Chapter 12 — Production Case Study

## Goal

Combine the memory and performance concepts from Chapters 01–11 into one production-style investigation.

The goal is not to memorize commands. The goal is to reason from a production symptom to a measurable cause, choose the smallest useful experiment, and validate the result.

## What You Should Learn

By the end of this chapter, you should be able to:

- distinguish allocation traffic from live memory
- recognize allocation churn
- identify memory retention caused by long-lived references
- use benchmarks to compare competing implementations
- use `pprof` to locate expensive code paths
- reason about GC behavior instead of blindly tuning it
- understand where `GOGC` and `GOMEMLIMIT` fit
- understand where PGO fits
- decide when an optimization is worth its complexity
- validate an optimization with measurements

## Production Problem

Imagine a Go API that generates reports.

The service has a symptom:

- CPU usage periodically increases
- request latency becomes less predictable under load
- memory usage grows during bursts
- restarting the process temporarily reduces memory
- the team suspects "the GC is too slow"

Do not immediately change `GOGC`.

The first question is:

> What is the application actually allocating and retaining?

## Case Study

This chapter contains a deliberately imperfect service.

The `/report` endpoint:

1. creates a large temporary report buffer
2. creates a small result from that buffer
3. stores the result in a long-lived cache
4. repeatedly performs the operation

The implementation contains two different memory problems:

1. allocation churn from repeatedly creating large temporary buffers
2. memory retention caused by cached slices sharing large backing arrays

Your job is to use measurements to separate these problems.

## Mental Model

Think about memory as three different questions:

```text
How much do we allocate?
        ↓
allocation traffic

How much remains reachable?
        ↓
live / retained memory

How much work does GC perform?
        ↓
runtime overhead
```

These are related, but they are not the same metric.

## Run the Service

```bash
go run .
```

The service listens on:

```text
http://localhost:8080
```

Try:

```bash
curl http://localhost:8080/report
```

The endpoint intentionally performs repeated report generation.

## Benchmark

Run:

```bash
go test -bench=. -benchmem
```

Look at:

- `ns/op`
- `B/op`
- `allocs/op`

Do not interpret `B/op` as "RAM permanently added per request."

It describes allocation volume attributed to benchmark operations.

## Memory Profile

Capture a memory profile:

```bash
go test -bench=BenchmarkReport -benchmem -memprofile=mem.out
```

Then inspect allocation history:

```bash
go tool pprof -sample_index=alloc_space -top mem.out
```

Inspect live memory:

```bash
go tool pprof -sample_index=inuse_space -top mem.out
```

The questions are different:

```text
alloc_space
    "Where did allocated bytes go over the profiling period?"

inuse_space
    "Where is live memory represented in this profile?"
```

## Source-Level Investigation

Open the interactive profile:

```bash
go tool pprof -http=:0 mem.out
```

Use the UI to inspect:

- the call graph
- source lines
- allocation attribution

You should be able to trace:

```text
reportHandler
    ↓
generateReports
    ↓
buildReport
    ↓
make(...)
```

The important skill is not finding a large number.

The important skill is connecting that number to the source line that created it.

## Retention Investigation

The cache intentionally keeps small slices.

A slice is a descriptor pointing to an underlying backing array.

If a small slice points into a large backing array:

```text
small slice
   │
   └──────────────► 10 MiB backing array
```

keeping the small slice alive can keep the whole backing array reachable.

That means:

```text
small logical value
        ↓
large physical allocation retained
```

The correct question is:

> Does this cached value need to retain the entire backing array?

If not, consider copying the required bytes into a right-sized allocation.

Do not blindly copy everything.

Copying has a cost:

- another allocation
- memory copy work
- additional allocation traffic

The optimization is justified when the reduction in retention matters.

## Experiment

The package contains two cache paths:

- a retaining implementation
- a bounded/copying implementation

Compare their behavior with the same workload.

Measure:

```text
allocation traffic
live memory
latency
GC behavior
```

Do not optimize based on one metric.

## GC Investigation

The service exposes a diagnostic endpoint:

```text
/debug/stats
```

Try:

```bash
curl http://localhost:8080/debug/stats
```

It reports selected runtime statistics.

Run the workload and observe:

- heap allocation
- heap objects
- number of GC cycles
- GC CPU fraction

The purpose is to connect application allocation behavior with runtime behavior.

## GOGC Experiment

Run the service with different values:

```bash
GOGC=50 go run .
```

```bash
GOGC=100 go run .
```

```bash
GOGC=200 go run .
```

Do not conclude that one value is universally better.

Ask:

- Did heap growth change?
- Did GC frequency change?
- Did latency change?
- Did CPU change?

GC tuning is a trade-off.

## GOMEMLIMIT

You can also experiment with:

```bash
GOMEMLIMIT=256MiB go run .
```

Remember:

`GOMEMLIMIT` is a soft runtime memory target. It is not a replacement for understanding why the application allocates or retains memory.

In a container, leave headroom for:

- goroutine stacks
- runtime structures
- executable/code memory
- libraries
- OS/kernel accounting
- non-Go allocations where applicable

## Benchmark Before and After

For a focused optimization, create two implementations and benchmark them under the same workload.

Keep these constant:

- input size
- number of operations
- machine
- Go version
- benchmark command
- relevant environment variables

Then compare repeated runs.

For modern Go, prefer:

```go
for b.Loop() {
    // measured operation
}
```

For noisy measurements, repeat:

```bash
go test -bench=BenchmarkCache -benchmem -count=10
```

Then use `benchstat` if available.

## PGO Decision

PGO belongs later in the optimization process.

A useful flow is:

```text
Production symptom
       ↓
Measure
       ↓
Profile
       ↓
Find hot path
       ↓
Understand cause
       ↓
Fix allocation / retention / algorithm
       ↓
Benchmark
       ↓
Production validation
       ↓
PGO if the remaining CPU hot paths justify it
```

PGO should not be the first response to high memory usage.

PGO primarily guides compiler optimization using representative workload profiles.

Memory retention is a different problem.

## Production Decision Framework

When investigating a memory/performance issue, ask:

### 1. Is the problem allocation traffic?

Look at:

```text
alloc_space
B/op
allocs/op
```

Potential actions:

- reduce unnecessary allocations
- reuse carefully where ownership/lifetime allow it
- reduce temporary object creation

### 2. Is the problem retained memory?

Look at:

```text
inuse_space
inuse_objects
heap profiles
long-lived references
cache sizes
```

Potential actions:

- bound caches
- add expiration where business semantics allow it
- delete entries
- reduce object lifetime
- copy small data out of oversized backing arrays when justified

### 3. Is GC consuming too much CPU?

Only after understanding allocation behavior, inspect:

- GC frequency
- GC CPU fraction
- heap growth
- latency
- allocation rate

Then consider runtime tuning.

### 4. Is CPU still the bottleneck?

Profile CPU.

If a representative profile shows stable hot paths and the workload justifies it, PGO may be useful.

## Common Mistakes

### Mistake 1 — Treating every allocation as a leak

Allocation is normal.

A leak-like problem is about memory remaining reachable longer than intended.

### Mistake 2 — Treating `B/op` as current RAM

`B/op` measures allocation volume attributed to benchmark operations.

It is not a live-memory measurement.

### Mistake 3 — Tuning GOGC first

If an unbounded cache is retaining hundreds of megabytes, changing GOGC does not fix the cache's lifecycle.

### Mistake 4 — Copying every slice

Copying can fix retention, but it also creates allocation and copy overhead.

Use it when the lifetime/retention trade-off justifies it.

### Mistake 5 — Trusting one benchmark run

Performance measurements contain noise.

Repeat measurements and compare distributions or summarized results.

### Mistake 6 — Assuming PGO must improve performance

PGO can help, but the effect depends on workload, profile quality, compiler decisions, and the remaining bottleneck.

### Mistake 7 — Optimizing without a production symptom

Optimization should answer a real question.

If there is no measurable bottleneck, complexity may be more expensive than the optimization.

## Commands

Basic:

```bash
go run .
curl http://localhost:8080/report
curl http://localhost:8080/debug/stats
```

Benchmark:

```bash
go test -bench=. -benchmem
go test -bench=. -benchmem -count=10
```

Memory profile:

```bash
go test -bench=BenchmarkReport -benchmem -memprofile=mem.out
go tool pprof -sample_index=alloc_space -top mem.out
go tool pprof -sample_index=inuse_space -top mem.out
go tool pprof -http=:0 mem.out
```

Compiler inspection:

```bash
go build -gcflags="-m=2" .
```

Runtime experiments:

```bash
GOGC=50 go run .
GOGC=100 go run .
GOGC=200 go run .
GOMEMLIMIT=256MiB go run .
```

## What to Observe

The most important observations are relationships:

```text
More allocation traffic
        ↓
potentially more GC work

More retained memory
        ↓
larger live heap
        ↓
different GC behavior

Lower GOGC
        ↓
generally less heap growth
        ↓
potentially more GC work

Higher GOGC
        ↓
generally more heap growth
        ↓
potentially less GC work

PGO
        ↓
compiler optimization guided by profile
```

Do not turn these into universal rules. Workload matters.

## Production Implication

A production engineer should be able to say:

> "The service is slow."

and then refine it into something measurable:

> "The service allocates heavily on this hot path."

or:

> "A long-lived cache retains large backing arrays."

or:

> "GC CPU increased because allocation rate increased."

or:

> "The remaining CPU hotspot is stable and representative, so PGO is worth testing."

That transition—from symptom to measurable mechanism—is the main skill this chapter is training.

## Summary

The complete memory-performance workflow is:

```text
Symptom
   ↓
Measure
   ↓
Benchmark
   ↓
Profile
   ↓
Identify allocation / retention / CPU cost
   ↓
Form a hypothesis
   ↓
Make the smallest justified change
   ↓
Benchmark again
   ↓
Validate under realistic workload
   ↓
Deploy carefully
   ↓
Observe production
```

The key lesson is not:

> "Use GOGC."

or:

> "Use PGO."

It is:

> **Understand the mechanism first, then choose the smallest tool that addresses the actual bottleneck.**

## Next Chapter

This chapter completes the `03-memory-performance` learning path.

Next, move to:

```text
04-production-go
```

There the focus shifts from memory/performance internals to building and operating production-grade Go services.
