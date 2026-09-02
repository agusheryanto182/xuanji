# Context Worker Pool: Graceful Shutdown and Channel Blocking

## Overview

This material explains how `context.Context`, channels, workers, and graceful shutdown work together in a Go worker pool.

The main focus is this pattern:

```go
select {
case <-ctx.Done():
    return
case jobs <- i:
    fmt.Println("submitted job", i)
}
```

It also explains why an unbuffered channel can temporarily block the producer.

## 1. Unbuffered Channel

An unbuffered channel has no internal queue:

```go
jobs := make(chan int)
```

When the producer executes:

```go
jobs <- 1
```

the send cannot complete until another goroutine receives the value:

```go
job := <-jobs
```

Conceptually:

```text
Producer                    Worker
   |                           |
   |---- send job 1 ---------->|
   |                           |
   |   waits for receiver      |
   |                           |
   +-- continues after receive +-- processes job
```

If no worker is ready to receive job `1`, the producer blocks at `jobs <- 1`.

It cannot continue to `jobs <- 2` until job `1` has been received.

## 2. `select` and Blocking

Consider:

```go
select {
case <-ctx.Done():
    fmt.Println("producer stopped:", ctx.Err())
    return

case jobs <- i:
    fmt.Println("submitted job", i)
}
```

The producer waits until one of the cases is ready.

If a worker is ready, the send succeeds.

If no worker is ready, the send case cannot complete, so the producer waits.

However, if the context becomes canceled while it is waiting, the `ctx.Done()` case becomes ready and the producer can stop.

Conceptually:

```text
                 +-- worker receives job --> send succeeds
select waits ----|
                 +-- context becomes done -> stop
```

This is why channel operations and context cancellation are often combined with `select`.

## 3. Unbuffered vs Buffered

### Unbuffered

```go
jobs := make(chan int)
```

A send requires a receiver.

```text
Producer -> Worker
          direct handoff
```

### Buffered

```go
jobs := make(chan int, 3)
```

The channel can temporarily store three jobs.

```text
Producer
   |
   +-- job 1 -> [1]
   +-- job 2 -> [1 2]
   +-- job 3 -> [1 2 3]
                    |
                    v
                  Worker
```

The producer does not need a worker to receive every job immediately while the buffer still has capacity.

Once the buffer is full, another send blocks until a worker receives something.

## 4. Why Use an Unbuffered Channel Here?

Using an unbuffered channel makes synchronization behavior obvious.

The producer submits a job only when a worker is ready to receive it.

This is useful for learning:

- blocking sends
- blocking receives
- producer/consumer synchronization
- `select`
- context cancellation

It is **not** a requirement for graceful shutdown.

A real worker pool may use a buffered channel when it needs a queue.

## 5. Graceful Shutdown

Graceful shutdown generally means:

```text
shutdown signal
      |
      v
stop accepting new work
      |
      v
finish existing/in-flight work
      |
      v
workers exit
      |
      v
WaitGroup completes
      |
      v
process exits
```

It does not normally mean killing every worker immediately.

For a pure graceful worker pool, the usual idea is to stop the producer and close the jobs channel:

```go
close(jobs)
```

Workers can then finish jobs that are already available and exit after the channel is drained.

## 6. Context Cancellation vs Graceful Shutdown

These concepts are related but different.

### Context cancellation

```go
case <-ctx.Done():
    return
```

means:

> Stop this operation because cancellation has been requested.

A worker may abandon work that is currently in progress if its code is designed to respond to cancellation.

### Graceful shutdown

Means:

> Stop accepting new work, then allow appropriate existing work to finish before exiting.

A shutdown deadline can still be used as a safety mechanism.

```text
graceful shutdown starts
        |
        v
workers drain existing jobs
        |
        v
deadline reached?
   +----+----+
  no        yes
   |          |
finish     cancel remaining work
```

## 7. Three Important Tools

### `close(jobs)`

Means:

> No more jobs will be sent.

It describes the lifecycle of the job channel.

### `ctx.Done()`

Means:

> Cancellation has been requested.

It describes cancellation of work.

### `sync.WaitGroup`

Means:

> Wait until all worker goroutines have exited.

It is synchronization, not cancellation.

## 8. Mental Model

Think of an unbuffered channel as a direct handoff:

```text
        "I have job 1"
Producer ----------------> Worker
        "I'll wait until
         someone takes it"
```

Think of a buffered channel as a queue:

```text
Producer -> [ job 1 | job 2 | job 3 ] -> Workers
```

Think of `select` as waiting for whichever event becomes possible:

```text
                 +-- job can be sent --> send
select waits ----|
                 +-- context done -----> stop
```

## 9. Key Takeaways

1. An unbuffered channel has no queue.
2. `jobs <- i` blocks until a receiver accepts `i`.
3. Because the send is inside `select`, context cancellation provides an alternative path.
4. A buffered channel allows jobs to wait in the channel until workers consume them.
5. Unbuffered channels are useful for demonstrating synchronization, but they are not required for graceful shutdown.
6. `close(jobs)` means no more jobs will be submitted.
7. `ctx.Done()` represents cancellation.
8. `WaitGroup` waits for goroutines to finish.
9. Graceful shutdown usually means finishing appropriate existing work before exiting.
10. A shutdown timeout can be used as a safety boundary when graceful shutdown takes too long.
