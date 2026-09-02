# 06 — INSERT, UPDATE, DELETE

## Goal

Learn the core SQL write operations used by production backend systems.

Focus:

- `INSERT`
- `UPDATE`
- `DELETE`
- `RETURNING`
- safe write conditions
- affected rows
- transactions
- parameterized writes
- bulk inserts
- soft delete
- common write-operation mistakes

The goal is:

> **Change data safely, predictably, and in a way the application can verify.**

---

## 1. Read vs Write

SQL operations can be broadly divided into:

```text
READ
→ SELECT

WRITE
→ INSERT
→ UPDATE
→ DELETE
```

Production applications usually perform both.

Example request:

```text
POST /users
→ INSERT

PATCH /users/123
→ UPDATE

DELETE /users/123
→ DELETE

GET /users/123
→ SELECT
```

Write operations deserve extra care because an incorrect query can permanently modify many rows.

---

## 2. INSERT

Basic syntax:

```sql
INSERT INTO users (name, email)
VALUES ('Agus', 'agus@example.com');
```

This creates one row.

Prefer explicitly listing columns:

```sql
INSERT INTO users (name, email)
VALUES ($1, $2);
```

Avoid relying on the table's implicit column order.

---

## 3. Parameterized INSERT

Application code should send values separately from SQL.

Conceptually:

```sql
INSERT INTO users (name, email)
VALUES ($1, $2);
```

The application supplies:

```text
$1 → Agus
$2 → agus@example.com
```

Do not build SQL by concatenating user input.

Bad:

```text
"INSERT ... VALUES ('" + name + "')"
```

Parameterized queries help prevent SQL injection and make query handling safer.

---

## 4. INSERT With Multiple Rows

You can insert multiple rows in one statement:

```sql
INSERT INTO products (name, price)
VALUES
    ('Keyboard', 500000),
    ('Mouse', 200000),
    ('Monitor', 3000000);
```

This can be more efficient than sending many individual insert statements.

For very large data imports, PostgreSQL-specific bulk loading mechanisms may be more appropriate.

---

## 5. INSERT and Defaults

Tables often define defaults:

```sql
CREATE TABLE users (
    id BIGSERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
```

Then:

```sql
INSERT INTO users (name)
VALUES ($1);
```

The database generates `created_at`.

Production rule:

> Let the database own values that are naturally database-generated when that is the chosen schema design.

---

## 6. INSERT and Constraints

An insert can fail because of constraints.

Examples:

```text
NOT NULL
UNIQUE
PRIMARY KEY
FOREIGN KEY
CHECK
```

Example:

```sql
INSERT INTO users (name, email)
VALUES (NULL, 'a@example.com');
```

If `name` is `NOT NULL`, PostgreSQL rejects it.

This is expected behavior.

Application validation improves user experience, while database constraints protect data integrity.

---

## 7. RETURNING

PostgreSQL supports:

```sql
INSERT INTO users (name, email)
VALUES ($1, $2)
RETURNING id, name, email, created_at;
```

This allows the application to receive the inserted row without immediately issuing another `SELECT`.

Common use:

```text
INSERT
↓
RETURN generated ID / timestamps
↓
build API response
```

---

## 8. UPDATE

Basic syntax:

```sql
UPDATE users
SET name = $1
WHERE id = $2;
```

This changes matching rows.

The `WHERE` clause is critical.

---

## 9. UPDATE Without WHERE

This is dangerous:

```sql
UPDATE users
SET name = 'Agus';
```

It updates every row.

Before running an update, ask:

> Exactly which rows should change?

Production habit:

```sql
UPDATE users
SET name = $1
WHERE id = $2;
```

Always make the target explicit.

---

## 10. UPDATE Multiple Columns

You can update multiple fields:

```sql
UPDATE users
SET
    name = $1,
    phone = $2,
    updated_at = now()
WHERE id = $3;
```

Keep the update limited to fields that are actually intended to change.

