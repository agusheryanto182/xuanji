# 03 — Sorting and Limiting

## Goal

Learn how to control the order and size of SQL results for production backend queries.

Focus:

- `ORDER BY`
- ascending and descending order
- multiple sort keys
- deterministic ordering
- `LIMIT`
- `OFFSET`
- basic pagination
- production query patterns
- common mistakes

The goal is:

> Return the right records, in the right order, without unnecessarily processing or returning huge result sets.

---

## 1. ORDER BY

`ORDER BY` controls the order of the result.

```sql
SELECT id, name, created_at
FROM users
ORDER BY created_at DESC;
```

This means:

```text
newest users first
```

Without `ORDER BY`, do not assume rows will come back in a particular order.

### Production rule

> If the API or business requirement cares about order, explicitly specify `ORDER BY`.

---

## 2. ASC and DESC

Two common directions:

```sql
ORDER BY created_at ASC;
```

```sql
ORDER BY created_at DESC;
```

Meaning:

```text
ASC  → ascending
DESC → descending
```

Examples:

```text
price ASC
→ cheapest first

price DESC
→ most expensive first

created_at DESC
→ newest first
```

---

## 3. Default Direction

If the direction is omitted:

```sql
ORDER BY created_at;
```

the default is ascending.

For production code, explicitly writing:

```sql
ORDER BY created_at ASC
```

or:

```sql
ORDER BY created_at DESC
```

can make the intended behavior clearer.

---

## 4. Multiple Sort Keys

You can sort by multiple columns.

```sql
SELECT id, name, created_at
FROM users
ORDER BY created_at DESC, id DESC;
```

The database sorts primarily by:

```text
created_at DESC
```

If two rows have the same timestamp, it uses:

```text
id DESC
```

as the tie-breaker.

This pattern is extremely useful in production.

---

## 5. Deterministic Ordering

Suppose:

```text
created_at
```

is identical for several rows.

This:

```sql
ORDER BY created_at DESC
```

does not define how those tied rows should be ordered relative to each other.

Add a stable tie-breaker:

```sql
ORDER BY created_at DESC, id DESC;
```

Now the ordering is deterministic for unique IDs.

### Production rule

> If result order matters, make ties deterministic.

This becomes especially important for pagination.

---

## 6. Why Deterministic Ordering Matters for Pagination

Imagine an API returns:

```text
20 rows
```

and the client requests the next page.

If rows can move around because the ordering is not deterministic, a row may:

```text
appear twice
```

or:

```text
be skipped
```

A common pattern is:

```sql
ORDER BY created_at DESC, id DESC
LIMIT 20;
```

The second ordering key provides a stable tie-breaker.

---

## 7. ORDER BY Expressions

You can order by an expression.

```sql
SELECT
    id,
    price,
    quantity,
    price * quantity AS subtotal
FROM order_items
ORDER BY subtotal DESC;
```

This can be useful for calculated results.

Keep the query readable when the expression becomes complicated.

---

## 8. Ordering by Alias

An output alias can be used in `ORDER BY`.

```sql
SELECT
    id,
    price * quantity AS subtotal
FROM order_items
ORDER BY subtotal DESC;
```

This is often cleaner than repeating:

```sql
price * quantity
```

in the `ORDER BY`.

---

## 9. NULL Ordering

PostgreSQL allows explicit NULL ordering:

```sql
ORDER BY deleted_at ASC NULLS FIRST;
```

or:

```sql
ORDER BY deleted_at DESC NULLS LAST;
```

This matters when a column is nullable.

Do not assume NULL will always behave the way your business requirement expects.

When NULL placement matters, make it explicit.

---

## 10. LIMIT

`LIMIT` restricts how many rows are returned.

```sql
SELECT id, name
FROM users
LIMIT 20;
```

This returns at most:

```text
20 rows
```

`LIMIT` is useful for:

- API responses
- dashboards
- top-N queries
- pagination
- protecting endpoints from unbounded results

---

## 11. LIMIT Needs ORDER BY for "Top" Results

This:

```sql
SELECT id, total
FROM orders
LIMIT 10;
```

means:

> Give me up to 10 rows.

It does **not** mean:

> Give me the 10 newest orders.

For newest:

```sql
SELECT id, total
FROM orders
ORDER BY created_at DESC, id DESC
LIMIT 10;
```

For highest value:

```sql
SELECT id, total
FROM orders
ORDER BY total DESC, id DESC
LIMIT 10;
```

### Production rule

> `LIMIT` without meaningful ordering is not a reliable top-N query.

---

## 12. OFFSET

`OFFSET` skips rows before returning the result.

```sql
SELECT id, name
FROM users
ORDER BY created_at DESC, id DESC
LIMIT 20
OFFSET 40;
```

