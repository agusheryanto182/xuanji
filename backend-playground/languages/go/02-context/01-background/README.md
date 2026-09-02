# 01 - Context Background

## What is `context.Background()`?

`context.Background()` is a **root context** used as a starting point for creating derived contexts.

```go
ctx := context.Background()
```

By itself, it does not provide cancellation, timeout, or deadline.

Think of it as the starting point of a context tree:

```text
context.Background()
        │
        ├── WithCancel()
        ├── WithTimeout()
        └── WithDeadline()
```

## Is `Background()` Automatically Created?

No.

Go does not automatically create one `Background()` context for every program or context tree.

You explicitly create it:

```go
root := context.Background()

ctx, cancel := context.WithCancel(root)
```

## Context Tree

Contexts commonly form a parent-child tree.

```go
root := context.Background()

ctx1, cancel1 := context.WithCancel(root)
ctx2, cancel2 := context.WithTimeout(ctx1, 5*time.Second)
```

The relationship is:

```text
Background
    │
    ▼
   ctx1
    │
    ▼
   ctx2
```

Each derived context has one parent, while one parent can have multiple children:

```text
             Background
                 │
                 ▼
                ctx1
             /   |               /    |             ctx2   ctx3   ctx4
```

## Does Every Context Have a `Background()`?

No.

It is more accurate to say:

> `Background()` is one possible root context used to start a context tree.

For example:

```go
root := context.Background()

ctx1, cancel1 := context.WithCancel(root)
ctx2, cancel2 := context.WithTimeout(ctx1, 5*time.Second)
```

Here `ctx2` has `ctx1` as its parent, not `Background()` directly.

```text
Background
    │
    ▼
   ctx1
    │
    ▼
   ctx2
```

## Can There Be Multiple `Background()` Contexts?

Yes.

```go
ctx1 := context.Background()
ctx2 := context.Background()
```

This creates two independent roots:

```text
Background        Background
    │                 │
    ▼                 ▼
   ctx1              ctx2
```

You normally create a root context at an appropriate application entry point and pass derived contexts downward.

## What If There Is No `Background()`?

A derived context needs a valid parent.

Do not do this:

```go
ctx, cancel := context.WithCancel(nil)
```

Instead:

```go
root := context.Background()

ctx, cancel := context.WithCancel(root)
```

Or use an existing context:

```go
ctx2, cancel := context.WithTimeout(ctx1, 5*time.Second)
```

Here `ctx1` is already the parent, so another `Background()` is unnecessary.

## `Background()` vs `TODO()`

Go also provides:

```go
context.TODO()
```

Basic distinction:

### `Background()`

Use it when you know you need a root context.

```go
ctx := context.Background()
```

Mental model:

```text
"I know this is my root context."
```

### `TODO()`

Use it when you have not yet decided which context should be used.

```go
ctx := context.TODO()
```

Mental model:

```text
"I still need to decide what context belongs here."
```

## Simple Example

```go
package main

import (
	"context"
	"fmt"
)

func service(ctx context.Context) {
	fmt.Println("service running")
}

func main() {
	ctx := context.Background()

	service(ctx)
}
```

Flow:

```text
main
 │
 │ create
 ▼
Background
 │
 │ pass
 ▼
service(ctx)
```

The context is passed into the function even though this example does not use cancellation or timeout yet.

## Key Takeaways

1. `context.Background()` is a root context.
2. It is not automatically created by Go.
3. It does not provide cancellation by itself.
4. It can be used as the parent of derived contexts.
5. A derived context has one parent.
6. One parent can have multiple children.
7. You can have multiple independent `Background()` roots.
8. Do not use `nil` as the parent of a derived context.
9. You do not need a new `Background()` for every derived context.
10. Contexts form a parent-child tree.

## Mental Model

```text
Background
    │
    ├── WithCancel
    │       │
    │       └── child context
    │
    ├── WithTimeout
    │       │
    │       └── child context
    │
    └── WithDeadline
            │
            └── child context
```

The simplest definition to remember:

> **`context.Background()` is a root context used as a starting point for building a context tree.**