Avoid accidentally overwriting unrelated fields with empty/default values.

---

## 11. Partial Updates

For an API such as:

```text
PATCH /users/123
```

the application may only want to update one field.

Example:

```sql
UPDATE users
SET name = $1
WHERE id = $2;
```

Do not automatically rewrite every column unless the API semantics require a full replacement.

---

## 12. UPDATE RETURNING

PostgreSQL also supports:

```sql
UPDATE users
SET name = $1,
    updated_at = now()
WHERE id = $2
RETURNING id, name, updated_at;
```

This is useful when the application needs the resulting database state.

---

## 13. Affected Rows

A write operation can tell the application how many rows were affected.

Example:

```sql
UPDATE users
SET name = $1
WHERE id = $2;
```

Possible result:

```text
1 row affected
```

or:

```text
0 rows affected
```

Zero affected rows can mean:

```text
record does not exist
record does not match condition
record was already excluded by another condition
```

The application should decide what that means for the API.

---

## 14. UPDATE and Business Conditions

You can make the update conditional:

```sql
UPDATE orders
SET status = 'paid'
WHERE id = $1
  AND status = 'pending';
```

Now only pending orders can transition to paid.

If:

```text
0 rows affected
```

the order may not be pending anymore.

This pattern can help enforce state transitions atomically.

---

## 15. DELETE

Basic syntax:

```sql
DELETE FROM users
WHERE id = $1;
```

Again, the `WHERE` clause defines the target.

---

## 16. DELETE Without WHERE

This:

```sql
DELETE FROM users;
```

deletes every row in the table.

Treat unbounded `DELETE` as a dangerous operation.

For production code, make the intended scope explicit.

---

## 17. DELETE RETURNING

PostgreSQL supports:

```sql
DELETE FROM users
WHERE id = $1
RETURNING id;
```

This can tell the application which row was deleted.

It can also return additional information when required.

---

## 18. Soft Delete

Many production systems do not physically delete records immediately.

Instead:

```sql
UPDATE users
SET deleted_at = now()
WHERE id = $1;
```

Normal queries then use:

```sql
WHERE deleted_at IS NULL
```

This is called soft delete.

Typical reasons:

```text
auditability
recovery
historical references
business requirements
```

---

## 19. Soft Delete Is a Data Model Decision

Soft delete is not automatically better.

It introduces responsibilities:

```text
every normal query must exclude deleted rows
unique constraints may need consideration
storage continues to grow
relationships need defined behavior
```

If using soft delete, make the convention explicit.

---

## 20. Restore

A soft-deleted record can sometimes be restored:

```sql
UPDATE users
SET deleted_at = NULL
WHERE id = $1;
```

This only makes sense if the product supports restoration.

---

## 21. Safe DELETE Pattern

Before executing:

```sql
DELETE FROM orders
WHERE user_id = $1;
```

verify the intended target.

A safer development workflow is:

```sql
SELECT id
FROM orders
WHERE user_id = $1;
```

inspect the result, then perform the delete.

For production application code, the correct `WHERE` condition should already be defined and tested.

---

## 22. Transaction

A transaction groups multiple database operations into one logical unit.

Example:

```text
create order
+
create order items
+
update inventory
```

If one required operation fails, the system may need to roll back the entire operation.

Conceptually:

```text
BEGIN
↓
operation A
↓
operation B
↓
operation C
↓
COMMIT
```

On failure:

```text
BEGIN
↓
operation A
↓
operation B fails
↓
ROLLBACK
```

---

## 23. Why Transactions Matter

Without a transaction:

```text
INSERT order        ✓
INSERT order items  ✓
UPDATE inventory    ✗
```

The database may be left in a partially completed state.

With a transaction:

```text
INSERT order        ✓
INSERT order items  ✓
UPDATE inventory    ✗
        ↓
ROLLBACK
```

The whole unit can be reverted.

---

## 24. Transaction Boundary

Do not put unrelated operations into the same transaction just because transactions exist.

