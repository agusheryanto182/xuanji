# 07 — SQL Query Thinking

## Goal

Learn how to design SQL queries from a production requirement instead of starting with random SQL syntax.

This material connects the previous fundamentals:

```text
SELECT
WHERE
ORDER BY
LIMIT
NULL
GROUP BY
HAVING
JOIN
INSERT
UPDATE
DELETE
```

The goal is:

> **Translate a business requirement into a correct query shape, then make that query safe, maintainable, and reasonably efficient.**

---

## 1. Start With the Result You Need

Do not begin with:

```text
"What SQL syntax should I use?"
```

Begin with:

```text
"What should one result row represent?"
```

Examples:

```text
one row per user
one row per order
one row per day
one row per product
one summary for the whole system
```

This is the **result grain**.

---

## 2. Grain Is the Core Mental Model

Suppose the requirement is:

> Show each user's total paid orders.

The result grain is:

```text
one row per user
```

That immediately suggests:

```sql
GROUP BY user_id
```

Then:

```text
COUNT paid orders
```

The query design follows the required result shape.

---

## 3. Example: Simple List

Requirement:

> Return the latest 20 active users.

Think:

```text
source
→ users

filter
→ active users

sort
→ newest first

limit
→ 20
```

Query:

```sql
SELECT
    id,
    name,
    email,
    created_at
FROM users
WHERE active = true
ORDER BY created_at DESC, id DESC
LIMIT 20;
```

The SQL is a translation of the requirement.

---

## 4. Query Design as a Pipeline

A useful mental model:

```text
FROM / JOIN
      ↓
WHERE
      ↓
GROUP BY
      ↓
HAVING
      ↓
SELECT
      ↓
ORDER BY
      ↓
LIMIT / OFFSET
```

This is not the exact SQL parser implementation, but it is a useful way to reason about query behavior.

Ask what each stage is supposed to accomplish.

---

## 5. FROM: Find the Source Data

Start by identifying where the required information lives.

Requirement:

> Show orders.

Source:

```sql
FROM orders
```

Requirement:

> Show order with customer name.

Sources:

```sql
FROM orders o
JOIN users u
    ON u.id = o.user_id
```

Do not add tables simply because they are available.

Every join should exist for a reason.

---

## 6. JOIN: Add Required Context

Suppose:

```text
orders
------
id
user_id
total
```

and:

```text
users
-----
id
name
```

Requirement:

> Show order ID and customer name.

Use:

```sql
SELECT
    o.id,
    u.name
FROM orders o
JOIN users u
    ON u.id = o.user_id;
```

The join adds information needed by the result.

---

## 7. Decide JOIN Type From the Requirement

Ask:

> Should rows without a matching record remain?

If no:

```sql
JOIN
```

If yes:

```sql
LEFT JOIN
```

Example:

> Show all users, including users with no orders.

```sql
SELECT
    u.id,
    u.name,
    COUNT(o.id) AS order_count
FROM users u
LEFT JOIN orders o
    ON o.user_id = u.id
GROUP BY u.id, u.name;
```

The requirement determines the join.

---

## 8. WHERE: Filter Individual Rows

Ask:

> Which source rows should participate?

Example:

> Paid orders only.

```sql
WHERE status = 'paid'
```

Example:

> Orders from the current reporting period.

```sql
WHERE created_at >= $1
  AND created_at < $2
```

Filter row-level conditions with `WHERE`.

---

## 9. Prefer Precise Time Ranges

For timestamps, prefer:

```sql
created_at >= $1
AND created_at < $2
```

instead of trying to construct an inclusive end-of-day timestamp.

Example:

```text
start = 2026-09-01 00:00
end   = 2026-10-01 00:00
```

Then:

```sql
WHERE created_at >= $1
  AND created_at < $2
```

This is easy to reason about and works naturally for adjacent ranges.

---

## 10. SELECT: Choose Only What the Consumer Needs

Avoid:

