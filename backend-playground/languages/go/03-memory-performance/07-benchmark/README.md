# Chapter 07 — Benchmark

## Goal

Learn how to measure Go performance correctly and interpret benchmark results without being fooled by compiler optimizations, setup cost, or allocation artifacts.

This chapter is about **measurement and reasoning**, not memorizing the `testing.B` API.

## What You Should Learn

- Understand `ns/op`, `B/op`, and `allocs/op`.
- Recognize when compiler optimization makes a benchmark misleading.
- Make benchmark results observable with a sink when necessary.
- Keep setup outside the measured operation.
- Use `b.ResetTimer()` when appropriate.
- Use `-benchmem` to inspect allocation behavior.
- Compare implementations using the same workload.
- Repeat important measurements.
- Understand when a microbenchmark is useful and when production profiling is needed.
- Prefer `B.Loop` for new benchmarks on modern Go versions.

## Production Problem

A benchmark might say:

```text
Implementation A: 120 ns/op, 64 B/op, 2 allocs/op
Implementation B:  80 ns/op,  0 B/op, 0 allocs/op
```

That is useful evidence, but it is not automatically proof that B is better in production.

First ask:

- Did both implementations do the same work?
- Was compiler optimization involved?
- Was setup included?
- Is this code actually a production hot path?
- Does the improvement matter at real traffic?
- Did memory, tail latency, or maintainability get worse?

A benchmark is evidence, not a production conclusion.

## Concept

A Go benchmark has the form:

```go
func BenchmarkXxx(b *testing.B)
```

Typical output:

```text
BenchmarkCreateUser-8    1000000000    1.2 ns/op    0 B/op    0 allocs/op
```

- `ns/op`: measured time per operation.
- `B/op`: allocation bytes attributed per operation.
- `allocs/op`: allocation events attributed per operation.

These are benchmark measurements, not total process memory.

## The Classic Bad Benchmark

This can be misleading:

```go
func BenchmarkBad(b *testing.B) {
	for i := 0; i < b.N; i++ {
		buildUsersValue(100)
	}
}
```

The result is discarded. If the compiler can prove the work has no observable effect, it may remove or simplify it.

The lesson is not that the compiler is cheating. The benchmark failed to make the intended work observable.

## Consume the Result

When necessary:

```go
var sinkUsers []User

func BenchmarkGood(b *testing.B) {
	for i := 0; i < b.N; i++ {
		sinkUsers = buildUsersValue(100)
	}
}
```

A sink is a measurement technique. Do not add sinks blindly; use them when the result could otherwise be optimized away.

## Setup vs Measurement

If you want to measure lookup, do not accidentally measure creation of the test data:

```go
func BenchmarkLookup(b *testing.B) {
	users := buildUsersValue(10000)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = users[i%len(users)]
	}
}
```

`ResetTimer` resets elapsed time and allocation counters before the measured section.

For new benchmarks on modern Go, prefer:

```go
func BenchmarkLookup(b *testing.B) {
	users := buildUsersValue(10000)

	for b.Loop() {
		_ = users[0]
	}
}
```

`B.Loop` was added in Go 1.24 and is the preferred style for new benchmarks. The runnable examples in this chapter use `b.N` so they can also be verified with older local toolchains.

## Allocation Measurement

Run:

```bash
go test -bench=. -benchmem
```

For example:

```text
BenchmarkBuildUsersValue-8       ...    3200 ns/op    3072 B/op    1 allocs/op
BenchmarkBuildUsersPointer-8     ...    6000 ns/op    4000 B/op  101 allocs/op
```

This shows allocation behavior for this exact workload. It does not prove that value-based code is always faster than pointer-based code.

## Comparing Implementations

Keep the comparison fair:

```text
same input
same output requirement
same machine
same Go version
same benchmark command
same workload
```

Compare:

```text
ns/op
B/op
allocs/op
```

For important A/B comparisons, repeat the benchmark:

```bash
go test -bench=. -benchmem -count=10
```

Then use `benchstat` from `golang.org/x/perf` to compare repeated results.

## Benchmark vs Profiling

Use a benchmark to answer:

> Is implementation A faster than B for this controlled workload?

Use profiling to answer:

> Where is the running production process actually spending CPU or memory?

A useful workflow is:

```text
Production symptom
       ↓
Profile / measure
       ↓
Find hot path
       ↓
Create focused benchmark
       ↓
Change implementation
       ↓
Benchmark A vs B
       ↓
Validate in production
```

## Common Pitfalls

### Compiler optimization

Unused results can make the benchmark measure less work than intended.

### Setup contamination

Expensive setup can dominate the measured operation.

### Different workloads

A comparison is meaningless if A and B process different amounts of data.

### Logging

Logging inside the benchmark can dominate the operation.

### Non-hot code

A huge microbenchmark improvement may not matter if the code rarely runs.

### Single-run conclusions

Benchmark noise exists. Repeat measurements when the decision matters.

### Optimizing the wrong metric

Lower `ns/op` is not automatically better if the change causes unacceptable memory use, complexity, or tail latency.

## Experiment

Run the HTTP example:

```bash
go run .
```

Then:

```bash
curl 'http://localhost:8080/users?n=1000'
```

Run all benchmarks:

```bash
go test -bench=. -benchmem
```

Run one benchmark:

```bash
go test -bench=BenchmarkBuildUsersValue -benchmem
```

Repeat it:

```bash
go test -bench=BenchmarkBuildUsersValue -benchmem -count=10
```

Inspect compiler optimization decisions:

```bash
go test -gcflags=-m=2
```

## What To Observe

Focus on:

1. `BenchmarkBadUnusedResult`.
2. `BenchmarkGoodConsumedResult`.
3. `ns/op`, `B/op`, and `allocs/op`.
4. The difference between setup and measured work.
5. Value vs pointer under the same workload.
6. Variation across repeated runs.
7. Whether the benchmark resembles the production hot path.

Do not memorize exact numbers. CPU, OS, Go version, compiler decisions, and workload affect them.

## Production Implication

The important question is not:

> How fast is this function?

It is:

> What exactly did I measure, and does the measurement support the production decision?

Microbenchmarks isolate one operation. Production telemetry and profiling show the behavior of the whole service.

## Commands

```bash
go run .
go test -bench=. -benchmem
go test -bench=BenchmarkBuildUsersValue -benchmem
go test -bench=. -benchmem -count=10
go test -gcflags=-m=2
```

For A/B comparison:

```bash
go test -bench=. -benchmem -count=10 > old.txt
go test -bench=. -benchmem -count=10 > new.txt
benchstat old.txt new.txt
```

## Summary

- A benchmark is a controlled measurement, not production reality.
- `ns/op` measures time per operation.
- `B/op` measures allocation bytes per operation.
- `allocs/op` measures allocation events per operation.
- Unused results can be optimized away.
- Keep setup outside the measured operation.
- Use sinks when necessary.
- Prefer `B.Loop` for new benchmarks on modern Go.
- Repeat important measurements and use statistical comparison.
- Use profiling and production telemetry to decide what actually needs optimization.

## Next Chapter

**Chapter 08 — Memory Profiling**

Next we move from:

> How expensive is this operation?

to:

> Which objects are actually consuming memory in my running program?
