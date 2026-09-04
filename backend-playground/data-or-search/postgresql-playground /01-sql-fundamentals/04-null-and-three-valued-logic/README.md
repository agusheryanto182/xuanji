# 04 — NULL and Three-Valued Logic

## Goal

Understand the NULL behavior that commonly causes SQL bugs in production.

Focus:

- what `NULL` means
- `IS NULL` / `IS NOT NULL`
- `TRUE` / `FALSE` / `UNKNOWN`
- `WHERE`
- `AND` / `OR` / `NOT`
- `IN` / `NOT IN`
- `COUNT(*)` vs `COUNT(column)`
- `COALESCE`
- NULL in joins
- nullable schema design
- common mistakes

The goal is not to memorize every NULL edge case.

The goal is:

> **Understand when a missing value changes the meaning of a query.**

---

## 1. What Is NULL?

`NULL` represents an absent, unknown, or not-applicable value.

It is not the same as:

```text
0
''
FALSE
```

Example:

```text
users
--------------------------------
id | name  | phone
---+-------+-------------------
1  | Alice | 08123456789
2  | Bob   | NULL
```

For Bob, `phone` has no value.

---

## 2. NULL Is Not an Empty String

These are different:

```text
phone = NULL
phone = ''
```

`NULL` means there is no value.

An empty string means there is a text value, but it contains zero characters.

Do not treat them as interchangeable.

---

## 3. NULL Is Not Zero

For numeric data:

```text
price = 0
```

means:

```text
the price is zero
```

while:

```text
price = NULL
```

means:

```text
the price is absent/unknown
```

These represent different business states.

---

## 4. NULL Is Not FALSE

For boolean data:

```text
TRUE
FALSE
NULL
```

can represent three different states.

For example:

```text
email_verified = TRUE
→ verified

email_verified = FALSE
→ explicitly not verified

email_verified = NULL
→ status not known/not set
```

Whether a nullable boolean is a good design depends on the domain.

If there are only two valid states, `NOT NULL` is often clearer.

---

## 5. Comparing With NULL

This is a common mistake:

```sql
WHERE phone = NULL
```

Do not use it to find NULL values.

Use:

```sql
WHERE phone IS NULL;
```

And:

```sql
WHERE phone IS NOT NULL;
```

Why?

Because `NULL` does not behave like an ordinary value.

---

## 6. TRUE, FALSE, UNKNOWN

SQL commonly works with three logical results:

```text
TRUE
FALSE
UNKNOWN
```

A comparison involving NULL often produces:

```text
UNKNOWN
```

For example:

```sql
NULL = 10
```

does not produce `FALSE`.

It produces:

```text
UNKNOWN
```

This is why:

```sql
WHERE price = 10
```

does not include rows where `price` is NULL.

---

## 7. WHERE Keeps TRUE

A useful production mental model is:

```text
WHERE condition
```

keeps rows where the condition is:

```text
TRUE
```

Rows where the condition is:

```text
FALSE
UNKNOWN
```

are not returned.

This explains many NULL-related bugs.

Example:

```sql
SELECT *
FROM users
WHERE phone = '08123456789';
```

Rows with:

```text
phone = NULL
```

do not match.

---

## 8. IS NULL

Use:

```sql
WHERE phone IS NULL;
```

to find rows where the value is NULL.

Example:

```sql
SELECT id, name
FROM users
WHERE phone IS NULL;
```

This is the standard production pattern.

---

## 9. IS NOT NULL

Use:

```sql
WHERE phone IS NOT NULL;
```

to find rows with a value.

Example:

```sql
SELECT id, name
FROM users
WHERE phone IS NOT NULL;
```

---

## 10. AND With NULL

Consider:

```sql
WHERE status = 'active'
  AND phone = NULL
```

The second comparison produces:

```text
UNKNOWN
```

The overall condition cannot become TRUE.

Use:

```sql
WHERE status = 'active'
  AND phone IS NULL;
```

The explicit `IS NULL` expresses the intended condition.

---

## 11. OR With NULL

Consider:

```sql
WHERE status = 'active'
   OR phone = NULL
```

The second condition is UNKNOWN.

If the first condition is TRUE:

```text
TRUE OR UNKNOWN
→ TRUE
```

So an active user can still match.

The important point is that NULL participates in three-valued logic.

Do not mentally treat UNKNOWN as ordinary FALSE.

---

## 12. NOT With NULL

Consider:

```sql
WHERE NOT (phone = NULL)
```

The comparison is:

```text
UNKNOWN
```

and:

