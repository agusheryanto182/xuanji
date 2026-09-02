# 05 — Aggregation and Grouping

## Goal

Learn the aggregation patterns commonly used in production backend systems for dashboards, reports, metrics, and summaries.

Focus:

- `COUNT`
- `SUM`
- `AVG`
- `MIN` / `MAX`
- `GROUP BY`
- `HAVING`
- `WHERE` vs `HAVING`
- NULL behavior
- conditional aggregation
- aggregation with joins
- avoiding double counting

The goal is:

> **Turn many database rows into useful business-level numbers without producing incorrect results.**

---

## 1. What Is Aggregation?

Aggregation combines multiple rows into a summarized result.

Example:

```sql
SELECT COUNT(*)
FROM orders;
```

Instead of returning every order, the query returns one number:

```text
total orders
```

Common aggregate functions:

```text
COUNT
SUM
AVG
MIN
MAX
```

Aggregation is common in:

```text
dashboards
reports
admin pages
analytics
business metrics
```

---

## 2. COUNT

`COUNT(*)` counts rows.

```sql
SELECT COUNT(*)
FROM orders;
```

Example:

```text
orders
------
1
2
3
4
5
```

Result:

```text
5
```

### Production use

Use it when the requirement is:

> How many rows are there?

---

## 3. COUNT(column)

`COUNT(column)` counts non-NULL values in that column.

```sql
SELECT COUNT(phone)
FROM users;
```

If:

```text
phone
------
111
NULL
333
```

the result is:

```text
2
```

Compare:

```sql
COUNT(*)
```

which would return:

```text
3
```

### Remember

```text
COUNT(*)
→ rows

COUNT(column)
→ non-NULL column values
```

---

## 4. COUNT(DISTINCT)

Use `DISTINCT` when the requirement is to count unique values.

```sql
SELECT COUNT(DISTINCT user_id)
FROM orders;
```

This answers:

> How many different users have placed orders?

It is different from:

```sql
COUNT(*)
```

which answers:

> How many orders exist?

The metric definition matters.

---

## 5. SUM

`SUM` adds numeric values.

```sql
SELECT SUM(total)
FROM orders;
```

This can answer:

> What is the total order value?

A common production query:

```sql
SELECT SUM(total)
FROM orders
WHERE status = 'paid';
```

The filtering happens before aggregation.

---

## 6. AVG

`AVG` calculates the average of numeric values.

```sql
SELECT AVG(total)
FROM orders;
```

For a filtered metric:

```sql
SELECT AVG(total)
FROM orders
WHERE status = 'paid';
```

### Production note

Always define what the average represents.

For example:

```text
average order value
```

is different from:

```text
average value per customer
```

The SQL can look similar while the business metric is different.

---

## 7. MIN and MAX

`MIN` returns the smallest value.

```sql
SELECT MIN(total)
FROM orders;
```

`MAX` returns the largest.

```sql
SELECT MAX(total)
FROM orders;
```

Common uses:

```text
lowest price
highest order
earliest timestamp
latest timestamp
```

Example:

```sql
SELECT MAX(created_at)
FROM orders;
```

This can answer:

> When was the latest order created?

---

## 8. Aggregation Without GROUP BY

Without `GROUP BY`, the aggregate usually summarizes the whole matching result set.

```sql
SELECT
    COUNT(*) AS order_count,
    SUM(total) AS revenue
FROM orders
WHERE status = 'paid';
```

Result:

```text
order_count | revenue
------------+--------
125         | 500000
```

One result row represents the whole filtered dataset.

---

## 9. GROUP BY

`GROUP BY` divides rows into groups before aggregation.

Example:

```sql
SELECT
    user_id,
    COUNT(*) AS order_count
FROM orders
GROUP BY user_id;
```

Result conceptually:

```text
user_id | order_count
--------+------------
1       | 3
2       | 5
3       | 1
```

The query answers:

> How many orders does each user have?

---

## 10. GROUP BY Changes the Result Shape

Without grouping:

```text
many orders
↓
one summary
```

With:

```sql
GROUP BY user_id
```

you get:

```text
one summary per user
```

This is the most important mental model for `GROUP BY`.

---

## 11. Multiple GROUP BY Columns

You can group by more than one column.

```sql
SELECT
    status,
    user_id,
    COUNT(*) AS order_count
FROM orders
GROUP BY status, user_id;
```

Now each group represents a unique combination:

```text
status + user_id
```

This is useful for metrics such as:

```text
orders per user per status
sales per store per day
tickets per team per priority
```

---

## 12. GROUP BY Date

A common reporting requirement is:

> How many orders were created each day?

One PostgreSQL approach:

```sql
SELECT
    created_at::date AS order_date,
    COUNT(*) AS order_count
FROM orders
GROUP BY created_at::date
ORDER BY order_date;
```

This creates:

```text
date       | order_count
-----------+------------
2026-09-01 | 120
2026-09-02 | 145
```

For production reporting, also think about timezone requirements before converting timestamps to dates.

---

## 13. WHERE vs GROUP BY

`WHERE` filters rows **before** grouping.

Example:

```sql
SELECT
    user_id,
    COUNT(*) AS order_count
FROM orders
WHERE status = 'paid'
GROUP BY user_id;
```

Think:

```text
all orders
↓
keep paid orders
↓
group by user
↓
count each group
```

This is the correct pattern when the filter applies to individual rows.

---

## 14. HAVING

`HAVING` filters groups after aggregation.

Example:

```sql
SELECT
    user_id,
    COUNT(*) AS order_count
FROM orders
GROUP BY user_id
HAVING COUNT(*) >= 10;
```

This means:

> Return users who have at least 10 orders.

The difference is:

```text
WHERE
→ filters rows

HAVING
→ filters groups
```

---

## 15. WHERE + HAVING

You can use both.

Example:

> Find users with at least 10 paid orders.

```sql
SELECT
    user_id,
    COUNT(*) AS order_count
FROM orders
WHERE status = 'paid'
GROUP BY user_id
HAVING COUNT(*) >= 10;
```

Mental model:

```text
all orders
↓
WHERE status = paid
↓
GROUP BY user
↓
COUNT
↓
HAVING count >= 10
```

This pattern is common in reporting queries.

---

## 16. Why WHERE Usually Comes First

Suppose the requirement is:

> Count paid orders per user.

Correct:

```sql
SELECT
    user_id,
    COUNT(*) AS order_count
FROM orders
WHERE status = 'paid'
GROUP BY user_id;
```

Do not aggregate all orders and then try to remove unpaid rows afterward.

Filtering early usually expresses the requirement more clearly and can reduce the amount of data that needs to be grouped.

---

## 17. Conditional Aggregation

Sometimes you need several metrics in one query.

Example:

```sql
SELECT
    COUNT(*) AS total_orders,
    COUNT(*) FILTER (WHERE status = 'paid') AS paid_orders,
    COUNT(*) FILTER (WHERE status = 'cancelled') AS cancelled_orders
FROM orders;
```

This can return:

```text
total_orders | paid_orders | cancelled_orders
-------------+-------------+----------------
1000         | 800         | 50
```

This is useful for dashboards.

---

## 18. Conditional Aggregation With CASE

A portable alternative is:

```sql
SELECT
    COUNT(*) AS total_orders,
    SUM(CASE WHEN status = 'paid' THEN 1 ELSE 0 END) AS paid_orders
FROM orders;
```

PostgreSQL's `FILTER` syntax is often cleaner:

```sql
COUNT(*) FILTER (WHERE status = 'paid')
```

Both patterns are useful to recognize.

---

## 19. SUM With Conditions

You can calculate conditional totals.

Example:

```sql
SELECT
    SUM(total) FILTER (WHERE status = 'paid') AS paid_revenue,
    SUM(total) FILTER (WHERE status = 'refunded') AS refunded_value
FROM orders;
```

This can produce multiple business metrics from one dataset.

Always verify that the metric definitions match the business requirement.

---

## 20. Aggregate NULL Behavior

Aggregates such as:

```text
SUM
AVG
MIN
MAX
```

normally ignore NULL input values.

Example:

```text
total
-----
100
200
NULL
```

Then:

```sql
AVG(total)
```

uses:

```text
100
200
```

not a third numeric value.

This matters when nullable columns represent incomplete data.

---

## 21. Aggregate With No Matching Rows

Consider:

```sql
SELECT SUM(total)
FROM orders
WHERE user_id = $1;
```

If there are no matching rows, `SUM` can return:

```text
NULL
```

If the business meaning should be zero:

```sql
SELECT COALESCE(SUM(total), 0)
FROM orders
WHERE user_id = $1;
```

Do not blindly use `COALESCE`.

Ask:

> Is zero actually the correct meaning when there is no data?

