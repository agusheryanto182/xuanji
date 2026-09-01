# Worker Pool

This playground demonstrates the worker pool pattern in Go using goroutines, channels, and `sync.WaitGroup`.

## Goal

Process many jobs with a limited number of concurrent workers.

```text
                  Producer
                     │
                     ↓
             ┌──────────────┐
             │ Job Channel  │
             │ 1 2 3 4 5 ...│
             └──────┬───────┘
                    │
          ┌─────────┼─────────┐
          ↓         ↓         ↓
       Worker 1  Worker 2  Worker 3
```

The goal is not to create one goroutine per job, but to limit the number of jobs being processed concurrently.

## Structure

```text
06-worker-pool/
├── README.md
├── go.mod
└── main.go
```

## Setup

```bash
go mod init worker-pool-playground
```

## Practice

Create `main.go`:

```go
package main

import (
    "fmt"
    "sync"
    "time"
)

func worker(id int, jobs <-chan int, wg *sync.WaitGroup) {
    defer wg.Done()

    for job := range jobs {
        fmt.Printf("worker %d processing job %d\n", id, job)
        time.Sleep(500 * time.Millisecond)
    }
}

func main() {
    jobs := make(chan int, 5)

    var wg sync.WaitGroup

    for i := 1; i <= 3; i++ {
        wg.Add(1)
        go worker(i, jobs, &wg)
    }

    for i := 1; i <= 10; i++ {
        jobs <- i
    }

    close(jobs)

    wg.Wait()

    fmt.Println("all jobs completed")
}
```

Run:

```bash
go run .
```

The output order is not guaranteed because workers run concurrently.

## How It Works

There are:

```text
10 jobs
3 workers
```

Jobs are sent to the shared channel:

```go
jobs <- job
```

Workers consume jobs:

```go
for job := range jobs {
    // process job
}
```

Architecture:

```text
             10 Jobs
                │
                ↓
        ┌───────────────┐
        │ jobs channel  │
        └───────┬───────┘
                │
       ┌────────┼────────┐
       ↓        ↓        ↓
      W1       W2       W3
```

Each job is received and processed by one worker.

## Why Use a Worker Pool?

Suppose there are:

```text
10,000 jobs
```

Instead of:

```text
10,000 jobs
      ↓
10,000 goroutines
```

we can use:

```text
10,000 jobs
      ↓
   Job Queue
      ↓
10 workers
```

This explicitly limits concurrency.

For example:

```text
10000 jobs
5 workers
```

means at most 5 jobs are being processed concurrently by this worker pool.

## Why Use a Channel?

The channel acts as the job queue:

```go
jobs := make(chan int, 5)
```

Producer:

```go
jobs <- job
```

Worker:

```go
job := <-jobs
```

Multiple workers can receive from the same channel.

## Why Is the Channel Buffered?

This example uses:

```go
jobs := make(chan int, 5)
```

The buffer can temporarily hold up to 5 jobs.

A worker pool does not require a buffered channel. This is also valid:

```go
jobs := make(chan int)
```

The difference is that an unbuffered channel requires sender and receiver synchronization for each send, while a buffered channel can queue values while capacity remains available.

## Why `close(jobs)`?

After all jobs have been sent:

```go
for i := 1; i <= 10; i++ {
    jobs <- i
}

close(jobs)
```

`close(jobs)` tells workers that no more jobs will be sent.

Workers using:

```go
for job := range jobs {
    // process job
}
```

finish existing jobs and stop when the channel is closed and empty.

## Why Use `WaitGroup`?

`wg.Wait()` ensures that `main` waits for all workers to finish.

```text
Workers
   ↓
wg.Done()
   ↓
wg.Wait()
   ↓
main continues
```

Without `WaitGroup`, `main` could return while workers are still processing.

Remember:

```text
select
→ waits for channel operations

WaitGroup
→ waits for goroutines to finish
```

## Worker Count Experiment

Try:

```go
for i := 1; i <= 1; i++ {
```

Then:

```bash
time go run .
```

Try again with:

```go
for i := 1; i <= 5; i++ {
```

Compare the execution time.

Each job sleeps for:

```go
time.Sleep(500 * time.Millisecond)
```

More workers allow more jobs to be processed concurrently.

## Key Takeaway

A worker pool combines:

```text
Goroutines
    +
Channel
    +
WaitGroup
```

to process many jobs with controlled concurrency.

```text
Producer
    ↓
Job Channel
    ↓
Limited Workers
    ↓
Processed Jobs
```

The important idea is:

> A worker pool controls how many jobs can execute concurrently instead of creating an unlimited number of workers.