Conceptually:

```text
skip 40
return next 20
```

This is common for simple page-based APIs.

---

## 13. Basic Offset Pagination

A simple pagination request might be:

```text
page = 3
page_size = 20
```

The offset is:

```text
(page - 1) × page_size
```

So:

```text
(3 - 1) × 20
= 40
```

Query:

```sql
SELECT id, name
FROM users
ORDER BY created_at DESC, id DESC
LIMIT 20
OFFSET 40;
```

This is easy to understand and can be perfectly reasonable for smaller datasets.

---

## 14. OFFSET Has a Cost

For a large offset:

```sql
LIMIT 20 OFFSET 100000;
```

PostgreSQL may still need to process/skip many rows before producing the requested page.

As the offset grows, pagination can become less efficient.

This does not mean `OFFSET` is always bad.

It means:

> Understand the workload and dataset size.

---

## 15. When OFFSET Pagination Is Fine

Offset pagination is often reasonable for:

```text
small datasets
admin pages
internal tools
low page depths
simple APIs
```

Its biggest advantage is simplicity.

Example:

```text
?page=2&page_size=20
```

is easy for clients to understand.

---

## 16. When OFFSET Becomes a Problem

Be careful when:

```text
tables are very large
users frequently request deep pages
queries are expensive
traffic is high
```

For large feeds or high-volume datasets, keyset/cursor pagination can be a better approach.

---

## 17. Keyset Pagination

Instead of saying:

```text
skip 100000 rows
```

keyset pagination says:

> Continue after this known position.

For:

```sql
ORDER BY created_at DESC, id DESC
```

a next-page condition can look like:

```sql
WHERE (created_at, id) < ($1, $2)
ORDER BY created_at DESC, id DESC
LIMIT 20;
```

The cursor contains the last row's:

```text
created_at
id
```

This can avoid large offsets.

---

## 18. Why Keyset Uses a Tie-Breaker

Suppose you only use:

```sql
ORDER BY created_at DESC;
```

Two rows can have the same timestamp.

A cursor based only on `created_at` cannot uniquely identify the position.

Using:

```text
created_at
+
id
```

creates a stable ordering position when `id` is unique.

This is why deterministic ordering and keyset pagination are closely related.

---

## 19. Offset vs Keyset

### Offset

```sql
LIMIT 20 OFFSET 1000;
```

Pros:

```text
simple
easy page numbers
easy to implement
```

Cons:

```text
deep offsets can become expensive
results can shift between requests
```

### Keyset

```sql
WHERE (created_at, id) < ($1, $2)
ORDER BY created_at DESC, id DESC
LIMIT 20;
```

Pros:

```text
efficient for deep navigation
good for large feeds
works naturally with stable ordering
```

Cons:

```text
more complex
cursor-based navigation
page numbers are less natural
```

Production choice depends on the use case.

---

## 20. Pagination Is a Data Consistency Problem Too

Consider:

```text
Request page 1
↓
new rows are inserted
↓
Request page 2
```

With offset pagination, the dataset may have shifted.

That can cause:

```text
duplicates
missing rows
```

depending on the ordering and concurrent writes.

Keyset pagination can provide more stable traversal of a changing ordered dataset, especially when designed around immutable/stable sort keys.

No pagination strategy completely removes the need to think about changing data.

---

## 21. Stable Sort Keys

For pagination, prefer ordering fields that are:

```text
stable
predictable
indexed appropriately
```

A common pattern is:

```sql
ORDER BY created_at DESC, id DESC;
```

where:

```text
created_at
→ business ordering

id
→ unique tie-breaker
```

Avoid using a frequently changing field as the sole cursor/order key when stable traversal is required.

---

## 22. Indexes and ORDER BY

Sorting can require work.

For example:

```sql
SELECT id, total
FROM orders
WHERE user_id = $1
ORDER BY created_at DESC
LIMIT 20;
```

A suitable index may allow PostgreSQL to retrieve rows in a useful order without sorting a large intermediate result.

A possible index pattern is:

```sql
CREATE INDEX orders_user_created_idx
ON orders(user_id, created_at DESC);
```

The exact index should be based on the real query workload and query plan.

Do not create this index blindly.

---

## 23. Filtering + Sorting + LIMIT

A very common production query combines all three:

```sql
SELECT id, total, created_at
FROM orders
WHERE user_id = $1
  AND status = $2
ORDER BY created_at DESC, id DESC
LIMIT 20;
```

When optimizing this kind of query, think about the complete pattern:

```text
WHERE
+
ORDER BY
+
LIMIT
```

Do not optimize each clause independently.

---

## 24. Top-N Queries

A top-N query usually means:

```text
filter
↓
order
↓
limit
```

Example:

