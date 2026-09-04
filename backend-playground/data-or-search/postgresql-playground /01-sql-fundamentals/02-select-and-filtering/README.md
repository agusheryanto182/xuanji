# 02 — SELECT and Filtering

## Goal

Learn the SQL patterns used constantly in backend production to retrieve the right rows.

Focus:

- `SELECT`
- selecting specific columns
- expressions and aliases
- `WHERE`
- comparison operators
- `AND` / `OR` / `NOT`
- `IN` / `NOT IN`
- `BETWEEN`
- `LIKE` / `ILIKE`
- parameterized queries
- common production mistakes

The goal is not to memorize syntax.

The goal is:

> **Translate a backend requirement into a precise database query.**

---

## 1. SELECT

`SELECT` retrieves data.

```sql
SELECT id, name
FROM users;
```

This means:

```text
get id and name
from users
```

Avoid starting every production query with:

```sql
SELECT *
```

Prefer the columns the application actually needs.

---

## 2. Select Specific Columns

Instead of:

```sql
SELECT *
FROM users;
```

use:

```sql
SELECT id, name, email
FROM users;
```

This makes the query's intent explicit.

### Production benefits

Selecting only required columns can reduce:

- data transferred
- database work
- application memory
- accidental exposure of internal fields

`SELECT *` is still useful when exploring data manually.

---

## 3. Column Aliases

Aliases give result columns clearer names.

```sql
SELECT
    id,
    name AS user_name
FROM users;
```

The database column remains:

```text
name
```

but the result column is:

```text
user_name
```

Aliases are useful when:

- joining tables
- returning calculated values
- making API mapping clearer

---

## 4. Expressions

A `SELECT` can return expressions, not only stored columns.

```sql
SELECT
    price,
    price * quantity AS subtotal
FROM order_items;
```

The database calculates:

```text
price × quantity
```

for each returned row.

This is often useful when the calculation naturally belongs to the query.

---

## 5. WHERE

`WHERE` filters rows.

```sql
SELECT id, name
FROM users
WHERE status = 'active';
```

Think:

```text
FROM users
    ↓
keep rows matching WHERE
    ↓
return selected columns
```

The key question is:

> Which rows should be included?

---

## 6. Comparison Operators

Common operators:

```text
=       equal
<>      not equal
>       greater than
>=      greater than or equal
<       less than
<=      less than or equal
```

Examples:

```sql
SELECT *
FROM products
WHERE price > 100;
```

```sql
SELECT *
FROM orders
WHERE status <> 'cancelled';
```

These operators are among the most common filtering tools in backend queries.

---

## 7. AND

`AND` requires both conditions to be true.

```sql
SELECT *
FROM users
WHERE status = 'active'
  AND age >= 18;
```

Meaning:

```text
status must be active
AND
age must be at least 18
```

Use `AND` when multiple requirements must all be satisfied.

---

## 8. OR

`OR` requires at least one condition to be true.

```sql
SELECT *
FROM users
WHERE role = 'admin'
   OR role = 'manager';
```

Meaning:

```text
role is admin
OR
role is manager
```

Be careful when combining `AND` and `OR`.

---

## 9. Parentheses Matter

Consider:

```sql
WHERE status = 'active'
  AND role = 'admin'
   OR role = 'manager'
```

This can be interpreted as:

```text
(status = active AND role = admin)
OR role = manager
```

If the requirement is:

```text
active users
who are either admins or managers
```

write it explicitly:

```sql
WHERE status = 'active'
  AND (
      role = 'admin'
      OR role = 'manager'
  );
```

### Production rule

> Use parentheses when mixing `AND` and `OR` if the intended logic is not immediately obvious.

---

## 10. NOT

`NOT` negates a condition.

```sql
SELECT *
FROM users
WHERE NOT status = 'inactive';
```

Often, a direct condition is clearer:

```sql
WHERE status <> 'inactive'
```

Do not use `NOT` just because it is available.

Prefer the expression that communicates the requirement clearly.

---

## 11. IN

`IN` checks whether a value belongs to a set.

Instead of:

```sql
WHERE status = 'active'
   OR status = 'pending'
   OR status = 'review'
```

you can write:

```sql
WHERE status IN ('active', 'pending', 'review');
```

This is common in backend filtering.

For example:

```text
GET /orders?status=paid,pending
```

can translate into a parameterized `IN` condition.

---

## 12. NOT IN

`NOT IN` excludes values from a set.

```sql
SELECT *
FROM orders
WHERE status NOT IN ('cancelled', 'refunded');
```

However, `NOT IN` has an important interaction with `NULL`.

For nullable data, this can produce surprising results.

Example:

```sql
WHERE user_id NOT IN (1, 2, NULL)
```

Because of SQL's three-valued logic, the condition cannot behave like a simple "not one of these values."

For nullable relationships, `NOT EXISTS` is often a safer pattern.

You will study NULL behavior separately.

---