```sql
SELECT *
FROM users;
```

when the API only needs:

```text
id
name
email
```

Prefer:

```sql
SELECT
    id,
    name,
    email
FROM users;
```

Benefits include:

```text
clear contract
less data transferred
less accidental coupling
easier query review
```

---

## 11. ORDER BY: Define What "Latest" Means

This:

```sql
ORDER BY created_at DESC
```

may be enough, but timestamps can tie.

For deterministic ordering:

```sql
ORDER BY created_at DESC, id DESC
```

The second field is a tie-breaker.

This matters especially for:

```text
pagination
feeds
admin tables
APIs
```

---

## 12. LIMIT: Control Result Size

If the API needs 20 rows:

```sql
LIMIT 20
```

Do not let a list endpoint accidentally return millions of rows.

Production APIs should normally define a reasonable maximum page size.

---

## 13. OFFSET: Simple but Not Always Ideal

Basic pagination:

```sql
LIMIT 20
OFFSET 40
```

means:

```text
skip 40
return next 20
```

This is simple and often fine for small datasets.

For large/high-traffic datasets, keyset pagination can be more efficient and stable.

---

## 14. Keyset Pagination

Suppose ordering is:

```sql
ORDER BY created_at DESC, id DESC
```

The next page can use the last row's values:

```sql
WHERE (created_at, id) < ($1, $2)
ORDER BY created_at DESC, id DESC
LIMIT 20;
```

This avoids progressively larger offsets.

The important part is:

```text
stable ordering
+
matching continuation condition
```

---

## 15. Aggregation Query Thinking

Requirement:

> How many orders does each user have?

Translate:

```text
source
→ orders

group
→ user_id

metric
→ COUNT(*)
```

SQL:

```sql
SELECT
    user_id,
    COUNT(*) AS order_count
FROM orders
GROUP BY user_id;
```

Start from the metric and grain, not from syntax.

---

## 16. Filter Before Aggregation

Requirement:

> How many paid orders does each user have?

Translate:

```text
rows
→ paid orders

groups
→ user

metric
→ count
```

SQL:

```sql
SELECT
    user_id,
    COUNT(*) AS paid_order_count
FROM orders
WHERE status = 'paid'
GROUP BY user_id;
```

`WHERE` removes irrelevant rows before grouping.

---

## 17. Filter Groups After Aggregation

Requirement:

> Which users have at least 10 paid orders?

```sql
SELECT
    user_id,
    COUNT(*) AS paid_order_count
FROM orders
WHERE status = 'paid'
GROUP BY user_id
HAVING COUNT(*) >= 10;
```

Think:

```text
WHERE
→ which rows?

GROUP BY
→ which groups?

HAVING
→ which groups survive?
```

---

## 18. NULL Changes Query Reasoning

Requirement:

> Find users who do not have a phone number.

Correct:

```sql
WHERE phone IS NULL
```

Not:

```sql
WHERE phone = NULL
```

`NULL` represents an unknown/absent value, not an ordinary value.

---

## 19. NOT IN and NULL

Be careful with:

```sql
WHERE user_id NOT IN (...)
```

If the list contains NULL, SQL's three-valued logic can produce surprising results.

When the requirement is:

> Rows for which no matching record exists

`NOT EXISTS` is often clearer:

```sql
WHERE NOT EXISTS (
    SELECT 1
    FROM blocked_users b
    WHERE b.user_id = u.id
)
```

Choose the query form based on the meaning you need.

---

## 20. EXISTS for Existence Checks

Requirement:

> Return users who have at least one paid order.

```sql
SELECT
    u.id,
    u.name
FROM users u
WHERE EXISTS (
    SELECT 1
    FROM orders o
    WHERE o.user_id = u.id
      AND o.status = 'paid'
);
```

This expresses:

```text
Does a matching row exist?
```

rather than:

```text
Join everything and then remove duplicates.
```

Use the form that matches the requirement.

---