---

## 22. COUNT and No Matching Rows

`COUNT(*)` behaves differently.

```sql
SELECT COUNT(*)
FROM orders
WHERE user_id = $1;
```

If no rows match, the result is:

```text
0
```

This is one reason count metrics are often simpler to consume than sum/average metrics.

---

## 23. GROUP BY and NULL

NULL can form a group.

Example:

```sql
SELECT
    category_id,
    COUNT(*)
FROM products
GROUP BY category_id;
```

If some products have:

```text
category_id = NULL
```

those rows belong to the NULL group.

This can be useful, but it may also expose incomplete data.

Ask what NULL means before interpreting the metric.

---

## 24. Aggregation With JOIN

A common requirement:

> Count orders for each user.

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

`LEFT JOIN` keeps users who have zero orders.

Those users receive:

```text
order_count = 0
```

because:

```sql
COUNT(o.id)
```

does not count the NULL produced by the missing order.

---

## 25. COUNT(\*) vs COUNT(joined_column)

This distinction is critical with `LEFT JOIN`.

Compare:

```sql
COUNT(*)
```

with:

```sql
COUNT(o.id)
```

For a user with no matching order:

```text
COUNT(*)   → 1
COUNT(o.id) → 0
```

Why?

The `LEFT JOIN` still produces one result row for the user, but `o.id` is NULL.

### Production rule

> When counting optional joined records, `COUNT(joined_table.id)` is often what you want.

---

## 26. The Double-Counting Problem

Aggregation with multiple one-to-many joins can produce incorrect totals.

Example:

```text
users
  ↓
orders
  ↓
order_items
```

If you join both:

```sql
users
JOIN orders
JOIN order_items
```

one order can produce multiple rows because it has multiple items.

If you then:

```sql
SUM(order.total)
```

the same order total can be counted multiple times.

This is one of the most important production aggregation bugs.

---

## 27. Why Double Counting Happens

Suppose:

```text
Order 1
total = 100

items:
A
B
C
```

After joining:

```text
Order 1 | A | 100
Order 1 | B | 100
Order 1 | C | 100
```

Then:

```sql
SUM(order.total)
```

produces:

```text
300
```

instead of:

```text
100
```

The SQL can be syntactically correct while the business result is wrong.

---

## 28. Preventing Double Counting

Before aggregating, understand the row shape.

Possible solutions include:

```text
aggregate at the correct level first
separate aggregates
pre-aggregate child data
use DISTINCT carefully
```

For example, if the metric is order revenue, aggregate orders before joining to item-level data when appropriate.

Do not use `SUM(DISTINCT order.total)` as a generic fix.

If two different orders legitimately have the same total, `DISTINCT` can undercount.

---

## 29. Aggregate at the Correct Grain

A useful concept is **grain**.

Ask:

> What does one row represent at this point in the query?

Examples:

```text
one row per order
one row per order item
one row per user
one row per day
```

If your metric is:

```text
revenue per order
```

but your query currently has:

```text
one row per order item
```

you may accidentally count the order multiple times.

Production habit:

> Know the grain of your result before applying an aggregate.

---

## 30. Aggregating Per User

Example:

```sql
SELECT
    user_id,
    COUNT(*) AS order_count,
    SUM(total) AS revenue
FROM orders
GROUP BY user_id;
```

This produces:

```text
one row per user
```

with:

```text
number of orders
total order value
```

This is a common dashboard/report query.

---

## 31. Aggregating Per Status

Example:

```sql
SELECT
    status,
    COUNT(*) AS order_count
FROM orders
GROUP BY status
ORDER BY order_count DESC;
```

This produces a status breakdown:

```text
status     | order_count
-----------+------------
paid       | 800
pending    | 120
cancelled  | 80
```

Useful for:

```text
admin dashboards
operational monitoring
workflow metrics
```

---

## 32. Aggregating Per Day

Example:

```sql
SELECT
    created_at::date AS day,
    COUNT(*) AS order_count,
    SUM(total) AS revenue
FROM orders
WHERE created_at >= $1
  AND created_at < $2
GROUP BY created_at::date
ORDER BY day;
```

This pattern is common in reports.

The date range should usually be bounded before grouping so the query does not process unrelated historical data.

---

## 33. Aggregation With Time Zones

If `created_at` is a timestamp with time zone, converting it to a date depends on the session/timezone context.

For business reports, define which timezone "day" means.