## 13. BETWEEN

`BETWEEN` checks an inclusive range.

```sql
SELECT *
FROM products
WHERE price BETWEEN 100 AND 500;
```

Conceptually:

```text
price >= 100
AND
price <= 500
```

### Important

Both boundaries are included.

For numeric values, this is usually straightforward.

For timestamps, be careful.

---

## 14. Time Range Filtering

A common production requirement is:

> Get records created during a day.

Avoid depending on a timestamp such as:

```text
2026-09-02 23:59:59
```

A safer pattern is a half-open interval:

```sql
WHERE created_at >= '2026-09-02 00:00:00'
  AND created_at <  '2026-09-03 00:00:00'
```

This means:

```text
start included
end excluded
```

It avoids precision problems involving milliseconds or microseconds.

This pattern is especially useful for:

- reports
- dashboards
- daily jobs
- date filters
- audit queries

---

## 15. LIKE

`LIKE` performs pattern matching.

```sql
SELECT id, name
FROM users
WHERE name LIKE 'Ag%';
```

`%` means:

```text
zero or more characters
```

Examples:

```text
'Ag%'   → starts with Ag
'%Ag'   → ends with Ag
'%Ag%'  → contains Ag
```

`_` represents one character.

---

## 16. ILIKE in PostgreSQL

PostgreSQL provides:

```sql
ILIKE
```

for case-insensitive pattern matching.

Example:

```sql
SELECT id, name
FROM users
WHERE name ILIKE '%agus%';
```

This can match variations in letter casing.

### Production note

A pattern search can become expensive on large tables, especially with a leading wildcard:

```sql
LIKE '%keyword%'
```

Do not assume a normal B-tree index will make every `LIKE` query fast.

For large-scale search requirements, PostgreSQL may need a more appropriate search/index strategy.

---

## 17. Filtering in the Database vs Application

Prefer filtering at the database when the database can perform the filtering.

Avoid:

```text
SELECT every user
↓
send every user to Go
↓
filter in Go
```

when you can do:

```sql
SELECT id, name
FROM users
WHERE status = 'active';
```

Database-side filtering usually means:

```text
less data transferred
less application memory
less application processing
```

The database is designed to filter data efficiently.

---

## 18. Parameterized Queries

Backend applications should not concatenate user input into SQL.

Bad:

```text
"SELECT ... WHERE email = '" + email + "'"
```

Good:

```sql
SELECT id, name
FROM users
WHERE email = $1;
```

The value is supplied separately.

In Go, conceptually:

```go
row := db.QueryRowContext(
    ctx,
    `SELECT id, name FROM users WHERE email = $1`,
    email,
)
```

Parameterized queries are a baseline production practice.

---

## 19. Filtering API Requests

Suppose an API supports:

```text
GET /orders?status=paid
```

The application can map that to:

```sql
SELECT id, user_id, total
FROM orders
WHERE status = $1;
```

The important separation is:

```text
SQL structure
+
parameter value
```

Do not construct SQL syntax directly from request values.

---

## 20. Dynamic Filtering

Real APIs may support optional filters:

```text
status
user_id
created_from
created_to
```

Conceptually:

```text
base query
+
only the filters requested
```

For example:

```sql
SELECT id, user_id, total
FROM orders
WHERE status = $1
  AND user_id = $2;
```

The application should build dynamic query structure carefully.

Values should still be parameterized.

For dynamic column names or sort fields, do not treat them like ordinary values. Whitelist allowed identifiers.

---

## 21. Filtering and Indexes

Filtering patterns often determine index design.

Suppose production frequently runs:

```sql
SELECT id, name
FROM users
WHERE email = $1;
```

An index on:

```text
email
```

may be appropriate.

For another query:

```sql
SELECT id, total
FROM orders
WHERE user_id = $1
ORDER BY created_at DESC
LIMIT 20;
```

the useful index strategy may be different.

The important lesson:

> Indexes should follow real query patterns.

Filtering and indexing should be considered together.

---

## 22. Equality Filters

Equality filtering is extremely common:

```sql
WHERE user_id = $1
```

```sql
WHERE status = $1
```

```sql
WHERE email = $1
```

These are common candidates for efficient indexed access when the workload justifies it.

But an index is not automatically useful for every low-cardinality column.

For example:

```text
boolean status
```

may have poor selectivity depending on the data and query.

Measure actual workloads.

---

## 23. Multiple Filters

A typical production query may combine several conditions:

```sql
SELECT id, total
FROM orders
WHERE user_id = $1
  AND status = $2
  AND created_at >= $3;
```

The database evaluates the combined condition according to SQL semantics and chooses an execution plan.

Do not automatically create one index for every condition.

Consider the complete query pattern.

---

## 24. NULL and WHERE

A key rule:

```sql
WHERE column = NULL
```

does not test whether the column is NULL.

Use:

```sql
WHERE column IS NULL
```

and:

```sql
WHERE column IS NOT NULL
```