## 21. JOIN vs EXISTS

Requirement:

> Return user details plus order information.

Use a `JOIN`.

Requirement:

> Return users only if a matching order exists.

`EXISTS` can be a natural fit.

Mental distinction:

```text
JOIN
→ I need data from the other table.

EXISTS
→ I only need to know whether a match exists.
```

This often makes queries easier to reason about.

---

## 22. Avoid Accidental Duplicate Rows

Suppose:

```text
users
↓
orders
```

One user can have many orders.

This query:

```sql
SELECT
    u.id,
    u.name
FROM users u
JOIN orders o
    ON o.user_id = u.id;
```

can return the same user multiple times.

If the requirement is:

> one row per user

then the query shape must account for that.

Possible approaches:

```text
EXISTS
GROUP BY
DISTINCT
```

Choose based on the actual requirement.

---

## 23. DISTINCT Is Not a Universal Fix

This:

```sql
SELECT DISTINCT u.id, u.name
FROM users u
JOIN orders o
    ON o.user_id = u.id;
```

can remove duplicate result rows.

But adding `DISTINCT` blindly can hide a query-design problem.

Ask:

> Why did the join create multiple rows?

If you only need existence, `EXISTS` may better represent the requirement.

---

## 24. The Double-Counting Problem

Requirement:

> Calculate order revenue.

Suppose an order has multiple items.

If the query joins:

```text
orders
+
order_items
```

then one order can become several rows.

This can make:

```sql
SUM(o.total)
```

incorrect.

Before aggregating, identify:

```text
What does one row represent?
```

If it is:

```text
one row per item
```

you cannot blindly sum an order-level value.

---

## 25. Fix the Grain Instead of Hiding It

Possible approaches:

```text
aggregate orders before joining item data
aggregate items separately
join already-aggregated subqueries
```

The correct solution depends on the required output.

Avoid using:

```sql
SUM(DISTINCT ...)
```

as a generic fix.

Distinct values are not the same as distinct business entities.

---

## 26. Build Complex Queries Incrementally

For a complicated requirement, do not write the final query immediately.

Start:

```sql
SELECT ...
FROM ...
```

Then add:

```text
JOIN
WHERE
GROUP BY
HAVING
ORDER BY
LIMIT
```

After each important step, verify:

```text
row count
row grain
duplicates
NULL behavior
```

This makes debugging much easier.

---

## 27. Validate the Base Dataset First

Suppose the final query is:

```sql
SELECT
    u.id,
    COUNT(o.id)
FROM users u
LEFT JOIN orders o
    ON o.user_id = u.id
WHERE o.status = 'paid'
GROUP BY u.id;
```

Notice that the `WHERE` condition on the joined table can change the effect of the `LEFT JOIN`.

If users with zero paid orders must remain, move the condition into the join:

```sql
SELECT
    u.id,
    COUNT(o.id)
FROM users u
LEFT JOIN orders o
    ON o.user_id = u.id
   AND o.status = 'paid'
GROUP BY u.id;
```

The placement of conditions matters.

---

## 28. ON vs WHERE With LEFT JOIN

Compare:

```sql
LEFT JOIN orders o
    ON o.user_id = u.id
WHERE o.status = 'paid'
```

with:

```sql
LEFT JOIN orders o
    ON o.user_id = u.id
   AND o.status = 'paid'
```

The first can remove users whose joined order is NULL.

The second keeps the user and only joins paid orders.

Production lesson:

> For `LEFT JOIN`, condition placement can change the result set.

---

## 29. Think in Business Grain

Common grains:

```text
one row per customer
one row per order
one row per order item
one row per product
one row per day
one row per month
```

When joining tables, grain can change.

Example:

```text
user
↓
orders
```

changes:

```text
one row per user
```

into potentially:

```text
many rows per user
```

unless you aggregate or otherwise constrain it.

---

## 30. Query Requirements as a Checklist

When given a backend requirement, extract:

```text
1. Source tables
2. Required relationships
3. Result grain
4. Row filters
5. Grouping
6. Aggregate metrics
7. Group filters
8. Sort order
9. Pagination
10. NULL behavior
11. Concurrency/write requirements
```

This turns vague requirements into a query plan.

---

## 31. Example: User Dashboard

Requirement:

> Show each active user with paid order count and paid revenue for September.

Break it down:

```text
source
→ users + orders

relationship
→ users.id = orders.user_id

base filter
→ users.active = true
→ paid orders
→ September range

result grain
→ one row per user

metrics
→ COUNT
→ SUM
```

Possible query:

```sql
SELECT
    u.id,
    u.name,
    COUNT(o.id) AS paid_order_count,
    COALESCE(SUM(o.total), 0) AS paid_revenue
FROM users u
LEFT JOIN orders o
    ON o.user_id = u.id
   AND o.status = 'paid'
   AND o.created_at >= $1
   AND o.created_at < $2
WHERE u.active = true
GROUP BY u.id, u.name
ORDER BY u.id;
```

The query follows the requirement's grain.

---

## 32. Example: Top Customers

Requirement:

> Return the top 10 customers by paid revenue.

Thinking:

```text
orders
→ paid only
→ group by user
→ sum total
→ sort revenue descending
→ top 10
```

SQL:

```sql
SELECT
    user_id,
    SUM(total) AS revenue
FROM orders
WHERE status = 'paid'
GROUP BY user_id
ORDER BY revenue DESC, user_id DESC
LIMIT 10;
```

This is a standard reporting pattern.

---

## 33. Example: Users Without Orders

Requirement:

> Find active users who have never placed an order.

Use existence logic:

```sql
SELECT
    u.id,
    u.name
FROM users u
WHERE u.active = true
  AND NOT EXISTS (
      SELECT 1
      FROM orders o
      WHERE o.user_id = u.id
  );
```

The query directly expresses:

```text
active user
AND no matching order exists
```

---

## 34. Example: Latest Record Per User

Requirement:

> Return each user's latest order.

This is more advanced because the requirement is:

```text
one row per user
+
latest order
```

A common PostgreSQL solution is `DISTINCT ON`:

```sql
SELECT DISTINCT ON (user_id)
    user_id,
    id,
    total,
    created_at
FROM orders
ORDER BY user_id, created_at DESC, id DESC;
```

This is PostgreSQL-specific.

The important lesson is not memorizing the syntax.

It is recognizing the query shape:

> **Pick one row from each group according to an ordering rule.**

---

## 35. Window Functions as Another Shape

The same class of problem can be solved with a window function:

```sql
SELECT
    user_id,
    id,
    total,
    created_at
FROM (
    SELECT
        user_id,
        id,
        total,
        created_at,
        ROW_NUMBER() OVER (
            PARTITION BY user_id
            ORDER BY created_at DESC, id DESC
        ) AS rn
    FROM orders
) x
WHERE rn = 1;
```

Window functions are useful when you need row-level data while also reasoning within groups.

They belong to the next level of query design, but recognizing this pattern is useful.

---

## 36. Query Correctness Before Performance

Use this priority:

```text
1. Correct result
2. Correct business meaning
3. Safe behavior
4. Maintainability
5. Performance optimization
```

A fast query with the wrong metric is still wrong.

---

## 37. Then Check Performance

Once the query is correct, investigate:

```text
row counts
filters
joins
indexes
sorts
aggregation
pagination
execution plan
```

Use:

```sql
EXPLAIN
```

and, when appropriate:

```sql
EXPLAIN ANALYZE
```

Do not optimize based only on intuition.

---

## 38. EXPLAIN Is a Diagnostic Tool

Example:

```sql
EXPLAIN
SELECT
    id,
    name
FROM users
WHERE email = $1;
```

It helps you understand how PostgreSQL plans to execute the query.

For actual runtime behavior:

```sql
EXPLAIN ANALYZE
SELECT ...
```

Use care with `EXPLAIN ANALYZE` for write statements because it actually executes them.

---

## 39. Index Thinking

Do not ask:

> "Should every WHERE column have an index?"

Instead ask:

```text
Which queries matter?
Which filters are selective?
Which columns are used for joins?
Which columns are used for ordering?
What is the actual workload?
```

Indexes are workload-driven.

---

## 40. Query Shape and Index Shape

Suppose the common query is:

```sql
SELECT
    id,
    created_at
FROM orders
WHERE user_id = $1
ORDER BY created_at DESC
LIMIT 20;
```

The query shape suggests that an index involving:

```text
user_id
created_at
```

may be useful.

The exact index design should be validated against the real workload and query plan.

---

## 41. Avoid Premature Optimization

Do not make every query complicated because:

```text
"Maybe it will be faster."
```

Start with:

```text
correct
clear
parameterized
bounded
```

Then measure.

Production engineering is about solving actual bottlenecks, not optimizing imaginary ones.

---

## 42. Query Result Size Matters

Even a correct query can be operationally bad if it returns too much data.

Bad API behavior:

```text
GET /orders
→ millions of rows
```

Better:

```text
pagination
projection
filters
reasonable limits
```

The database and API should have intentional result boundaries.

---

## 43. Avoid N+1 Query Patterns

Suppose:

```text
SELECT users
```

returns 100 users.

Then the application runs:

```text
SELECT orders WHERE user_id = 1
SELECT orders WHERE user_id = 2
...
SELECT orders WHERE user_id = 100
```

That is an N+1 pattern.

Depending on the use case, consider:

```text
JOIN
IN
batch query
aggregation
```

The best solution depends on the required result shape.

---

## 44. But Do Not Fear Multiple Queries

N+1 is bad when it creates unnecessary repeated round trips.

That does not mean:

> "One SQL query is always better."

Sometimes separate queries are clearer and more efficient.

Choose based on:

```text
result shape
data volume
latency
database load
maintainability
measured performance
```

---

## 45. Query Thinking for API Development

For an endpoint:

```text
GET /orders?status=paid&page=2
```

translate the API contract into:

```text
table
→ orders

filters
→ status

sort
→ deterministic order

pagination
→ limit + cursor/offset
```

Then produce parameterized SQL.

The API layer should not dictate unsafe SQL construction.

---

## 46. Dynamic Filtering

Suppose the API supports:

```text
status
customer
date range
```

Build conditions deliberately.

Conceptually:

```text
WHERE 1=1
+ optional trusted conditions
+ bound values
```

The exact implementation is application-specific.

Do not concatenate raw user input into SQL.

---

## 47. Dynamic ORDER BY

Values can be parameters:

```sql
WHERE status = $1
```

but arbitrary identifiers generally cannot be treated the same way.

If the API supports:

```text
sort=name
sort=created_at
```

map them through a whitelist:

```text
name       → users.name
created_at → users.created_at
```

Never directly insert an untrusted column name into SQL.

---

## 48. Query Thinking for Writes

For:

```text
PATCH /orders/123
```

ask:

```text
What row?
What fields?
What state must it currently have?
What should happen if it does not?
Does this need a transaction?
Could two requests race?
```

Then design the SQL.

Example:

```sql
UPDATE orders
SET status = 'paid'
WHERE id = $1
  AND status = 'pending'
RETURNING id, status;
```

This is safer than:

```text
SELECT status
↓
application checks
↓
UPDATE
```

when the state check can be expressed atomically.

---

## 49. Query Thinking for Transactions

If a business operation needs:

```text
write A
+
write B
+
write C
```

ask:

> Must all three succeed together?

If yes, consider:

```text
transaction
```

Then define:

```text
BEGIN
↓
A
↓
B
↓
C
↓
COMMIT
```

On required failure:

```text
ROLLBACK
```

Do not use transactions merely because multiple queries exist.