A good transaction boundary represents one logical business operation.

Example:

```text
PlaceOrder
→ create order
→ create items
→ reserve inventory
```

These operations belong together if they must succeed or fail together.

---

## 25. Keep Transactions Short

Production transactions should generally be as short as practical.

Avoid:

```text
BEGIN
↓
database work
↓
HTTP request to another service
↓
user interaction
↓
long computation
↓
COMMIT
```

Holding a transaction open unnecessarily can increase:

```text
lock duration
contention
resource usage
latency
```

Prefer doing external work outside the transaction when the business design allows it.

---

## 26. Application Transaction Flow

Typical backend structure:

```text
BEGIN
↓
write A
↓
write B
↓
write C
↓
COMMIT
```

On any required error:

```text
ROLLBACK
```

The exact implementation depends on the language/framework, but the database principle is the same.

---

## 27. Parameterized UPDATE

Use parameters:

```sql
UPDATE products
SET price = $1
WHERE id = $2;
```

Do not construct:

```text
UPDATE products SET price = <user input>
```

through raw string concatenation.

Values should be bound parameters.

---

## 28. Dynamic UPDATE Fields

Values can be parameterized:

```sql
UPDATE users
SET name = $1
WHERE id = $2;
```

Identifiers generally cannot be supplied as ordinary value parameters.

For example, an application should not blindly accept:

```text
sortField = user_input
```

and place it directly into SQL.

Use a whitelist:

```text
"name" → users.name
"email" → users.email
"created_at" → users.created_at
```

Then construct SQL from trusted identifiers.

---

## 29. UPDATE With Timestamp

A common production pattern:

```sql
UPDATE users
SET
    name = $1,
    updated_at = now()
WHERE id = $2;
```

The database generates the update timestamp.

Whether the application or database owns timestamps should be consistent across the system.

---

## 30. UPDATE With Version

Optimistic concurrency can use a version column:

```sql
UPDATE documents
SET
    content = $1,
    version = version + 1
WHERE id = $2
  AND version = $3;
```

If zero rows are affected, another update may have changed the document first.

This is a useful production pattern for avoiding lost updates.

Detailed locking and isolation belong to later PostgreSQL materials.

---

## 31. INSERT Then SELECT?

Sometimes code does:

```text
INSERT
↓
SELECT inserted row
```

PostgreSQL `RETURNING` can often simplify this:

```sql
INSERT INTO users (name)
VALUES ($1)
RETURNING id, name, created_at;
```

This avoids an unnecessary second round trip when all required data can be returned directly.

---

## 32. DELETE Then SELECT?

Similarly:

```sql
DELETE FROM users
WHERE id = $1
RETURNING id;
```

can tell the application whether the row existed and was deleted.

This is often cleaner than:

```text
SELECT first
↓
DELETE second
```

when the only goal is to know whether the delete succeeded.

---

## 33. Write Operations Should Be Atomic

A useful production question is:

> Can another request observe an invalid intermediate state?

If several writes must change together, use a transaction.

If one SQL statement is enough, a single statement is already atomic at the statement level.

Prefer a single well-defined statement when it fully expresses the required operation.

---

## 34. DELETE and Foreign Keys

Deleting a parent can fail if child rows reference it.

Example:

```text
users
  ↓
orders
```

A foreign key can prevent:

```sql
DELETE FROM users
WHERE id = $1;
```

when orders still reference that user.

The schema may instead define behavior such as:

```text
RESTRICT
CASCADE
SET NULL
```

The correct choice is a business/data-model decision.

Do not add `CASCADE` simply to make deletes easier.

---

## 35. UPDATE and Foreign Keys

Changing a foreign key can also be constrained.

Example:

```sql
UPDATE orders
SET user_id = $1
WHERE id = $2;
```

The new user must satisfy the foreign key constraint unless the schema explicitly allows another behavior.

Database constraints protect referential integrity.

---

## 36. Write Validation: Application + Database

