# 01 - Concurrency

This section covers Go concurrency fundamentals through isolated hands-on playgrounds.

The goal is to understand how goroutines communicate, how shared state can become unsafe, how to control concurrency, and how to manage goroutine lifecycles.

## Topics

```text
01-concurrency/
├── 01-goroutine/
├── 02-race-condition/
├── 03-mutex/
├── 04-channel/
├── 05-select/
├── 06-worker-pool/
└── 07-goroutine-leak/
```

---

## 01. Goroutine

A goroutine is a lightweight concurrent function execution.

Basic example:

```go
go sayHello()
```

Instead of waiting for `sayHello()` to finish before continuing, Go schedules it to run concurrently.

Mental model:

```text
main
 │
 ├── start goroutine
 │      │
 │      └── work
 │
 └── continue
```

Key idea:

> A goroutine allows a function to execute concurrently with the caller.

---

## 02. Race Condition

A race condition can occur when multiple goroutines access shared data concurrently and at least one of them modifies that data.

Example:

```go
counter++
```

performed by many goroutines can produce an unexpected result.

```text
goroutine 1 ──┐
goroutine 2 ──┤
goroutine 3 ──┼──> shared counter
goroutine 4 ──┤
goroutine 5 ──┘
```

The Go race detector can help find these problems:

```bash
go run -race .
```

Key idea:

> Concurrent access to shared mutable state must be synchronized correctly.

---

## 03. Mutex

A mutex protects shared data from concurrent access.

```go
mu.Lock()
counter++
mu.Unlock()
```

Only one goroutine can hold the lock at a time.

Mental model:

```text
goroutine 1 ──> LOCK ──> modify ──> UNLOCK
                              │
goroutine 2 ───────── WAIT ───┘
```

Use a mutex when multiple goroutines need to safely access shared mutable state.

---

## 04. Channel

Channels allow goroutines to communicate by sending and receiving values.

```go
ch := make(chan int)

ch <- 10

value := <-ch
```

Conceptually:

```text
goroutine A
    │
    │ send
    ▼
 channel
    │
    │ receive
    ▼
goroutine B
```

### Unbuffered Channel

```go
ch := make(chan int)
```

The sender and receiver synchronize directly.

### Buffered Channel

```go
ch := make(chan int, 3)
```

The channel can temporarily store up to 3 values.

```text
producer
   │
   ▼
[ 1 ][ 2 ][ 3 ]
      channel buffer
   │
   ▼
consumer
```

Key idea:

> Channels provide a way for goroutines to communicate and coordinate.

---

## 05. Select

`select` waits for multiple channel operations.

Example:

```go
select {
case value := <-ch1:
    fmt.Println(value)

case value := <-ch2:
    fmt.Println(value)
}
```

If one case becomes ready, `select` executes that case.

If multiple cases are ready, Go chooses among the ready cases pseudo-randomly.

A common use is timeout handling:

```go
select {
case value := <-ch:
    fmt.Println(value)

case <-time.After(time.Second):
    fmt.Println("timeout")
}
```

Key idea:

> `select` is useful when a goroutine needs to wait for multiple possible channel events.

---

## 06. Worker Pool

A worker pool limits the number of goroutines processing jobs concurrently.

Example architecture:

```text
              jobs
               │
               ▼
        ┌──────────────┐
        │ Job Channel  │
        └──────────────┘
          │     │     │
          ▼     ▼     ▼
       worker worker worker
          │     │     │
          └─────┼─────┘
                ▼
             results
```

Instead of creating one goroutine for every job:

```text
job 1 → goroutine
job 2 → goroutine
job 3 → goroutine
...
job 10000 → goroutine
```

a worker pool can use a fixed number of workers:

```text
10000 jobs
    ↓
job queue
    ↓
3 workers
```

This provides controlled concurrency.

Typical implementation:

```go
for i := 0; i < 3; i++ {
    go worker(jobs)
}
```

Key idea:

> A worker pool separates the number of jobs from the number of concurrent workers.

---

## 07. Goroutine Leak

A goroutine leak happens when a goroutine that should have stopped remains alive because it is blocked or has no clear exit condition.

Example:

```go
func worker(jobs <-chan int) {
    for job := range jobs {
        fmt.Println(job)
    }
}
```

If the channel remains open, no new jobs arrive, and nothing closes the channel:

```text
worker
  ↓
waiting for next job
  ↓
BLOCKED
  ↓
never returns
```

If the application is still running, that goroutine remains alive and consumes resources.

A common solution is to define a clear lifecycle using channel closure or context cancellation.