```text
NOT UNKNOWN
→ UNKNOWN
```

It does not become TRUE.

This is another reason why:

```sql
NOT (column = NULL)
```

is not a valid way to test non-NULL values.

Use:

```sql
column IS NOT NULL
```

---

## 13. Common Boolean Logic

A practical simplified model:

```text
TRUE AND TRUE
→ TRUE

TRUE AND FALSE
→ FALSE

TRUE AND UNKNOWN
→ UNKNOWN

TRUE OR FALSE
→ TRUE

FALSE OR UNKNOWN
→ UNKNOWN

TRUE OR UNKNOWN
→ TRUE

NOT TRUE
→ FALSE

NOT FALSE
→ TRUE

NOT UNKNOWN
→ UNKNOWN
```

You do not need to memorize every combination.

Remember:

> UNKNOWN is a third state, not another spelling of FALSE.

---

## 14. NULL and `IN`

Suppose:

```sql
WHERE status IN ('paid', 'pending')
```

This is straightforward when `status` has a value.

But:

```sql
status = NULL
```

does not match the values.

For nullable columns, remember that normal equality matching does not turn NULL into an ordinary value.

Use:

```sql
IS NULL
```

when NULL itself is part of the requirement.

---

## 15. The `NOT IN` Trap

This is one of the most important NULL-related production mistakes.

Consider:

```sql
WHERE user_id NOT IN (1, 2, NULL)
```

The presence of NULL can make the predicate evaluate to UNKNOWN for values that would otherwise appear to be valid matches.

This can result in unexpectedly returning no rows.

### Production rule

> Be very careful with `NOT IN` when NULL can appear in the compared data or subquery.

---

## 16. `NOT EXISTS` as a Safer Pattern

Suppose the requirement is:

> Find users who do not have an order.

A robust pattern is:

```sql
SELECT u.id, u.name
FROM users u
WHERE NOT EXISTS (
    SELECT 1
    FROM orders o
    WHERE o.user_id = u.id
);
```

This expresses the relationship directly:

```text
return users
for which no matching order exists
```

It avoids the classic `NOT IN` + NULL problem.

Do not assume `NOT EXISTS` is always faster. The main point here is correctness and clear semantics.

---

## 17. NULL in Aggregates

Aggregate functions generally ignore NULL values for the column being aggregated.

Example:

```text
scores
------
10
20
NULL
```

Then:

```sql
SELECT AVG(score)
FROM scores;
```

averages:

```text
10 and 20
```

The NULL does not contribute a numeric value.

This matters when interpreting reports.

---

## 18. COUNT(\*) vs COUNT(column)

This distinction is extremely important.

```sql
COUNT(*)
```

counts rows.

```sql
COUNT(phone)
```

counts non-NULL values in `phone`.

Example:

```text
id | phone
---+------
1  | 111
2  | NULL
3  | 333
```

Then:

```sql
COUNT(*)
→ 3

COUNT(phone)
→ 2
```

### Production use

For:

> How many users exist?

use:

```sql
COUNT(*)
```

For:

> How many users have a phone number?

use:

```sql
COUNT(phone)
```

---

## 19. SUM, AVG, MIN, MAX and NULL

For normal aggregate use, NULL values generally do not contribute values.

Example:

```sql
SELECT
    SUM(total),
    AVG(total),
    MIN(total),
    MAX(total)
FROM orders;
```

Rows with:

```text
total = NULL
```

are ignored by these aggregate calculations.

But if all input values are NULL, some aggregates can return NULL rather than a numeric zero.

This is why result semantics should be understood before exposing aggregate results through an API.

---

## 20. COALESCE

`COALESCE` returns the first non-NULL value.

Example:

```sql
SELECT
    COALESCE(phone, 'Not provided') AS phone
FROM users;
```

If:

```text
phone = NULL
```

the result becomes:

```text
Not provided
```

Common uses:

```text
display fallback
default calculated result
API/report formatting
```

---

## 21. COALESCE for Aggregate Results

Suppose an aggregate can return NULL:

```sql
SELECT SUM(total)
FROM orders
WHERE user_id = $1;
```

You may want:

```text
0
```

when there are no applicable values.

Use:

```sql
SELECT COALESCE(SUM(total), 0)
FROM orders
WHERE user_id = $1;
```

Now the application can receive a numeric zero instead of NULL.

Use this only when `0` is actually the correct business meaning.

---

## 22. COALESCE Is Not a Substitute for Good Schema Design

This:

```sql
COALESCE(status, 'unknown')
```