A production system often validates at two levels.

Application:

```text
required fields
format
business rules
API error messages
```

Database:

```text
NOT NULL
UNIQUE
CHECK
FOREIGN KEY
```

The application provides a good user experience.

The database provides the final integrity boundary.

---

## 37. Idempotency and INSERT

Repeated requests can accidentally create duplicate records.

Example:

```text
client retries payment request
↓
same INSERT executes twice
↓
two payments
```

Depending on the business operation, use mechanisms such as:

```text
unique constraint
idempotency key
UPSERT
transaction
```

The correct design depends on the operation.

---

## 38. Unique Constraint as Protection

Suppose email must be unique:

```sql
CREATE UNIQUE INDEX users_email_unique
ON users(email);
```

Then:

```sql
INSERT INTO users (email)
VALUES ($1);
```

cannot silently create duplicates.

The application should handle the resulting constraint violation appropriately.

---

## 39. Do Not Rely Only on Pre-Checks

This pattern is unsafe under concurrency:

```text
SELECT whether email exists
↓
if not exists
↓
INSERT
```

Two requests can both observe:

```text
does not exist
```

and then both attempt the insert.

A database `UNIQUE` constraint is the real protection.

This is an important production principle:

> Validation before a write does not replace a database constraint.

---

## 40. Upsert Preview

PostgreSQL supports:

```sql
INSERT INTO users (email, name)
VALUES ($1, $2)
ON CONFLICT (email)
DO UPDATE
SET name = EXCLUDED.name;
```

This means:

```text
try INSERT
↓
conflict on email
↓
UPDATE instead
```

Upsert deserves its own material, but recognize the pattern here.

---

## 41. Affected Rows Are Part of the Contract

For:

```sql
UPDATE users
SET name = $1
WHERE id = $2;
```

the application should distinguish:

```text
1 row affected
→ target matched

0 rows affected
→ target did not match
```

Depending on the query, zero can represent:

```text
not found
invalid state
already processed
concurrent change
```

Do not automatically map every zero-row result to the same API behavior.

---

## 42. Example: State Transition

Suppose an order can move:

```text
pending → paid
```

Use:

```sql
UPDATE orders
SET
    status = 'paid',
    paid_at = now()
WHERE id = $1
  AND status = 'pending'
RETURNING id, status, paid_at;
```

Possible outcomes:

```text
row returned
→ transition succeeded

no row returned
→ order did not satisfy the expected state
```

This combines:

```text
conditional update
+
atomic state check
+
RETURNING
```

---

## 43. Bulk Writes

For a small number of rows:

```sql
INSERT INTO products (name, price)
VALUES
    ($1, $2),
    ($3, $4),
    ($5, $6);
```

can reduce round trips.

For very large workloads, use appropriate PostgreSQL bulk-loading mechanisms instead of generating enormous SQL statements.

The production goal is not:

> One giant SQL statement at all costs.

The goal is:

> Efficient and manageable database writes.

---

## 44. Write Size Matters

Avoid unnecessarily large writes.

For example, if only one field changed:

```sql
UPDATE users
SET name = $1
WHERE id = $2;
```

is clearer than rewriting every column.

Smaller write scope can reduce accidental changes and simplify concurrency reasoning.

---

## 45. Common Mistake: Missing WHERE

Dangerous:

```sql
UPDATE users
SET active = false;
```

Dangerous:

```sql
DELETE FROM users;
```

Production habit:

```text
Before UPDATE/DELETE:
→ identify exact target
→ verify WHERE
→ understand expected affected rows
```

---

## 46. Common Mistake: SQL String Concatenation

Bad:

```text
"UPDATE users SET name = '" + name + "' WHERE id = " + id
```

Use parameters instead:

```sql
UPDATE users
SET name = $1
WHERE id = $2;
```

This protects values from being interpreted as SQL syntax.

---

## 47. Common Mistake: SELECT Before Every Write

Sometimes code does:

```text
SELECT row
↓
UPDATE row
```

only to verify that the row exists.

Often the update itself can provide that information:

```sql
UPDATE ...
RETURNING ...
```

or through affected-row count.

Avoid unnecessary database round trips when a single atomic statement expresses the requirement.

---

## 48. Common Mistake: Long Transactions

Avoid holding a transaction while performing:

```text
external HTTP calls
long calculations
file uploads
slow user interactions
```

when those operations do not need to happen inside the transaction.

Long transactions can increase contention and lock duration.

---

## 49. Common Mistake: Treating Application Validation as Integrity

This is not sufficient:

```text
SELECT email
↓
if available
↓
INSERT
```

Concurrent requests can race.

Use:

```text
UNIQUE constraint
```

as the database guarantee.

---

## 50. Common Mistake: Blind DELETE

Physical deletion is irreversible unless another recovery mechanism exists.

Before using `DELETE`, determine whether the product requires:

```text
audit history
restore
retention
soft delete
legal/compliance retention
```

Deletion strategy is a business decision, not merely a SQL syntax choice.

---

## 51. Production Write Checklist

Before shipping a write query:

```text
□ Are all values parameterized?
□ Is the WHERE clause correct?
□ Could UPDATE/DELETE affect more rows than intended?
□ Is the expected affected-row count known?
□ Should RETURNING be used?
□ Are database constraints protecting integrity?
□ Does this operation need a transaction?
□ Is the transaction boundary correct?
□ Is the transaction short enough?
□ Could concurrent requests conflict?
□ Is idempotency required?
□ Could a foreign key prevent the operation?
□ Should this be soft delete instead of DELETE?
□ Is a pre-check being incorrectly used instead of a constraint?
□ Is the write larger than necessary?
```

---

## 52. Practical Mental Model

For writes:

```text
Define target
   ↓
Validate input
   ↓
Execute parameterized SQL
   ↓
Database constraints protect integrity
   ↓
Check result / RETURNING
   ↓
Commit if transactional
```

For multiple related writes:

```text
BEGIN
   ↓
write A
   ↓
write B
   ↓
write C
   ↓
COMMIT
```

If a required operation fails:

```text
ROLLBACK
```

---

## 53. Core Production Patterns

### Create

```sql
INSERT INTO users (name, email)
VALUES ($1, $2)
RETURNING id, name, email, created_at;
```

### Update

```sql
UPDATE users
SET
    name = $1,
    updated_at = now()
WHERE id = $2
RETURNING id, name, updated_at;
```

### Delete

```sql
DELETE FROM users
WHERE id = $1
RETURNING id;
```

### Soft delete

```sql
UPDATE users
SET deleted_at = now()
WHERE id = $1
RETURNING id, deleted_at;
```

### Conditional state transition

```sql
UPDATE orders
SET status = 'paid'
WHERE id = $1
  AND status = 'pending'
RETURNING id, status;
```

These patterns cover a large amount of everyday backend database work.

---

## 54. What to Remember

1. `INSERT` creates rows.
2. `UPDATE` changes matching rows.
3. `DELETE` removes matching rows.
4. Always make the target of `UPDATE` and `DELETE` explicit.
5. Parameterize values; do not concatenate user input into SQL.
6. PostgreSQL `RETURNING` can return affected rows directly.
7. Zero affected rows should be interpreted according to the business operation.
8. Database constraints are the final integrity boundary.
9. A transaction groups related writes into one logical unit.
10. Keep transaction boundaries focused and transactions short.
11. `UNIQUE` constraints protect against concurrent duplicate creation.
12. A pre-check does not replace a database constraint.
13. Soft delete is a schema/product decision, not a universal rule.
14. Conditional updates can enforce state transitions atomically.
15. Always reason about concurrency when a write depends on current database state.

---

## Next

```text
07-sql-query-thinking
```

The next material focuses on combining the fundamentals into a production-oriented way of thinking about SQL queries.
