# 07 - Goroutine Leak

## What is a Goroutine Leak?

A **goroutine leak** occurs when a goroutine that is no longer needed remains alive because it has no way to stop.

A leaked goroutine is commonly:

- blocked
- waiting on a channel
- waiting on a context or event
- waiting for something that will never happen
- missing a clear exit condition

As long as the program is running, the leaked goroutine continues to consume resources.

## Example of a Goroutine Leak

```go
package main

import (
	"fmt"
	"time"
)

func worker(jobs <-chan int) {
	fmt.Println("worker started")

	for job := range jobs {
		fmt.Println("processing job:", job)
	}

	fmt.Println("worker stopped")
}

func main() {
	jobs := make(chan int)

	go worker(jobs)

	jobs <- 1
	jobs <- 2
	jobs <- 3

	time.Sleep(1 * time.Second)

	fmt.Println("main still running")

	time.Sleep(10 * time.Second)

	fmt.Println("main finished")
}
```

After job `3` is processed, the worker tries to receive another job through:

```go
for job := range jobs
```

But the channel is still open and no new data is being sent.

The worker therefore becomes:

```text
processing job: 3
        ↓
waiting for the next job
        ↓
BLOCKED
        ↓
still alive
```

The worker never reaches:

```go
fmt.Println("worker stopped")
```

because `jobs` is never closed.

## Why Is It Called a Leak?

In a long-running server, leaked goroutines can accumulate:

```text
request 1 → goroutine → blocked
request 2 → goroutine → blocked
request 3 → goroutine → blocked
...
```

Over time:

```text
1 goroutine
    ↓
10 goroutines
    ↓
100 goroutines
    ↓
1,000 goroutines
    ↓
more resources being consumed
```

This is a goroutine leak.

## Why Can `main()` Still Finish?

Go does **not** automatically wait for all goroutines to finish.

When:

```go
main()
```

returns, the program exits.

Any goroutine that is still blocked also stops because the entire process has exited.

So a goroutine leak does **not** mean that the program can never exit.

The actual problem is:

> While the program is still running, the leaked goroutine remains alive and consumes resources.

This becomes especially important for long-running applications such as servers.

## Open Channel vs Closed Channel

For:

```go
for job := range jobs
```

there are three important conditions.

### 1. Channel is open + has data

```text
channel
   ↓
has job
   ↓
worker receives it
   ↓
process
```

### 2. Channel is open + empty

```text
channel
   ↓
no job
   ↓
worker waits
   ↓
BLOCKED
```

If nothing will ever send another job, this can become a goroutine leak.

### 3. Channel is closed + empty

```text
channel closed
   ↓
no data remaining
   ↓
range ends
   ↓
worker returns
   ↓
goroutine finishes
```

This is why `close(jobs)` can be part of the worker lifecycle.

## Simple Fix

If there are no more jobs:

```go
close(jobs)
```

Example:

```go
func worker(jobs <-chan int) {
	for job := range jobs {
		fmt.Println("processing job:", job)
	}

	fmt.Println("worker stopped")
}

func main() {
	jobs := make(chan int)

	go worker(jobs)

	jobs <- 1
	jobs <- 2
	jobs <- 3

	close(jobs)
}
```

The flow becomes:

```text
job 1
 ↓
job 2
 ↓
job 3
 ↓
close(jobs)
 ↓
range ends
 ↓
worker stopped
 ↓
goroutine finishes
```

## Don't Close a Channel Randomly

The channel should generally be closed by the side responsible for **sending values**, not by the receiver.

The producer knows when there will be no more jobs:

```go
jobs <- 1
jobs <- 2
close(jobs)
```

The worker simply receives:

```go
for job := range jobs {
	// process
}
```

## Goroutine Leaks Are Not Only Caused by Channels

A goroutine can also leak while waiting for something else.

Example:

```go
func worker(ch <-chan int) {
	value := <-ch
	fmt.Println(value)
}
```

If no sender ever sends a value:

```text
worker
  ↓
<-ch
  ↓
BLOCKED
  ↓
waits forever
```

## Context Cancellation

For goroutines that need to run continuously, `context.Context` can provide a clear cancellation mechanism.

```go
func worker(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			fmt.Println("worker stopped")
			return

		default:
			fmt.Println("worker working...")
			time.Sleep(500 * time.Millisecond)
		}
	}
}
```

When:

```go
cancel()
```

is called:

```text
cancel()
   ↓
ctx.Done()
   ↓
worker receives the signal
   ↓
return
   ↓
goroutine finishes
```

## Key Takeaway

> **A goroutine leak is a goroutine that gets "stuck" and cannot, or does not have a chance to, stop when it should.**

Whenever you create a goroutine, ask:

1. Who creates this goroutine?
2. What makes it stop?
3. Can it `return`?
4. Will the channel be closed?
5. Is there a cancellation mechanism?
6. Can it become blocked forever?

If you cannot answer:

> **"When does this goroutine stop?"**

then the goroutine's lifecycle should be reviewed.

## Mental Model

A healthy goroutine:

```text
START
  ↓
WORK
  ↓
STOP CONDITION
  ↓
RETURN
```

A leaked goroutine:

```text
START
  ↓
WORK
  ↓
BLOCKED
  ↓
never returns
```
