# Map Memory

## Goal

Understand how Go maps use memory in production, especially what happens when a large map is emptied and reused.

## What You Should Learn

- Understand that a map stores entries in runtime-managed storage.
- Distinguish map length from the memory held by the map.
- Understand why deleting entries does not mean the map immediately returns all allocated storage.
- Know when `clear(m)` is useful for reusing a map.
- Know when replacing or dropping a map reference can make an old map eligible for garbage collection.
- Measure map behavior instead of guessing.

## Production Problem

A service keeps a large in-memory cache:

```go
cache := make(map[int]int)
```

The cache grows to millions of entries. Later, it is invalidated:

```go
clear(cache)
```

Now:

```go
len(cache) == 0
```

but that does not mean the map's memory footprint instantly becomes the same as a newly created empty map.

This matters when a long-lived process repeatedly grows a large map, empties it, and expects memory usage to immediately fall.

## Concept

A useful mental model is:

```text
map
 |
 +-- map metadata
 |
 +-- runtime-managed storage
 |
 +-- keys / values
```

`len(m)` tells you how many entries are currently present. It does not tell you how much storage the map has allocated internally.

Deleting entries removes them from the map. It does not imply that all storage previously needed by the map is immediately returned to the operating system.

`clear(m)` removes all entries and leaves the map usable:

```go
clear(m)
m[1] = 100
```

This is useful when the map is expected to be reused.

If the map is no longer needed, dropping the reference can be a better lifetime decision:

```go
cache = nil
```

If no other references remain, the old map becomes eligible for garbage collection.

## Example

A long-lived cache might look like:

```go
type Service struct {
    cache map[string]Item
}
```

If it is periodically invalidated:

```go
clear(s.cache)
```

the map remains available for reuse.

If the previous size was a temporary spike and the map is unlikely to grow again soon, replacing or dropping the reference may be more appropriate:

```go
s.cache = nil
```

or:

```go
s.cache = make(map[string]Item)
```

Choose based on the expected lifetime and reuse pattern, then measure.

## Experiment

Run:

```bash
go run .
```

The program:

1. Builds a map with one million entries.
2. Measures heap statistics.
3. Calls `clear(cache)`.
4. Measures again.
5. Drops the map reference with `cache = nil`.
6. Measures again after GC.

Exact numbers depend on the Go version, architecture, allocator, and machine.

Focus on the relationship between:

```go
len(cache)
```

and the heap statistics.

An empty map does not necessarily mean zero map-related memory.

## Benchmark

Run:

```bash
go test -bench=. -benchmem
```

The benchmarks compare:

- map lookup
- map construction
- deleting every entry
- `clear`

These benchmarks measure operation cost. They are not a complete production memory-retention test.

For a real memory investigation, use heap profiling and runtime memory statistics.

## What To Observe

### 1. `len(m)` is not a memory metric

After:

```go
clear(m)
```

this is true:

```go
len(m) == 0
```

But the map may still have storage available for future inserts.

### 2. `delete` and `clear` are about entries

Removing entries makes their storage available for reuse. It does not mean the entire map is reconstructed as a tiny empty map.

### 3. Replacing the map changes lifetime

Compare:

```go
clear(cache)
```

with:

```go
cache = nil
```

The second operation drops the reference held by `cache`. If no other references exist, the old map becomes eligible for GC.

### 4. GC is not the same as RSS dropping

Even after GC can reclaim objects, the Go runtime may retain memory for future allocations rather than immediately returning every page to the OS.

Therefore, do not use one heap number or one RSS sample as proof of a leak.

## Production Implication

Use this decision process:

```text
Does the map need to be reused soon?
        |
       yes
        |
      clear(m)
        |
        v
Reuse the existing map

        no
        |
        v
Drop or replace the reference
        |
        v
Old map becomes eligible for GC
```

For long-lived caches, queues, aggregators, indexes, and temporary maps, map lifetime can matter as much as the number of entries.

If memory usage is unexpectedly high:

1. Confirm the symptom with metrics.
2. Capture a heap profile.
3. Identify which objects are retaining memory.
4. Check whether a long-lived map grew much larger than its normal working set.
5. Decide whether reuse or release is appropriate.
6. Measure again.

Do not rewrite every `delete` loop into `clear`, and do not replace every map with `nil`. The correct choice depends on the map's lifecycle and expected reuse.

## Common Mistakes

### Mistake 1: "`len(map) == 0` means the map uses zero memory"

False.

An empty map can still have runtime-managed storage.

### Mistake 2: "`delete` returns all map memory"

Not necessarily.

Deleting entries removes the entries from the map, but the map's storage can remain available for reuse.

### Mistake 3: "`clear` always reduces memory"

`clear` removes all entries. It is primarily useful for emptying a map while keeping it usable. Do not assume it will make process memory immediately fall.

### Mistake 4: "Setting a map to nil immediately reduces RSS"

No.

It makes the old map unreachable if no other references exist. GC and the runtime then determine when that memory is reclaimed or returned to the OS.

### Mistake 5: Optimizing without measuring

A large map is not automatically a problem.

First establish memory growth, allocation rate, GC behavior, object retention, request traffic, and the map's expected size.

## Commands

Run the experiment:

```bash
go run .
```

Run tests:

```bash
go test ./...
```

Run benchmarks:

```bash
go test -bench=. -benchmem
```

Build:

```bash
go build .
```

## Go 1.27 Note

This chapter targets Go 1.27.

Exact benchmark and memory numbers are version- and workload-dependent. Treat benchmark output as a measurement from your machine, not a universal constant.

## Summary

- A map's length and memory footprint are different concepts.
- Removing entries does not mean all previously allocated map storage immediately disappears.
- `clear(m)` is useful when an existing map should be emptied and reused.
- Dropping the last reference can make an old map eligible for garbage collection.
- GC eligibility does not guarantee an immediate reduction in process RSS.
- Production memory tuning should be driven by measurements and heap profiles.

## Next Chapter

Next: **Garbage Collector** — understand how Go decides when to run GC, how allocation rate affects GC work, and which runtime controls matter in production.