For example:

```text
business day = Asia/Jakarta
```

may differ from:

```text
UTC day
```

This can change daily metrics around midnight.

### Production point

> Timezone is part of the metric definition.

---

## 34. HAVING With Multiple Conditions

You can filter groups using multiple conditions.

```sql
SELECT
    user_id,
    COUNT(*) AS order_count,
    SUM(total) AS revenue
FROM orders
GROUP BY user_id
HAVING COUNT(*) >= 5
   AND SUM(total) >= 1000;
```

This means:

```text
at least 5 orders
AND
at least 1000 total revenue
```

The conditions apply to groups, not individual rows.

---

## 35. WHERE vs HAVING: Simple Rule

Use:

```text
WHERE
```

when the condition describes individual rows.

Use:

```text
HAVING
```

when the condition describes an aggregate/group.

Example:

```text
status = 'paid'
→ WHERE

COUNT(*) >= 10
→ HAVING
```

This simple distinction solves many query-design problems.

---

## 36. Aggregation and ORDER BY

You can sort aggregated results.

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
group by user
↓
count orders
↓
highest count first
↓
top 10
```

This is a common production dashboard pattern.

---

## 37. Aggregation and Pagination

Aggregated reports can also be paginated.

Example:

```sql
SELECT
    user_id,
    COUNT(*) AS order_count
FROM orders
GROUP BY user_id
ORDER BY order_count DESC, user_id DESC
LIMIT 20
OFFSET 0;
```

Notice the deterministic tie-breaker:

```text
order_count DESC
user_id DESC
```

If aggregated results are paginated, ordering still needs to be stable.

---

## 38. Aggregation in APIs

Suppose an API needs:

```text
total orders
paid orders
cancelled orders
revenue
```

Instead of making many separate queries, conditional aggregation may combine related metrics:

```sql
SELECT
    COUNT(*) AS total_orders,
    COUNT(*) FILTER (WHERE status = 'paid') AS paid_orders,
    COUNT(*) FILTER (WHERE status = 'cancelled') AS cancelled_orders,
    COALESCE(SUM(total) FILTER (WHERE status = 'paid'), 0) AS revenue
FROM orders;
```

Whether combining metrics is better depends on workload and query complexity, but it can reduce unnecessary round trips.

---

## 39. Aggregation Performance

Aggregation can become expensive when processing large datasets.

Important factors include:

```text
number of rows
filter selectivity
indexes
group cardinality
joins
sorting
memory
```

A filter such as:

```sql
WHERE created_at >= $1
```

can reduce the amount of historical data that needs to be aggregated.

Do not assume aggregation is cheap just because the final result has one row.

---

## 40. Indexes and Aggregation

Indexes can help filtering before aggregation.

For example:

```sql
SELECT COUNT(*)
FROM orders
WHERE user_id = $1
  AND status = 'paid';
```

The right index may make locating matching rows more efficient.

But indexes do not automatically make every aggregate fast.

Always consider the full query and actual execution plan.

---

## 41. COUNT(\*) and Indexes

Do not assume:

```sql
SELECT COUNT(*)
FROM huge_table;
```

is always an O(1) metadata lookup.

PostgreSQL generally needs to account for visible rows according to MVCC and query semantics.

For exact counts over large tables, the database may need substantial work.

If an application needs approximate counts, that is a separate design decision.

---

## 42. Aggregation and Business Definitions

A technically correct query can still produce the wrong business metric.

Example:

> Monthly revenue

Could mean:

```text
all orders
paid orders
completed orders
orders excluding refunds
recognized revenue
```

Before writing SQL, define the metric.

Then translate that definition into:

```text
filters
grouping
aggregation
```

Production rule:

> SQL correctness starts with a precise business definition.

---

## 43. Common Mistake: WHERE vs HAVING

Wrong mental model:

```text
HAVING filters rows
```

Correct:

```text
WHERE
→ rows

