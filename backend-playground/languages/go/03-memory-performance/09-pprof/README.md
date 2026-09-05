# Chapter 09 — pprof

## Goal

Learn how to use Go's `pprof` tooling to turn a CPU or memory symptom into a concrete call path and source location.

## What You Should Learn

- What `pprof` represents.
- How profiles become call graphs.
- How to inspect profiles with `top`, `list`, and the web UI.
- How `flat` and `cum` help locate expensive code paths.
- How to connect a profile back to source code.
- Why profiling is sampled evidence.
- How CPU and memory profiling fit into production debugging.

## Production Problem

A service reports high CPU or memory usage, but metrics only tell you **that** something is wrong.

You need to answer:

```text
Which function is consuming CPU?
Which function is allocating memory?
Which caller triggered that work?
Where in the source should I investigate?
```

That is where `pprof` becomes useful.

## Concept

A simplified mental model:

```text
Application
    |
    | profiling samples
    v
Profile data
    |
    v
Call graph
    |
    +--> top
    +--> list
    +--> web UI
```

`pprof` gives you evidence about where resources are attributed. It does not automatically tell you the root cause.

## Example

This chapter uses a deliberately simple report generator:

```go
func buildReport(rows int) []byte {
    report := make([]byte, rows*1024)

    for i := range report {
        report[i] = byte(i)
    }

    return report
}
```

The important question is:

> Does profiling show `buildReport` as an expensive part of the workload?

## Experiment

Run:

```bash
go test -bench=. -benchmem
```

Create a memory profile:

```bash
go test -bench=BenchmarkGenerateReports -benchmem -memprofile=mem.out
```

Inspect it:

```bash
go tool pprof -top mem.out
```

Interactive mode:

```bash
go tool pprof mem.out
```

Useful commands:

```text
top
top -cum
list buildReport
```

Visual call graph:

```bash
go tool pprof -http=:0 mem.out
```

## Reading `top`

Example:

```text
flat    flat%   sum%    cum     cum%
2.38GB  99.90%  99.90%  2.38GB  99.90%  buildReport
0       0%      99.90%  2.38GB  99.90%  BenchmarkGenerateReports
```

- `flat` = value attributed directly to that function.
- `cum` = value attributed to that function plus descendants.
- `sum%` = running percentage of selected flat contributions.

When asking where the direct cost is attributed, start with `flat`.

Do not add `cum` values together with `flat` values.

## `list`

`list` connects profile samples to source lines:

```text
(pprof) list buildReport
```

Conceptually:

```text
ROUTINE ======================== buildReport
      2.38GB      2.38GB   4  func buildReport(rows int) []byte {
                              5      report := make([]byte, rows*1024)
                              6      for i := range report {
                              7          report[i] = byte(i)
                              8      }
                              9      return report
                             10  }
```

This moves the investigation from:

```text
"buildReport is expensive"
```

to:

```text
"the allocation inside buildReport is code I should inspect"
```

## Call Graph Mental Model

Suppose:

```text
HTTP handler
    |
    +--> service
           |
           +--> generateReports
                    |
                    +--> buildReport
```

The leaf may receive the direct attribution while the call path explains why that work happened.

Ask both:

1. What is expensive?
2. Why is this path being called this often?

Sometimes the real problem is caller frequency rather than the leaf function.

## Sampling Matters

Profiles are based on samples, not an exact ledger of every event.

Therefore:

- tiny numerical differences are not automatically meaningful;
- repeatable patterns matter more than tiny changes;
- compare profiles using the same workload when possible.

For memory profiles:

```text
alloc_*  -> allocation history / churn
inuse_*  -> live memory / objects
```

For CPU profiles:

```text
CPU profile -> where CPU time is being sampled
```

## Profiling vs Benchmarking

Benchmark:

```text
Is A faster or more allocation-efficient than B
under a controlled workload?
```

Profiling:

```text
Where is this process spending resources?
```

Production workflow:

```text
Production symptom
       |
       v
Profile
       |
       v
Find expensive path
       |
       v
Inspect source
       |
       v
Form hypothesis
       |
       v
Focused benchmark
       |
       v
Optimize
       |
       v
Profile again
```

## CPU Profiling

Create a CPU profile:

```bash
go test -bench=BenchmarkGenerateReports -cpuprofile=cpu.out
```

Inspect:

```bash
go tool pprof -top cpu.out
```

Or:

```bash
go tool pprof -http=:0 cpu.out
```

The question becomes:

```text
Where is CPU time being spent?
```

Do not assume a memory-heavy function is also CPU-hot.

## Memory Profiling

Create a memory profile:

```bash
go test -bench=BenchmarkGenerateReports -memprofile=mem.out
```

Inspect:

```bash
go tool pprof -top mem.out
```

Change the sample type interactively:

```text
(pprof) sample_index=alloc_space
(pprof) top

(pprof) sample_index=alloc_objects
(pprof) top

(pprof) sample_index=inuse_space
(pprof) top

(pprof) sample_index=inuse_objects
(pprof) top
```

## Production Implication

When a metric tells you **that** something is wrong but not **where**, profiling gives you a path to investigate.

```text
High CPU
    -> CPU profile
    -> find hot call path
    -> inspect source

Memory keeps growing
    -> heap profile
    -> inspect inuse_space / inuse_objects
    -> find retained objects

High allocation rate
    -> allocation profile
    -> inspect alloc_space / alloc_objects
    -> find allocation-heavy path
```

The production skill is:

```text
metric
  -> profile
  -> call path
  -> source line
  -> hypothesis
  -> fix
  -> validation
```

## Common Mistakes

### 1. Treating `alloc_space` as current RAM

`alloc_space` is cumulative allocation represented by the profile period. It is not current process memory.

### 2. Treating every `runtime.*` function as the root bug

Functions such as `runtime.mallocgc` are Go runtime machinery. Investigate which application path caused the work.

### 3. Adding `cum` values together

`cum` includes descendants. Adding multiple cumulative rows can double-count work.

### 4. Optimizing before profiling

Do not guess. Profile when you have a real production symptom.

### 5. Comparing unrelated profiles

Profile results depend on workload, machine, Go version, runtime state, and sampling. Keep comparisons controlled.

## Commands

```bash
go test -bench=. -benchmem

go test -bench=BenchmarkGenerateReports -cpuprofile=cpu.out
go tool pprof -top cpu.out

go test -bench=BenchmarkGenerateReports -memprofile=mem.out
go tool pprof -top mem.out

go tool pprof mem.out
go tool pprof -http=:0 mem.out
```

Interactive:

```text
top
top -cum
list buildReport
sample_index=alloc_space
sample_index=alloc_objects
sample_index=inuse_space
sample_index=inuse_objects
```

## Summary

The core mental model:

```text
pprof
  |
  +--> shows WHERE resource usage is attributed
  +--> shows call paths
  +--> connects profiles to source
  +--> helps form production hypotheses
```

Remember:

1. Profile first when you have a real symptom.
2. Use `flat` for direct attribution.
3. Use `cum` to understand the call path.
4. Use `list` to connect the profile to source.
5. Use CPU profiles for CPU questions.
6. Use memory profiles for allocation/live-memory questions.
7. Treat profiles as sampled evidence, not an exact event ledger.
8. After finding a hot path, benchmark the proposed optimization.

## Next Chapter

**Chapter 10 — Production Memory Tuning**

Next we turn profiling evidence into production decisions around allocation rate, GC pressure, heap growth, memory limits, `GOMEMLIMIT`, GOGC trade-offs, bounded caches, and object lifetime.
