# 06 — Memory Retention

## Goal

Understand why memory can remain live in a Go process even after the application no longer needs the data, when a long-lived reference still keeps that data reachable.

## What You Should Learn

- Distinguish garbage from retained memory.
- Recognize unbounded caches as a production memory-retention risk.
- Understand why GC cannot reclaim reachable objects.
- Design explicit eviction/release points for long-lived data.
- Use heap measurements to validate whether a retention fix worked.

## Production Problem

A service keeps a cache in memory for fast access. Entries are added continuously, but old entries are never evicted.

The objects are not garbage: the cache still references them. As the cache grows, the live heap grows too. Eventually the process can hit its container memory limit or spend more time under memory pressure.

This is different from the slice backing-array problem in Chapter 03. Here, the core problem is the application's **long-lived reference** itself.

## Concept

GC only reclaims objects that are no longer reachable.

```text
root
  |
  v
cache
  |
  +--> entry A --> []byte
  +--> entry B --> []byte
  +--> entry C --> []byte
```

As long as `cache` remains reachable and its entries remain inside the map, the payloads remain reachable.

Therefore:

```text
reachable + no longer useful
        =
memory retention
```

The GC cannot infer that a cache entry is semantically obsolete. The application must remove it.

## Example

This is dangerous when the number of entries has no effective bound:

```go
type Cache struct {
    items map[string][]byte
}

func (c *Cache) Put(key string, value []byte) {
    c.items[key] = value
}
```

If `Put` runs continuously and there is no eviction policy, memory usage can grow with traffic.

A release point can be as simple as:

```go
clear(c.items)
```

For real caches, a bounded-size or TTL-based eviction policy is usually more useful than clearing everything at once.

## Experiment

Run:

```bash
go run .
```

The program creates 2,000 cache entries with a 32 KiB payload each, then forces GC to make the retention visible.

The important sequence is:

```text
populate
    ↓
cache still references every payload
    ↓
GC
    ↓
payloads remain live

clear(cache)
    ↓
entries become unreachable
    ↓
GC
    ↓
payloads become reclaimable
```

The forced `runtime.GC()` calls are for this experiment only. Do not put manual GC calls into normal request handling as a memory-management strategy.

## Benchmark

Run:

```bash
go test -bench=. -benchmem
```

The benchmark measures cache population and clearing. Do not interpret the benchmark as proof that a particular cache implementation is production-ready.

The more important production question is:

> How much data is retained, for how long, and under what workload?

## What To Observe

When running the experiment, pay attention to:

- `HeapAlloc`: current allocated Go heap memory.
- `HeapObjects`: current heap object count.
- `NumGC`: number of completed garbage collections.
- `len(cache.items)`: the application's logical retained data.

A forced GC does not reclaim entries that are still reachable from the cache.

After `clear(cache)` and another GC, the payloads are no longer reachable through that map and can be reclaimed.

## Production Implication

Common sources of memory retention include:

- Unbounded in-memory caches.
- Global registries that only grow.
- Maps keyed by user/session/request IDs without cleanup.
- Long-lived queues or buffers.
- Metrics/debug structures that accidentally keep request data.
- Background workers retaining references longer than intended.

The fix is usually **lifecycle management**, not GC tuning:

```text
bad:
request → store data forever

better:
request → store data → TTL/size bound → evict
```

For a cache, define a policy such as:

- Maximum number of entries.
- Maximum total bytes.
- TTL / expiration.
- LRU-style eviction.
- Explicit invalidation when data becomes obsolete.

The exact policy depends on the application's correctness and latency requirements.

## Common Mistakes

### 1. "GC should clean it up."

Only unreachable objects can be reclaimed. A cache entry is reachable until the application removes it.

### 2. "HeapAlloc is high, so GC is broken."

First determine whether the memory is live and reachable. A high live heap can be correct if the application intentionally keeps that data.

### 3. "Just call runtime.GC()."

Manual GC can help experiments and diagnostics, but it does not solve an application-level retention problem.

### 4. "Set GOGC lower."

GC tuning cannot reclaim reachable objects. If a cache retains 2 GB of live data, changing GOGC does not make those entries collectible.

### 5. "Delete old data eventually."

"Eventually" needs an actual lifecycle policy. Production systems should define when data expires or is evicted.

## Commands

Run the experiment:

```bash
go run .
```

Run benchmarks:

```bash
go test -bench=. -benchmem
```

Build:

```bash
go build .
```

Inspect compiler diagnostics when needed:

```bash
go build -gcflags=-m=2 .
```

For a real service, use heap profiling to find where live memory is coming from. Go's diagnostics documentation describes heap profiling and runtime memory statistics.

## Production Debugging Workflow

```text
1. Confirm memory growth with runtime/process metrics
                    ↓
2. Capture a heap profile
                    ↓
3. Look at in-use memory
                    ↓
4. Identify the retaining data structure
                    ↓
5. Check whether the data should still be live
                    ↓
6. Add eviction/release/lifecycle management
                    ↓
7. Profile again
```

A heap profile is especially useful because it can show which parts of the program account for live memory. Go's profiling documentation distinguishes current in-use memory from cumulative allocation history.

## Summary

- Garbage is unreachable heap data.
- Retained memory is still reachable, even if the application no longer needs it.
- GC cannot reclaim reachable objects.
- Unbounded caches are a common production retention problem.
- Fix retention with explicit lifecycle rules: eviction, TTL, bounds, or invalidation.
- Measure with heap statistics and heap profiles before and after the fix.
- GC tuning is not a substitute for removing unnecessary references.

## Next Chapter

**07 — Benchmark**