---

## 50. Query Thinking for Concurrency

Ask:

> What happens if two requests execute this at the same time?

This is especially important for:

```text
balances
inventory
state transitions
unique resources
counters
payments
```

A query that is correct in isolation can still be incorrect under concurrency.

Later PostgreSQL materials cover:

```text
transactions
isolation
locking
```

in more depth.

---

## 51. Production Query Review

Before approving a query, ask:

```text
What does one result row represent?

Which table provides the base rows?

Which joins are required?

Can a join multiply rows?

Which conditions filter rows?

Which conditions filter groups?

How are NULLs handled?

Is the ordering deterministic?

Is the result bounded?

Are values parameterized?

Could the query return duplicate business entities?

Could an aggregate double-count?

Does the query need an index?

Does the query need a transaction?

What happens under concurrency?
```

These questions catch many real production bugs.

---

## 52. A Practical Query-Building Workflow

Use this workflow:

```text
1. Write the requirement in plain language.
2. Define the result grain.
3. Identify source tables.
4. Identify required relationships.
5. Add row filters.
6. Decide whether grouping is required.
7. Add aggregate metrics.
8. Add group filters.
9. Define deterministic ordering.
10. Add pagination/limits.
11. Handle NULL explicitly.
12. Parameterize values.
13. Check duplicate/double-counting risks.
14. Review performance.
15. Test against realistic data.
```

This is much more reliable than writing SQL by trial and error.

---

## 53. Production Test Cases

For important queries, test more than the happy path.

Think about:

```text
zero rows
one row
many rows
NULL values
duplicate relationships
missing related records
boundary timestamps
large datasets
concurrent writes
```

For aggregation:

```text
no matching rows
one matching row
multiple matching rows
users with zero children
```

For pagination:

```text
empty page
last page
duplicate timestamps
new rows arriving between requests
```

The query should be validated against the cases its business meaning allows.

---

## 54. Readability Is a Production Feature

Prefer:

```sql
SELECT
    o.id,
    o.user_id,
    o.total,
    o.status,
    o.created_at
FROM orders o
WHERE o.status = $1
ORDER BY o.created_at DESC, o.id DESC
LIMIT $2;
```

over compressed SQL that is difficult to review.

Production SQL is maintained by humans.

Clear aliases, explicit columns, and meaningful names reduce mistakes.

---

## 55. Avoid Clever SQL Without a Reason

SQL has many ways to express a requirement.

The best query is not necessarily the shortest.

Prefer the query that is:

```text
correct
clear
maintainable
compatible with the workload
```

Use advanced PostgreSQL features when they solve a real problem.

---

## 56. ORM Does Not Remove Query Thinking

An ORM may generate SQL, but the database still receives SQL.

You still need to understand:

```text
JOIN
WHERE
GROUP BY
indexes
pagination
transactions
N+1
locking
```

If an ORM query is slow or incorrect, SQL knowledge is still required to diagnose it.

---

## 57. Query Thinking and Production Debugging

When a query produces unexpected results:

```text
1. Inspect the base table.
2. Check the join.
3. Check row multiplication.
4. Check WHERE.
5. Check NULL behavior.
6. Check grouping.
7. Check aggregation.
8. Check HAVING.
9. Check ordering.
10. Check pagination.
```

Do not immediately rewrite the entire query.

Find the stage where the result first becomes incorrect.

---

## 58. Example Debugging Strategy

Expected:

```text
10 users
```

Actual:

```text
35 rows
```

Possible causes:

```text
one-to-many JOIN
```

Inspect:

```sql
SELECT
    u.id,
    COUNT(*)
FROM users u
JOIN orders o
    ON o.user_id = u.id
GROUP BY u.id;
```

If users have multiple orders, the join naturally creates multiple rows.

The next question becomes:

> Do I actually need order rows, or only order existence/count?

That determines whether to use:

```text
EXISTS
COUNT
GROUP BY
```

---

