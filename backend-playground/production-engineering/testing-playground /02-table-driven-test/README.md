# Table-Driven Test

This playground demonstrates the **Table-Driven Test** pattern in Go.

## Structure

```text
02-table-driven-test/
├── README.md
├── calculator.go
└── calculator_test.go
```

## What is a Table-Driven Test?

A Table-Driven Test stores multiple test cases in a slice and runs them using the same test logic.

Instead of creating separate test functions:

```text
TestAdd1
TestAdd2
TestAdd3
```

we define test cases in a table:

```text
Test Cases
├── 1 + 1 = 2
├── 2 + 2 = 4
└── 3 + 3 = 6
```

---

## Basic Pattern

```go
func TestAdd(t *testing.T) {
    tests := []struct {
        name string
        a    int
        b    int
        want int
    }{
        {
            name: "1 + 1 = 2",
            a:    1,
            b:    1,
            want: 2,
        },
        {
            name: "2 + 2 = 4",
            a:    2,
            b:    2,
            want: 4,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := Add(tt.a, tt.b)

            if got != tt.want {
                t.Errorf("got %d, want %d", got, tt.want)
            }
        })
    }
}
```

---

## Flow

```text
Test Cases
    ↓
Loop
    ↓
t.Run()
    ↓
Execute Function
    ↓
Compare Actual vs Expected
    ↓
PASS / FAIL
```

---

## Why Use It?

Table-driven tests are useful when the same logic needs to be tested with many inputs.

Benefits:

```text
Table-Driven Test
├── Less duplicated code
├── Easy to add test cases
├── Clear test cases
└── Easy to identify failing cases
```

Example output:

```text
TestAdd
├── 1 + 1 = 2     PASS
├── 2 + 2 = 4     PASS
└── 3 + 3 = 6     PASS
```

If one case fails:

```text
TestAdd
├── 1 + 1 = 2     PASS
├── 2 + 2 = 5     FAIL
└── 3 + 3 = 6     PASS
```

The test name makes it easy to identify the failing case.

---

## Key Takeaway

```text
Table-Driven Test
=
Multiple test cases
+
One reusable test logic
+
t.Run()
```

Common structure:

```go
tests := []struct {
    name string
    // inputs
    // expected output
}{}

for _, tt := range tests {
    t.Run(tt.name, func(t *testing.T) {
        // test
    })
}
```

---

## Run Tests

From the project root:

```bash
go test ./...
```

Verbose:

```bash
go test -v ./...
```