```go
select {
case <-ctx.Done():
    return
}
```

Key idea:

> Every goroutine should have a clear way to stop.

---

# How the Topics Connect

These topics are not isolated concepts. They build on each other.

```text
Goroutine
    │
    ▼
Multiple goroutines
    │
    ▼
Race Condition
    │
    ▼
Mutex
    │
    ├──────────────┐
    ▼              ▼
Channel          Shared State
    │
    ▼
Select
    │
    ▼
Worker Pool
    │
    ▼
Goroutine Lifecycle
    │
    ▼
Goroutine Leak
```

A more practical view:

```text
                GOROUTINES
                     │
          ┌──────────┴──────────┐
          ▼                     ▼
   Shared Memory            Communication
          │                     │
          ▼                     ▼
   Race Condition           Channels
          │                     │
          ▼                     ▼
        Mutex                 Select
                                │
                                ▼
                           Worker Pool
                                │
                                ▼
                        Lifecycle Management
                                │
                                ▼
                         Goroutine Leak
```

---

# Concurrency vs Parallelism

Concurrency and parallelism are related but not identical.

### Concurrency

Multiple tasks make progress during the same period.

```text
Task A ── work ── wait ── work
Task B ─── wait ── work ── wait
```

### Parallelism

Multiple tasks actually execute at the same time on different CPU cores.

```text
CPU 1: Task A ─────────────
CPU 2: Task B ─────────────
```

Go supports concurrency naturally through goroutines and channels. Actual parallel execution depends on available CPU resources and Go's runtime scheduling.

---

# Synchronization Tools

Different concurrency problems call for different tools.

| Problem                                       | Common Tool                 |
| --------------------------------------------- | --------------------------- |
| Run work concurrently                         | Goroutine                   |
| Protect shared mutable state                  | Mutex                       |
| Communicate between goroutines                | Channel                     |
| Wait for multiple channel events              | Select                      |
| Wait for goroutines to finish                 | WaitGroup                   |
| Limit concurrent workers                      | Worker Pool                 |
| Cancel running work                           | Context                     |
| Prevent goroutines from staying alive forever | Context / channel lifecycle |

---

# Important Mental Models

## Goroutine

```text
start
  ↓
work
  ↓
finish
```

## Mutex

```text
LOCK
  ↓
critical section
  ↓
UNLOCK
```

## Channel

```text
producer
   ↓
channel
   ↓
consumer
```

## Select

```text
       select
      /      \
 channel A  channel B
      │        │
      └── ready┘
          ↓
      one case runs
```

## Worker Pool

```text
many jobs
    ↓
job queue
    ↓
limited workers
    ↓
processed jobs
```

## Goroutine Lifecycle

```text
START
  ↓
WORK
  ↓
STOP CONDITION
  ↓
RETURN
```

A goroutine leak occurs when the final part is missing:

```text
START
  ↓
WORK
  ↓
BLOCKED
  ↓
never returns
```

---

# Practice Checklist

Before moving to the next section, make sure you can explain:

- What is a goroutine?
- Why can a race condition happen?
- What does a mutex protect?
- What is the difference between buffered and unbuffered channels?
- Why can a channel operation block?
- What does `select` wait for?
- Why does `select` not automatically wait for all goroutines?
- What problem does a worker pool solve?
- Why is `close(channel)` important when using `range` over a channel?
- What is a goroutine leak?
- Why can a blocked goroutine remain alive while the application is running?
- How can context cancellation help prevent goroutine leaks?

If you can answer these questions and understand the examples, the core concurrency foundation is in place.

---

# Recommended Learning Order

Follow the topics in this order:

```text
01 goroutine
      ↓
02 race condition
      ↓
03 mutex
      ↓
04 channel
      ↓
05 select
      ↓
06 worker pool
      ↓
07 goroutine leak
```

The order is intentional:

- First understand how goroutines run.
- Then understand the problems caused by concurrency.
- Learn how mutexes protect shared state.
- Learn channels for communication.
- Learn `select` for coordinating channel operations.
- Combine channels and goroutines into worker pools.
- Finally, learn how poorly managed goroutines can leak.

---

# Final Takeaway

The most important concept in Go concurrency is not simply:

```go
go something()
```

It is understanding the **lifecycle and communication of concurrent work**.

Always ask:

```text
Who starts this goroutine?
        ↓
What does it do?
        ↓
How does it communicate?
        ↓
What happens if it blocks?
        ↓
What tells it to stop?
        ↓
When does it return?
```

If you can answer those questions, you are thinking about Go concurrency correctly.