> Get the 10 highest-value paid orders.

```sql
SELECT id, user_id, total
FROM orders
WHERE status = 'paid'
ORDER BY total DESC, id DESC
LIMIT 10;
```

This pattern appears frequently in:

```text
dashboards
leaderboards
admin panels
reports
recommendation screens
```

---

## 25. Latest Record

A common backend requirement:

> Get the latest order for a user.

```sql
SELECT id, total, created_at
FROM orders
WHERE user_id = $1
ORDER BY created_at DESC, id DESC
LIMIT 1;
```

The important pieces are:

```text
filter by user
↓
sort newest first
↓
limit to one
```

This pattern is extremely common.

---

## 26. Oldest Record

Reverse the ordering:

```sql
SELECT id, total, created_at
FROM orders
WHERE user_id = $1
ORDER BY created_at ASC, id ASC
LIMIT 1;
```

Again:

```text
filter
↓
order
↓
limit
```

---

## 27. Pagination API Design

A simple offset API might expose:

```text
GET /users?page=3&page_size=20
```

The backend translates that into:

```text
LIMIT 20
OFFSET 40
```

A cursor API might expose:

```text
GET /users?limit=20&cursor=...
```

The cursor represents the last position in the ordered result.

The database query remains responsible for retrieving the correct slice.

---

## 28. Validate Page Size

Do not blindly accept:

```text
?page_size=1000000
```

A production API should normally enforce a maximum.

For example:

```text
default = 20
maximum = 100
```

The exact values depend on the endpoint.

The principle is:

> Keep database result size bounded by server-side rules.

---

## 29. Dynamic Sort Fields

APIs sometimes support:

```text
?sort=name
?sort=created_at
```

Do not directly insert the request value into SQL:

```text
ORDER BY <raw user input>
```

Column names are SQL identifiers, not normal parameter values.

Instead, whitelist supported fields.

Conceptually:

```text
created_at → "created_at"
name       → "name"
price      → "price"
```

and reject unsupported values.

The direction should also be validated:

```text
asc
desc
```

This prevents SQL injection and unexpected queries.

---

## 30. Sorting by User Input

Safe conceptually:

```text
request:
sort = created_at

application:
created_at → known SQL identifier

SQL:
ORDER BY created_at DESC
```

Unsafe:

```text
request:
sort = arbitrary SQL expression

application:
concatenate directly into query
```

Parameterized queries protect values, but dynamic identifiers need a different strategy: validation/whitelisting.

---

## 31. LIMIT as a Safety Boundary

An endpoint that returns a list should generally have a bounded result size.

Instead of:

```sql
SELECT id, name
FROM users;
```

a list endpoint might use:

```sql
SELECT id, name
FROM users
ORDER BY created_at DESC, id DESC
LIMIT $1;
```

with the application enforcing a maximum allowed limit.

This protects:

```text
database resources
application memory
network bandwidth
API latency
```

---

## 32. ORDER BY and NULL

When a column can be NULL, make the intended behavior explicit when it matters.

Example:

```sql
ORDER BY last_login_at DESC NULLS LAST;
```

This expresses:

```text
recently active users first
users who never logged in last
```

This is often clearer than relying on implicit NULL ordering.

---

## 33. ORDER BY and Joins

After a join, be explicit about which column you are sorting by.

Instead of:

```sql
ORDER BY created_at DESC;
```

prefer:

```sql
ORDER BY orders.created_at DESC;
```

when multiple tables contain similarly named columns.

This makes the query easier to understand and avoids ambiguity.

---

## 34. Sorting After Aggregation

Aggregation can also be sorted.

Example:

```sql
SELECT
    user_id,
    COUNT(*) AS order_count
FROM orders
GROUP BY user_id
ORDER BY order_count DESC
LIMIT 10;
```

This means:

```text
group orders by user
↓
count orders
↓
highest count first
↓
return top 10
```

This pattern is common in reporting and dashboards.

---

## 35. Common Mistake: Assuming Natural Row Order

Do not rely on:

```sql
SELECT *
FROM users;
```

returning:

```text
oldest → newest
```

or:

```text
id ascending
```

unless the query explicitly specifies it.

Use:

```sql
ORDER BY ...
```

whenever order matters.

---

## 36. Common Mistake: LIMIT Without ORDER BY

This:

```sql
SELECT *
FROM users
LIMIT 10;
```

does not define which 10 users you get.

If you want newest:

```sql
ORDER BY created_at DESC, id DESC
LIMIT 10;
```

If you want highest price:

```sql
ORDER BY price DESC, id DESC
LIMIT 10;
```

---

## 37. Common Mistake: Non-Deterministic Pagination

Avoid:

```sql
ORDER BY created_at DESC
LIMIT 20;
```