HAVING
→ groups
```

Example:

```sql
WHERE status = 'paid'
HAVING COUNT(*) >= 10
```

This is a clear separation of responsibilities.

---

## 44. Common Mistake: COUNT(\*) After LEFT JOIN

For:

```sql
users
LEFT JOIN orders
```

this:

```sql
COUNT(*)
```

counts the user's result row even when there is no order.

Usually you want:

```sql
COUNT(orders.id)
```

when measuring the number of matching orders.

---

## 45. Common Mistake: Double Counting After JOIN

If one parent row joins to many child rows:

```text
order
↓
many items
```

then:

```sql
SUM(order.total)
```

can count the same order multiple times.

Always inspect the result grain before aggregating.

---

## 46. Common Mistake: `SUM(DISTINCT ...)` as a Quick Fix

This looks tempting:

```sql
SUM(DISTINCT order.total)
```

But it can be wrong.

Suppose:

```text
Order A → 100
Order B → 100
```

`SUM(DISTINCT order.total)` returns:

```text
100
```

while the real revenue is:

```text
200
```

Fix the query's grain instead of blindly adding `DISTINCT`.

---

## 47. Common Mistake: Ignoring NULL

Remember:

```text
COUNT(*)
```

counts rows.

But:

```text
COUNT(column)
```

ignores NULL.

Also:

```text
SUM
AVG
MIN
MAX
```

generally ignore NULL inputs.

Understand what NULL means before interpreting the result.

---

## 48. Common Mistake: Forgetting the Timezone

A report grouped by:

```sql
created_at::date
```

may produce different days depending on timezone.

For business reporting, define the intended timezone explicitly.

---

## 49. Common Mistake: Aggregating Too Much Data

This:

```sql
SELECT
    created_at::date,
    COUNT(*)
FROM orders
GROUP BY created_at::date;
```

may process the entire orders table.

If the report only needs the current month, add the appropriate range:

```sql
WHERE created_at >= $1
  AND created_at < $2
```

Bound the data before aggregation when possible.

---

## 50. Common Mistake: Assuming One Metric Per Query

Sometimes separate queries are clearer.

Sometimes conditional aggregation is more efficient.

Do not combine unrelated metrics into one enormous SQL statement merely to reduce query count.

Optimize for:

```text
correctness
clarity
maintainability
measured performance
```

---

## 51. Production Checklist

Before shipping an aggregation query:

```text
□ What exactly does the metric mean?
□ What does one row represent before grouping?
□ What should one result row represent?
□ Which rows should be filtered by WHERE?
□ Which conditions belong in HAVING?
□ How does NULL affect the aggregate?
□ Should COUNT(*) or COUNT(column) be used?
□ Could a JOIN multiply rows?
□ Could this query double-count a metric?
□ Is the timezone defined for time-based reports?
□ Can the input dataset be bounded?
□ Does the query need pagination or LIMIT?
□ Does the index strategy support the filtering?
□ Have I checked the query plan at realistic data size?
```

---

## 52. Practical Mental Model

Think:

```text
Raw rows
   ↓
WHERE
   ↓
GROUP BY
   ↓
Aggregate
   ↓
HAVING
   ↓
ORDER BY
   ↓
LIMIT / pagination
```

Example:

```text
orders
↓
paid orders
↓
group by user
↓
count + sum
↓
users with >= 10 orders
↓
highest revenue first
↓
top 20
```

SQL:

```sql
SELECT
    user_id,
    COUNT(*) AS order_count,
    SUM(total) AS revenue
FROM orders
WHERE status = 'paid'
GROUP BY user_id
HAVING COUNT(*) >= 10
ORDER BY revenue DESC, user_id DESC
LIMIT 20;
```

This is the core aggregation workflow.

---

## 53. What to Remember

1. `COUNT(*)` counts rows.
2. `COUNT(column)` counts non-NULL values.
3. `COUNT(DISTINCT column)` counts unique non-NULL values.
4. `SUM`, `AVG`, `MIN`, and `MAX` aggregate values and generally ignore NULL inputs.
5. Without `GROUP BY`, an aggregate summarizes the matching dataset.
6. `GROUP BY` produces one result group per grouping key.
7. `WHERE` filters rows before grouping.
8. `HAVING` filters groups after aggregation.
9. Always understand the grain of rows before aggregating.
10. One-to-many joins can cause double counting.
11. `COUNT(joined_table.id)` is often useful when counting optional rows from a `LEFT JOIN`.
12. Do not use `SUM(DISTINCT ...)` as a generic double-counting fix.
13. Time-based reports need an explicit timezone definition.
14. Bound the input dataset when possible.
15. Define the business metric before writing the SQL.

---

## Next

Next material:

```text
06-insert-update-delete
```

Focus:

```text
INSERT
UPDATE
DELETE
RETURNING
safe UPDATE / DELETE
transactions
upsert basics
affected rows
production write patterns
common mistakes
```