Why?

Because SQL uses a third logical state:

```text
UNKNOWN
```

for many NULL comparisons.

The dedicated NULL material covers this in detail.

---

## 25. Filtering Soft-Deleted Rows

A common application pattern is:

```sql
SELECT id, name
FROM users
WHERE deleted_at IS NULL;
```

Here:

```text
deleted_at IS NULL
→ active record
```

If soft deletion is part of the design, this condition often appears throughout application queries.

Production concern:

> Repeated filtering rules should be deliberately designed and indexed when necessary, not copied blindly everywhere.

---

## 26. Filtering by Status

A typical backend query:

```sql
SELECT id, total
FROM orders
WHERE status = 'paid';
```

For multiple statuses:

```sql
SELECT id, total
FROM orders
WHERE status IN ('paid', 'processing');
```

Status filtering appears frequently in:

```text
admin dashboards
background jobs
order processing
notifications
workflow systems
```

---

## 27. Filtering by User

A common authorization/data-scope pattern:

```sql
SELECT id, title
FROM posts
WHERE user_id = $1;
```

The application supplies the authenticated user's ID.

This is different from trusting a user-provided ID without checking authorization.

### Production point

Filtering data by identity is a query concern.

Whether the current user is allowed to access that data is an authorization concern.

Both need to be correct.

---

## 28. Filtering by Time

Typical query:

```sql
SELECT id, total
FROM orders
WHERE created_at >= $1
  AND created_at < $2;
```

The application can provide:

```text
start
end
```

This pattern works well for:

```text
reports
analytics
scheduled processing
date-range APIs
```

Use consistent timezone handling across the application and database.

---

## 29. Query Thinking

Before writing SQL, translate the requirement into pieces.

Example:

> Get the 20 paid orders for a user created today.

Break it down:

```text
entity
→ orders

user
→ user_id

status
→ paid

date
→ today

amount
→ 20
```

Then:

```sql
SELECT id, total, created_at
FROM orders
WHERE user_id = $1
  AND status = 'paid'
  AND created_at >= $2
  AND created_at < $3
ORDER BY created_at DESC, id DESC
LIMIT 20;
```

This way of thinking scales better than memorizing query templates.

---

## 30. Common Mistakes

### Using `SELECT *` everywhere

Prefer explicit columns in production queries.

### Filtering in application code unnecessarily

Let the database filter data when appropriate.

### Mixing `AND` and `OR` without parentheses

Make the intended logic explicit.

### Using `= NULL`

Use `IS NULL`.

### Ignoring `NOT IN` + NULL behavior

Nullable values can change the result because of three-valued logic.

### Using an inclusive timestamp end

Prefer:

```text
>= start
< end
```

for date/time ranges.

### Concatenating user input into SQL

Use parameters.

### Trusting dynamic column names

Whitelist allowed identifiers instead of inserting arbitrary request values into SQL.

### Assuming every filter should have an index

Index design depends on workload.

### Returning unbounded results

Use filtering, limits, and pagination.

---

## 31. Production Checklist

Before shipping a filtering query:

```text
□ Am I selecting only required columns?
□ Is WHERE expressing the exact business requirement?
□ Are AND/OR conditions grouped correctly?
□ Are NULL cases understood?
□ Are values parameterized?
□ Is the result size bounded?
□ Is ordering required?
□ Can joins multiply rows?
□ Does the query match an important production access pattern?
□ Does the relevant index strategy make sense?
```

---

## 32. Practical Mental Model

Think:

```text
Requirement
    ↓
What entity?
    ↓
What rows?
    ↓
What filters?
    ↓
What columns?
    ↓
What ordering?
    ↓
What limit?
    ↓
How will this behave at production scale?
```

For example:

```text
"Show recent active users"

entity
→ users

filter
→ status = active

ordering
→ created_at DESC

result
→ required columns only
```

Then write the SQL.

---

## 33. What to Remember

1. `SELECT` retrieves data.
2. Select only the columns you need in production queries.
3. `WHERE` determines which rows are included.
4. `AND` means all conditions must match.
5. `OR` means at least one condition matches.
6. Use parentheses when combining complex boolean logic.
7. `IN` is useful for matching a set of values.
8. `BETWEEN` is inclusive; be careful with timestamps.
9. For time ranges, `>= start AND < end` is a robust production pattern.
10. `LIKE`/`ILIKE` are useful for text patterns but can require specialized indexing for large-scale search.
11. Use parameterized queries.
12. Filter in the database instead of unnecessarily loading large datasets into the application.
13. Query patterns should influence index design.
14. Always consider NULL behavior.
15. Think about the production-sized result, not only the development-sized result.

---

## Next

Next material:

```text
03-sorting-and-limiting
```

Focus:

```text
ORDER BY
ASC / DESC
multiple sort keys
deterministic ordering
LIMIT
OFFSET
basic pagination
production query patterns
common mistakes
```
