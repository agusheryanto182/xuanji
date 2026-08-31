# HTTP Handler Testing

This playground demonstrates how to test HTTP handlers in Go using the standard `net/http/httptest` package.

## Structure

```text
06-http-handler-testing/
├── README.md
├── handler.go
└── handler_test.go
```

## Flow

```text
HTTP Request
     ↓
HTTP Handler
     ↓
Usecase / Service
     ↓
HTTP Response
```

During unit testing:

```text
HTTP Request
     ↓
Handler
     ↓
Fake Service
     ↓
HTTP Response
```

No real database is required.

---

## What We Test

### Success

```text
GET /users?id=1
        ↓
     Handler
        ↓
   Service returns User
        ↓
     200 OK
        ↓
    JSON Response
```

### Not Found

```text
GET /users?id=999
        ↓
     Handler
        ↓
   Service returns error
        ↓
   404 Not Found
```

---

## `httptest`

Go provides the `net/http/httptest` package for testing HTTP handlers.

Create a request:

```go
req := httptest.NewRequest(
    http.MethodGet,
    "/users?id=1",
    nil,
)
```

Create a response recorder:

```go
rec := httptest.NewRecorder()
```

Execute the handler:

```go
handler.GetUser(rec, req)
```

Then inspect the response:

```go
rec.Code
rec.Body
rec.Header()
```

---

## What Should Be Tested?

For a handler, focus on the HTTP behavior:

```text
Handler Test
├── Request
├── Status Code
├── Response Body
└── Headers
```

Example:

```go
if rec.Code != http.StatusOK {
    t.Errorf(
        "got status %d, want %d",
        rec.Code,
        http.StatusOK,
    )
}
```

Decode JSON response:

```go
var got User

err := json.NewDecoder(rec.Body).Decode(&got)
if err != nil {
    t.Fatal(err)
}
```

Then verify the response:

```go
if got.ID != 1 {
    t.Errorf("got ID %d, want 1", got.ID)
}
```

---

## Fake Service

The handler should not need a real service implementation during the unit test.

Instead, use a fake:

```go
type fakeUserService struct {
    user User
    err  error
}

func (f *fakeUserService) GetUser(id int) (User, error) {
    return f.user, f.err
}
```

Flow:

```text
Handler
   ↓
Fake Service
   ↓
Predefined Result
```

This keeps the test focused on the handler.

---

## Key Takeaway

```text
HTTP Handler Test
=
Create Request
+
Execute Handler
+
Inspect Response
```

The main things to verify are:

```text
Request
   ↓
Handler
   ↓
Status Code
   ↓
Response Body
   ↓
Headers
```

The database and external dependencies should be tested separately.

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