## 59. Query Thinking: A Compact Example

Requirement:

> Get the top 20 active customers by paid revenue this month.

Translate:

```text
source
→ users + orders

row filters
→ active users
→ paid orders
→ current month

grain
→ one row per customer

metric
→ SUM(order.total)

sort
→ revenue DESC
→ customer ID DESC as tie-breaker

limit
→ 20
```

Possible SQL:

```sql
SELECT
    u.id,
    u.name,
    SUM(o.total) AS revenue
FROM users u
JOIN orders o
    ON o.user_id = u.id
WHERE u.active = true
  AND o.status = 'paid'
  AND o.created_at >= $1
  AND o.created_at < $2
GROUP BY u.id, u.name
ORDER BY revenue DESC, u.id DESC
LIMIT 20;
```

The query is easier to write once the requirement has been converted into a shape.

---

## 60. The Most Important Habit

When you see a SQL problem, ask:

```text
What is my input row?

What is my output row?

What changes the row count?

Where should filtering happen?

What relationships can multiply rows?

What metric am I calculating?

What does NULL mean here?

How large can the result become?

What happens under concurrency?
```

These questions are more valuable than memorizing isolated SQL tricks.

---

## 61. Production Checklist

Before shipping a production query:

```text
□ Business requirement is precise
□ Result grain is defined
□ Source tables are correct
□ Joins are intentional
□ Join cardinality is understood
□ Row filters are in WHERE/ON appropriately
□ GROUP BY matches the desired grain
□ HAVING is used only for group-level filtering
□ NULL behavior is understood
□ Aggregates cannot double-count unexpectedly
□ SELECT returns only required columns
□ Ordering is deterministic
□ Pagination is bounded and stable
□ Values are parameterized
□ Dynamic identifiers are whitelisted
□ Query does not create unnecessary N+1 calls
□ Transaction boundary is correct for writes
□ Concurrency behavior is understood
□ Query has been tested with edge cases
□ Performance has been checked with realistic data
□ EXPLAIN is available for investigation
```

---

## 62. Final Mental Model

Production SQL is not mainly about remembering syntax.

Think:

```text
Business requirement
        ↓
Result grain
        ↓
Source data
        ↓
Relationships
        ↓
Row filtering
        ↓
Grouping / aggregation
        ↓
Group filtering
        ↓
Ordering
        ↓
Pagination
        ↓
Correctness checks
        ↓
Performance checks
        ↓
Application
```

For writes:

```text
Business operation
        ↓
Target rows
        ↓
Safe condition
        ↓
Parameterized write
        ↓
Constraints
        ↓
RETURNING / affected rows
        ↓
Transaction if required
        ↓
Concurrency review
```

---

## 63. What to Remember

1. Start from the business requirement, not SQL syntax.
2. Define what one result row represents.
3. Always understand the grain of the data at each query stage.
4. Use joins only when you need their data or relationship.
5. Use `EXISTS` when the requirement is primarily existence.
6. Remember that one-to-many joins multiply rows.
7. Never assume an aggregate is correct until you understand row multiplication.
8. `WHERE` filters rows; `HAVING` filters groups.
9. Condition placement matters with `LEFT JOIN`.
10. Handle NULL intentionally.
11. Make ordering deterministic for production pagination.
12. Bound API result sizes.
13. Parameterize values.
14. Whitelist dynamic identifiers such as sort fields.
15. Correctness comes before optimization.
16. Use `EXPLAIN` when performance needs investigation.
17. Do not assume one query is always better than multiple queries.
18. ORM usage still requires SQL knowledge.
19. Think about concurrency for state-changing operations.
20. Good SQL is SQL whose result shape, safety, and performance you can explain.

---

# End of SQL Fundamentals

The next PostgreSQL section moves from query fundamentals into schema design and data integrity:

```text
02-schema-and-constraints
```

The priority remains production-first:

```text
correctness
→ integrity
→ maintainability
→ performance
```