for pagination when timestamps can tie.

Prefer:

```sql
ORDER BY created_at DESC, id DESC
LIMIT 20;
```

The second key provides deterministic ordering.

---

## 38. Common Mistake: Deep OFFSET Pagination

This:

```sql
LIMIT 20 OFFSET 100000;
```

may require substantial work.

For large datasets and deep navigation, evaluate keyset/cursor pagination.

Do not switch automatically; measure and choose based on the use case.

---

## 39. Common Mistake: Unbounded Page Size

Never assume clients will send reasonable values.

Protect endpoints against:

```text
limit = 1000000
```

Use:

```text
default
+
maximum
```

server-side.

---

## 40. Common Mistake: Trusting Dynamic Sort Input

Do not concatenate arbitrary request values into:

```sql
ORDER BY
```

Whitelist valid sort fields and directions.

---

## 41. Common Mistake: Sorting on the Wrong Business Field

For:

> newest users

use a meaningful timestamp:

```sql
ORDER BY created_at DESC
```

not simply:

```sql
ORDER BY id DESC
```

unless the ID semantics explicitly guarantee the required ordering.

---

## 42. Common Mistake: Ignoring Indexes

A query such as:

```sql
WHERE user_id = $1
ORDER BY created_at DESC
LIMIT 20
```

may become important enough to justify an index.

Do not guess the index.

Inspect the workload and query plan.

---

## 43. Production Checklist

Before shipping a list/query endpoint:

```text
□ Is the order explicitly defined?
□ Is the ordering deterministic?
□ Is a stable tie-breaker needed?
□ Is the result size bounded?
□ Is LIMIT validated server-side?
□ Is OFFSET appropriate for this dataset?
□ Would keyset pagination be better?
□ Are filters combined with ordering correctly?
□ Are dynamic sort fields whitelisted?
□ Are NULL ordering rules understood?
□ Does the query have an appropriate index strategy?
□ Have I checked the query plan at realistic data size?
```

---

## 44. Practical Mental Model

Think:

```text
What rows?
    ↓
WHERE

What order?
    ↓
ORDER BY

How many?
    ↓
LIMIT

Which portion?
    ↓
OFFSET / cursor
```

For a production list endpoint:

```text
filter
↓
stable ordering
↓
bounded result
↓
appropriate pagination
```

---

## 45. Example: Production List Query

Requirement:

> Show the latest 20 paid orders for a user.

Think:

```text
entity
→ orders

filter
→ user_id

filter
→ paid

order
→ newest

tie-breaker
→ id

limit
→ 20
```

SQL:

```sql
SELECT id, total, created_at
FROM orders
WHERE user_id = $1
  AND status = 'paid'
ORDER BY created_at DESC, id DESC
LIMIT 20;
```

This is a very typical backend query.

---

## 46. Example: Simple Pagination

Requirement:

> Get page 3 with 20 users per page.

```sql
SELECT id, name, created_at
FROM users
ORDER BY created_at DESC, id DESC
LIMIT 20
OFFSET 40;
```

The API calculates:

```text
offset = (page - 1) × page_size
```

and should enforce a reasonable maximum page size.

---

## 47. Example: Cursor Pagination

Requirement:

> Get the next 20 users after the last row from the previous page.

Previous cursor:

```text
created_at = $1
id = $2
```

Query:

```sql
SELECT id, name, created_at
FROM users
WHERE (created_at, id) < ($1, $2)
ORDER BY created_at DESC, id DESC
LIMIT 20;
```

This is a common foundation for large feed-style APIs.

---

## 48. What to Remember

1. `ORDER BY` defines result order.
2. Without `ORDER BY`, do not rely on row order.
3. `ASC` means ascending; `DESC` means descending.
4. Multiple sort keys provide deterministic tie-breaking.
5. `LIMIT` bounds the number of returned rows.
6. `LIMIT` alone does not define which rows are returned.
7. `OFFSET` provides simple page-based pagination.
8. Deep offsets can become expensive on large datasets.
9. Keyset/cursor pagination can be better for large ordered datasets.
10. Stable ordering is essential for reliable pagination.
11. Dynamic sort fields must be whitelisted.
12. Page size should have a server-side maximum.
13. Index design should consider `WHERE + ORDER BY + LIMIT` together.
14. Measure query performance with realistic data.

---

## Next

Next material:

```text
04-null-and-three-valued-logic
```

Focus:

```text
NULL
IS NULL / IS NOT NULL
TRUE / FALSE / UNKNOWN
WHERE behavior
AND / OR / NOT with NULL
IN / NOT IN + NULL
COUNT(*) vs COUNT(column)
COALESCE
NULL in joins
common production mistakes
```

The material will stay concise and focus on NULL behavior that actually causes production bugs.