can hide the fact that `status` should perhaps never be NULL.

If a column must always have a meaningful value, prefer:

```sql
status TEXT NOT NULL
```

and use `COALESCE` where a real fallback is needed.

Production principle:

> Fix invalid or ambiguous schema states at the schema level instead of masking them everywhere in queries.

---

## 23. NULL in LEFT JOIN

Consider:

```text
users
orders
```

and:

```sql
SELECT
    u.id,
    u.name,
    o.id AS order_id
FROM users u
LEFT JOIN orders o
    ON o.user_id = u.id;
```

A user with no orders can still appear.

For that user:

```text
order_id = NULL
```

This is an important and intentional use of NULL.

It represents:

```text
no matching order
```

---

## 24. Finding Rows Without a Match

A common pattern is:

```sql
SELECT
    u.id,
    u.name
FROM users u
LEFT JOIN orders o
    ON o.user_id = u.id
WHERE o.id IS NULL;
```

Meaning:

```text
LEFT JOIN users with orders
↓
keep users where no order matched
```

This is useful for requirements such as:

```text
users without orders
products without sales
customers without payments
```

For many such cases, `NOT EXISTS` is also a clear alternative.

---

## 25. NULL and JOIN Conditions

Be careful when filtering columns from the right side of a `LEFT JOIN`.

Compare:

```sql
SELECT u.id, o.id
FROM users u
LEFT JOIN orders o
    ON o.user_id = u.id
WHERE o.status = 'paid';
```

The `WHERE` condition removes rows where `o.status` is NULL.

This can make the result behave much more like an inner join.

If the requirement is:

> Keep all users, but only match paid orders

put the condition in the join:

```sql
SELECT u.id, o.id
FROM users u
LEFT JOIN orders o
    ON o.user_id = u.id
   AND o.status = 'paid';
```

### Production rule

> With `LEFT JOIN`, think carefully about whether a condition belongs in `ON` or `WHERE`.

---

## 26. Nullable Foreign Keys

A foreign key can be nullable.

Example:

```sql
CREATE TABLE tasks (
    id BIGINT PRIMARY KEY,
    assigned_user_id BIGINT REFERENCES users(id)
);
```

Then:

```text
assigned_user_id = NULL
```

can mean:

```text
task is not assigned
```

This is a valid model if "unassigned" is a real business state.

If every task must have an owner, use:

```sql
assigned_user_id BIGINT NOT NULL REFERENCES users(id)
```

The schema should represent the actual business rule.

---

## 27. Nullable Columns Need Meaning

For every nullable column, ask:

> What does NULL mean?

Possible meanings:

```text
unknown
not provided
not applicable
not yet processed
not assigned
not deleted
```

These are not necessarily the same.

If one column needs to represent multiple meanings of "missing," the model may need improvement.

---

## 28. NULL and Soft Delete

A common soft-delete design is:

```sql
deleted_at TIMESTAMPTZ
```

with:

```text
NULL
→ not deleted

timestamp
→ deleted
```

Query active records:

```sql
SELECT id, name
FROM users
WHERE deleted_at IS NULL;
```

This is a practical production use of NULL.

The design still needs consideration for:

```text
indexes
uniqueness
retention
restoration
queries
```

---

## 29. NOT NULL Is Often Better Than Nullable

If a field is required, make that explicit.

Instead of:

```sql
name TEXT
```

use:

```sql
name TEXT NOT NULL
```

Instead of:

```sql
status TEXT
```

when every record must have a status, use:

```sql
status TEXT NOT NULL
```

This reduces ambiguous states.

### Production principle

> Nullable should be intentional, not the default.

---

## 30. NULL and API Responses

Database NULL often maps to:

```json
null
```

in JSON.

Example:

```json
{
  "id": 1,
  "phone": null
}
```

That is different from:

```json
{
  "id": 1,
  "phone": ""
}
```

The API contract should decide whether the distinction matters.

Do not automatically convert every NULL to an empty string.

---

## 31. NULL and Go

When using Go with SQL, nullable database values need to be represented correctly.

Common standard-library types include:

```go
sql.NullString
sql.NullInt64
sql.NullBool
sql.NullTime
```

Modern code can also use pointers or other explicit nullable representations depending on the domain and project style.

The important point is:

> A nullable database column needs a representation that can distinguish "no value" from the type's zero value.

For example:

```text
NULL string
```

is not necessarily the same as:

```text
""
```

---

## 32. NULL and Zero Values

This is especially important in Go.

A Go string's zero value is:

```go
""
```

A nullable SQL string can represent:

```text
NULL
```

Those states are different.

If the distinction matters to the business logic, the Go model must preserve it.

---

## 33. Common Mistake: `= NULL`

Wrong:

```sql
WHERE deleted_at = NULL;
```

Correct:

```sql
WHERE deleted_at IS NULL;
```

Wrong:

```sql
WHERE deleted_at <> NULL;
```

Correct:

```sql
WHERE deleted_at IS NOT NULL;
```

This is the first NULL rule to memorize.

---

## 34. Common Mistake: Treating UNKNOWN as FALSE

SQL has:

```text
TRUE
FALSE
UNKNOWN
```

A condition that evaluates to UNKNOWN is not selected by `WHERE`.

Do not reason about nullable expressions using only two-state boolean logic.

---

## 35. Common Mistake: `NOT IN` With NULL

Be careful with:

```sql
NOT IN
```

when NULL can enter the comparison set.

Prefer:

```sql
NOT EXISTS
```

for anti-join style requirements when it makes the intended semantics clearer.

---

## 36. Common Mistake: Using COALESCE Everywhere

This:

```sql
COALESCE(column, default)
```

is useful.

But if every query needs to compensate for a nullable field, ask whether the schema should instead use:

```sql
NOT NULL
DEFAULT ...
```

when appropriate.

Do not hide an unclear data model behind query expressions.

---

## 37. Common Mistake: Confusing NULL With "Not Found"

These are different concepts.

A query can return:

```text
a row whose column is NULL
```

or:

```text
no row at all
```

For example:

```sql
SELECT phone
FROM users
WHERE id = $1;
```

can find a user whose phone is NULL.

That is different from the user not existing.

Application code should distinguish:

```text
record not found
```

from:

```text
record found with NULL field
```

---

## 38. Common Mistake: Breaking a LEFT JOIN

This pattern:

```sql
FROM users u
LEFT JOIN orders o
  ON o.user_id = u.id
WHERE o.status = 'paid'
```

may remove users without orders.

If you need to preserve all users and only match paid orders:

```sql
FROM users u
LEFT JOIN orders o
  ON o.user_id = u.id
 AND o.status = 'paid'
```

Always check whether the condition changes the intended join semantics.

---

## 39. Production Checklist

When working with nullable data:

```text
□ What exactly does NULL mean here?
□ Is the column actually allowed to be NULL?
□ Am I using IS NULL / IS NOT NULL?
□ Could UNKNOWN affect this WHERE condition?
□ Could NOT IN encounter NULL?
□ Should this be NOT EXISTS instead?
□ Does COUNT(*) or COUNT(column) match the requirement?
□ Can an aggregate return NULL?
□ Should COALESCE be used?
□ Does a LEFT JOIN need the condition in ON or WHERE?
□ Does the API need to distinguish null from an empty/zero value?
□ Does the Go model preserve nullable state?
```

---

## 40. Practical Mental Model

Think:

```text
NULL
↓
not an ordinary value
↓
comparisons can produce UNKNOWN
↓
WHERE keeps TRUE
↓
use IS NULL / IS NOT NULL
```

For production queries:

```text
nullable data
    ↓
understand its meaning
    ↓
write explicit NULL logic
    ↓
check joins and aggregates
    ↓
preserve the meaning in the API
```

---

## 41. What to Remember

1. `NULL` is not `0`, `''`, or `FALSE`.
2. Use `IS NULL` to test for NULL.
3. Use `IS NOT NULL` to test for a value.
4. SQL can produce `TRUE`, `FALSE`, and `UNKNOWN`.
5. `WHERE` keeps rows where the condition is TRUE.
6. `NOT IN` can behave unexpectedly when NULL is involved.
7. `NOT EXISTS` is often a clear pattern for "no matching row."
8. `COUNT(*)` counts rows; `COUNT(column)` ignores NULL values.
9. `COALESCE` provides a fallback for NULL.
10. Nullable columns should have an intentional business meaning.
11. `LEFT JOIN` conditions can change meaning depending on whether they are in `ON` or `WHERE`.
12. A database NULL and an application zero value are not automatically equivalent.

---

## Next

Next material:

```text
05-aggregation-and-grouping
```

Focus:

```text
COUNT
SUM
AVG
MIN / MAX
GROUP BY
HAVING
WHERE vs HAVING
aggregate NULL behavior
conditional aggregation
aggregation with JOINs
avoiding double counting
production reporting queries
common mistakes
```

The material will stay concise and focus on aggregation patterns commonly used in backend APIs, dashboards, and reporting.
