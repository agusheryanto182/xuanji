# UNIQUE and CHECK Constraints

Production-first PostgreSQL material for enforcing uniqueness and business invariants at the database level.

---

## 1. What This Material Covers

This chapter focuses on:

- `UNIQUE`
- composite uniqueness
- `CHECK`
- primary key vs `UNIQUE`
- unique indexes
- nullable unique columns
- case sensitivity
- business invariants
- constraints vs application validation
- constraints during concurrent writes
- migration considerations
- common production mistakes

Core idea:

> **Use `UNIQUE` to prevent duplicate values or relationships. Use `CHECK` to prevent invalid values or states.**

---

# 2. Why Constraints Matter

Suppose an application checks whether an email exists:

```sql
SELECT id
FROM users
WHERE email = $1;
```

Then inserts:

```sql
INSERT INTO users (email)
VALUES ($1);
```

Two requests can execute concurrently:

```text
Request A                 Request B

check email              check email
     ↓                         ↓
not found                 not found
     ↓                         ↓
insert                    insert
```

Without a database constraint, both may succeed.

A database `UNIQUE` constraint provides the final protection.

---

# 3. UNIQUE

A `UNIQUE` constraint prevents duplicate values.

Example:

```sql
CREATE TABLE users (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    email TEXT NOT NULL UNIQUE
);
```

Now:

```text
alice@example.com
bob@example.com
```

are valid.

A second:

```text
alice@example.com
```

is rejected.

---

# 4. Primary Key vs UNIQUE

A primary key identifies a row.

```sql
id BIGINT PRIMARY KEY
```

A unique constraint prevents duplicate values.

```sql
email TEXT UNIQUE
```

A table normally has one primary key, but it can have multiple unique constraints.

Example:

```sql
CREATE TABLE users (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    email TEXT NOT NULL UNIQUE,
    username TEXT NOT NULL UNIQUE
);
```

Here:

```text
id
  → row identity

email
  → unique business attribute

username
  → unique business attribute
```

---

# 5. Named UNIQUE Constraint

For larger schemas, explicit names can improve debugging.

```sql
CREATE TABLE users (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    email TEXT NOT NULL,

    CONSTRAINT uq_users_email
        UNIQUE (email)
);
```

If a write violates the rule, the constraint name helps identify which invariant failed.

---

# 6. Composite UNIQUE

Sometimes uniqueness depends on multiple columns.

Example:

```sql
CREATE TABLE memberships (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    user_id BIGINT NOT NULL,
    organization_id BIGINT NOT NULL,

    CONSTRAINT uq_memberships_user_org
        UNIQUE (user_id, organization_id)
);
```

This prevents:

```text
user_id | organization_id
--------+----------------
1       | 10
1       | 10
```

but allows:

```text
1 | 10
1 | 20
2 | 10
```

The combination must be unique.

---

# 7. Why Composite UNIQUE Matters

A many-to-many relationship often needs this.

Example:

```text
users
  |
  +---- memberships ----+
                        |
organizations ----------+
```

The same user should not be attached to the same organization twice.

Use:

```sql
UNIQUE (user_id, organization_id)
```

or make those columns the composite primary key.

---

# 8. UNIQUE and Relationships

`UNIQUE` can enforce cardinality.

Normal foreign key:

```sql
user_id BIGINT NOT NULL
    REFERENCES users(id)
```

allows:

```text
one user → many rows
```

Add:

```sql
UNIQUE (user_id)
```

and the relationship becomes:

```text
one user → at most one row
```

This is a common way to enforce one-to-one relationships.

---

# 9. CHECK

A `CHECK` constraint ensures a condition is satisfied.

Example:

```sql
CREATE TABLE products (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    price NUMERIC(12, 2) NOT NULL,

    CHECK (price >= 0)
);
```

This rejects:

```sql
price = -10
```

The database protects the invariant:

> Product price cannot be negative.

---

# 10. CHECK for Numeric Ranges

Example:

```sql
quantity INTEGER NOT NULL
    CHECK (quantity >= 0)
```

For a bounded value:

```sql
rating INTEGER NOT NULL
    CHECK (rating BETWEEN 1 AND 5)
```

This is useful for values whose valid range is part of the data model.

---

# 11. CHECK for Status Values

A status field can use:

```sql
status TEXT NOT NULL DEFAULT 'pending'
    CHECK (
        status IN (
            'pending',
            'processing',
            'completed',
            'cancelled'
        )
    )
```

Now arbitrary values such as:

```text
banana
foobar
finishedddd
```

cannot enter the database.

---

# 12. CHECK for Text

Suppose a name must not be empty.

```sql
name TEXT NOT NULL
    CHECK (length(trim(name)) > 0)
```

`NOT NULL` protects against:

```text
NULL
```

`CHECK` protects against:

```text
''
'   '
```

Different constraints can work together.

---

# 13. CHECK and NULL

A subtle but important rule:

A `CHECK` constraint passes when its expression evaluates to `TRUE` or `NULL`; it rejects `FALSE`.

Example:

```sql
age INTEGER CHECK (age >= 18)
```

If `age` is nullable, `NULL` does not violate this check.

If age is required:

```sql
age INTEGER NOT NULL
    CHECK (age >= 18)
```

Use `NOT NULL` when absence itself is invalid.

---

# 14. UNIQUE and NULL

PostgreSQL's normal `UNIQUE` behavior allows multiple `NULL` values because `NULL` values are not considered equal in the usual uniqueness comparison.

Example:

```sql
CREATE TABLE users (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    referral_code TEXT UNIQUE
);
```

Multiple rows can have:

```text
referral_code = NULL
```

while non-null referral codes must be unique.

This can be exactly what you want for optional unique values.

---

# 15. Required Unique Values

If every row must have a unique value:

```sql
email TEXT NOT NULL UNIQUE
```

This expresses two rules:

```text
NOT NULL
  → every user has an email

UNIQUE
  → no two users share the same email
```

Do not assume `UNIQUE` alone means the column is required.

---

# 16. Optional Unique Values

Sometimes a value is optional but, when present, must be unique.

Example:

```sql
username TEXT UNIQUE
```

This allows:

```text
NULL
alice
bob
```

but prevents two non-null `alice` values.

This is useful when the business model genuinely allows the value to be absent.

---

# 17. Case Sensitivity

For text, ordinary uniqueness is generally case-sensitive.

For example:

```text
Alice@example.com
alice@example.com
```

may be treated as different values by a normal `TEXT UNIQUE` constraint.

If the business rule says email addresses should be unique case-insensitively, you need to model that intentionally.

One common PostgreSQL approach is a unique index on a normalized expression:

```sql
CREATE UNIQUE INDEX uq_users_email_lower
ON users (lower(email));
```

Then:

```text
Alice@example.com
alice@example.com
```

conflict.

The important lesson:

> **Database uniqueness should match the business definition of equality.**

---

# 18. Normalize Before Comparing

The same idea applies to values such as usernames or codes.

If the business treats these as case-insensitive:

```text
Agus
agus
AGUS
```

should represent the same identity.

Then enforce uniqueness on the normalized representation.

For example:

```sql
CREATE UNIQUE INDEX uq_users_username_lower
ON users (lower(username));
```

Do not rely only on application-side normalization if duplicate prevention is critical.

---

# 19. UNIQUE Is Concurrency-Safe

A major reason to use database uniqueness is concurrency.

Suppose:

```text
Request A → create username "agus"
Request B → create username "agus"
```

Both arrive at nearly the same time.

Application-only checks can race.

A database unique constraint serializes the conflict at the storage layer.

One request succeeds.

The conflicting request receives a uniqueness violation.

This is much safer than:

```text
check first
then insert
```

as the only protection.

---

# 20. Handling UNIQUE Violations

In backend code, a uniqueness error is often an expected business outcome.

For example:

```text
POST /users
email already exists
```

The API can return an appropriate client error such as:

```http
409 Conflict
```

The exact response contract depends on the application.

The important point:

> A constraint violation should be handled deliberately, not treated as an unexpected database disaster.

---

# 21. UNIQUE and INSERT ... ON CONFLICT

Unique constraints work naturally with PostgreSQL upserts.

Example:

```sql
CREATE TABLE users (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    email TEXT NOT NULL UNIQUE,
    name TEXT NOT NULL
);
```

Then:

```sql
INSERT INTO users (email, name)
VALUES ($1, $2)
ON CONFLICT (email)
DO UPDATE
SET name = EXCLUDED.name
RETURNING id;
```

The unique constraint defines what counts as a conflict.

---

# 22. DO NOTHING

If duplicate insertion should simply be ignored:

```sql
INSERT INTO user_roles (user_id, role_id)
VALUES ($1, $2)
ON CONFLICT (user_id, role_id)
DO NOTHING;
```

This is useful for idempotent relationship creation.

Example:

```text
assign user to role
```

If the relationship already exists, no duplicate row is created.

---

# 23. CHECK Is Not a Workflow Engine

A `CHECK` constraint is good for:

```text
price >= 0
quantity >= 0
status in (...)
```

It is not usually the right tool for complex workflow transitions.

For example:

```text
pending → processing → completed
```

is a business process.

The database can restrict valid status values, but application logic or a carefully designed database mechanism should control transitions.

---

# 24. CHECK for Cross-Column Rules

`CHECK` becomes especially useful when a rule involves multiple columns.

Example:

```sql
CREATE TABLE discounts (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    start_at TIMESTAMPTZ NOT NULL,
    end_at TIMESTAMPTZ NOT NULL,

    CHECK (end_at > start_at)
);
```

This prevents an impossible interval.

Another example:

```sql
CREATE TABLE payments (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    amount NUMERIC(12, 2) NOT NULL,
    refunded_amount NUMERIC(12, 2) NOT NULL DEFAULT 0,

    CHECK (amount >= 0),
    CHECK (refunded_amount >= 0),
    CHECK (refunded_amount <= amount)
);
```

The database protects the relationship between columns.

---

# 25. CHECK vs Application Validation

Suppose an API validates:

```text
quantity >= 0
```

That is useful.

But data may also come from:

- background jobs
- scripts
- admin tools
- imports
- migrations
- other services

If the rule is a fundamental database invariant, put it in the database too.

Example:

```sql
CHECK (quantity >= 0)
```

Application validation improves feedback.

Database constraints protect stored data.

---

# 26. Constraints Are Not Business Logic Everywhere

Do not put every business rule into SQL.

Good candidates:

```text
value cannot be negative
email must be unique
status must be one of known values
end_at must be after start_at
relationship must be unique
```

More complex candidates may belong in application logic:

```text
user may cancel only within 30 minutes
order may transition only if payment is confirmed
promotion eligibility depends on many external rules
```

Use constraints for stable invariants that the database can express clearly.

---

# 27. UNIQUE Index vs UNIQUE Constraint

PostgreSQL implements uniqueness using an index internally.

You can write:

```sql
UNIQUE (email)
```

as a table constraint.

Or create a unique index:

```sql
CREATE UNIQUE INDEX uq_users_email_lower
ON users (lower(email));
```

The constraint is useful when expressing a straightforward schema rule.

The unique index is especially useful when uniqueness depends on:

- expressions
- partial conditions
- specific indexing behavior

---

# 28. Partial Unique Index

Suppose only active records need a unique value.

Example:

```sql
CREATE UNIQUE INDEX uq_active_users_email
ON users (email)
WHERE deleted_at IS NULL;
```

This allows historical soft-deleted rows to retain the same email while ensuring active users cannot duplicate it.

Conceptually:

```text
active rows
    ↓
must be unique

deleted rows
    ↓
excluded from uniqueness
```

This is a powerful PostgreSQL production pattern.

---

# 29. Soft Delete and UNIQUE

Consider:

```sql
email TEXT NOT NULL UNIQUE,
deleted_at TIMESTAMPTZ
```

A soft-deleted user still occupies the unique email value.

If you want the email reusable after deletion, a partial unique index may be more appropriate:

```sql
CREATE UNIQUE INDEX uq_users_active_email
ON users (email)
WHERE deleted_at IS NULL;
```

The correct design depends on whether historical deleted records should continue reserving the value.

---

# 30. Multiple Unique Rules

A table can have multiple uniqueness rules.

Example:

```sql
CREATE TABLE users (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    email TEXT NOT NULL,
    username TEXT NOT NULL,

    CONSTRAINT uq_users_email
        UNIQUE (email),

    CONSTRAINT uq_users_username
        UNIQUE (username)
);
```

Each constraint protects a different invariant.

---

# 31. Composite UNIQUE and Query Design

Suppose:

```sql
UNIQUE (organization_id, username)
```

This means:

```text
username must be unique inside an organization
```

So:

```text
org 1 | agus
org 2 | agus
```

can both exist.

But:

```text
org 1 | agus
org 1 | agus
```

cannot.

This is a common multi-tenant pattern.

---

# 32. Multi-Tenant Uniqueness

For SaaS systems, uniqueness is often scoped.

Example:

```sql
CONSTRAINT uq_org_users_username
UNIQUE (organization_id, username)
```

The business rule becomes:

> A username is unique within an organization, not globally.

Always ask:

> Unique where?

Possible scopes:

```text
global
organization
tenant
project
parent record
```

The constraint should represent the actual scope.

---

# 33. CHECK and Data Types

Choose the strongest simple type first.

For example, instead of:

```sql
age TEXT
```

prefer:

```sql
age INTEGER
```

Then:

```sql
CHECK (age >= 0)
```

The database handles the basic type and range rules.

Similarly:

```sql
quantity INTEGER NOT NULL CHECK (quantity >= 0)
```

is better than storing numbers as text and validating them everywhere.

---

# 34. Constraint Naming

Consistent names make production debugging easier.

A practical convention:

```text
pk_<table>
fk_<table>_<column>
uq_<table>_<column>
ck_<table>_<rule>
```

Examples:

```text
pk_users
fk_orders_customer
uq_users_email
ck_products_price_positive
```

Exact naming conventions can vary.

Consistency matters more than the exact prefix.

---

# 35. Inspecting Constraints

PostgreSQL can show table definitions and constraints through system catalogs or `psql` commands.

In `psql`, a useful command is:

```text
\d users
```

This helps verify:

```text
primary keys
foreign keys
unique constraints
check constraints
indexes
```

When debugging a schema, inspect the actual database rather than relying only on migration files.

---

# 36. Migration: Adding UNIQUE

Before adding:

```sql
ALTER TABLE users
ADD CONSTRAINT uq_users_email UNIQUE (email);
```

check for duplicates:

```sql
SELECT email, COUNT(*)
FROM users
GROUP BY email
HAVING COUNT(*) > 1;
```

If duplicates exist, the migration can fail.

Production migration process:

```text
find duplicates
      ↓
decide correct records
      ↓
clean/backfill
      ↓
verify
      ↓
add constraint
```

---

# 37. Migration: Adding CHECK

Before adding:

```sql
ALTER TABLE products
ADD CONSTRAINT ck_products_price_nonnegative
CHECK (price >= 0);
```

find invalid rows:

```sql
SELECT id, price
FROM products
WHERE price < 0;
```

Repair existing data first.

A new constraint applies to existing rows too.

---

# 38. Large-Table Considerations

Adding constraints to large production tables can require planning.

Consider:

```text
table size
existing invalid data
lock behavior
migration duration
application compatibility
deployment strategy
```

Do not assume every schema change is instant just because the SQL statement is short.

For large systems, migration strategy is part of production engineering.

---

# 39. Constraint Validation Strategy

For large tables, PostgreSQL supports patterns such as adding a constraint and validating it separately for certain constraint types.

For example, a `CHECK` constraint can be added as:

```sql
ALTER TABLE products
ADD CONSTRAINT ck_products_price_nonnegative
CHECK (price >= 0) NOT VALID;
```

Then later:

```sql
ALTER TABLE products
VALIDATE CONSTRAINT ck_products_price_nonnegative;
```

This is an advanced migration technique useful when minimizing deployment impact matters.

Do not use it blindly; understand the operational tradeoffs first.

---

# 40. Constraint vs Index: Mental Separation

Think:

```text
UNIQUE constraint
    ↓
correctness rule

unique index
    ↓
index structure that can also enforce uniqueness
```

In practice PostgreSQL uses indexes to enforce unique constraints.

But conceptually keep the purposes separate:

```text
constraint
  → "this must be true"

index
  → "this should be efficiently searchable"
```

---

# 41. Production Example: Users

```sql
CREATE TABLE users (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,

    email TEXT NOT NULL,

    username TEXT NOT NULL,

    is_active BOOLEAN NOT NULL DEFAULT TRUE,

    CONSTRAINT uq_users_email
        UNIQUE (email),

    CONSTRAINT uq_users_username
        UNIQUE (username),

    CONSTRAINT ck_users_username_nonempty
        CHECK (length(trim(username)) > 0)
);
```

This protects:

```text
id
  → identity

email
  → required + unique

username
  → required + unique + non-empty

is_active
  → known default
```

---

# 42. Production Example: Products

```sql
CREATE TABLE products (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,

    sku TEXT NOT NULL,

    price NUMERIC(12, 2) NOT NULL,

    stock INTEGER NOT NULL DEFAULT 0,

    CONSTRAINT uq_products_sku
        UNIQUE (sku),

    CONSTRAINT ck_products_price
        CHECK (price >= 0),

    CONSTRAINT ck_products_stock
        CHECK (stock >= 0)
);
```

The database prevents:

```text
duplicate SKU
negative price
negative stock
```

These are stable data invariants.

---

# 43. Production Example: Organization Users

```sql
CREATE TABLE organization_users (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,

    organization_id BIGINT NOT NULL,
    username TEXT NOT NULL,

    CONSTRAINT uq_org_users_username
        UNIQUE (organization_id, username)
);
```

The same username can exist in different organizations:

```text
organization 1 → agus
organization 2 → agus
```

but not twice in the same organization.

---

# 44. Production Example: Time Interval

```sql
CREATE TABLE campaigns (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,

    start_at TIMESTAMPTZ NOT NULL,
    end_at TIMESTAMPTZ NOT NULL,

    CONSTRAINT ck_campaigns_valid_range
        CHECK (end_at > start_at)
);
```

This prevents invalid intervals.

The database does not need to understand the whole campaign business process to protect this simple invariant.

---

# 45. Production Review

For every important column, ask:

### Uniqueness

- Must the value be unique?
- Is uniqueness global or scoped?
- Can it be `NULL`?
- Is comparison case-sensitive?

### Validity

- What values are invalid?
- Can a simple `CHECK` express the rule?
- Does the data type already provide part of the protection?

### Concurrency

- Could two requests create the same value simultaneously?
- Is a database constraint required?

### Migration

- Does existing data already satisfy the rule?
- How will duplicates or invalid rows be repaired?
- Is the table large enough to require a staged rollout?

---

# 46. Common Mistake: Check-Then-Insert

Avoid relying on:

```text
SELECT ...
if not exists:
    INSERT ...
```

for uniqueness.

It can race under concurrency.

Prefer:

```sql
UNIQUE (...)
```

and handle the resulting conflict.

This is one of the most important production uses of database constraints.

---

# 47. Common Mistake: UNIQUE Without Understanding NULL

This:

```sql
email TEXT UNIQUE
```

does not necessarily mean:

```text
exactly one row per possible value, including NULL
```

If the field is required:

```sql
email TEXT NOT NULL UNIQUE
```

If it is optional:

```sql
email TEXT UNIQUE
```

may be correct.

Understand what `NULL` means before designing uniqueness.

---

# 48. Common Mistake: Wrong Uniqueness Scope

Bad:

```sql
UNIQUE (username)
```

when the actual rule is:

```text
username unique per organization
```

Correct:

```sql
UNIQUE (organization_id, username)
```

Always model the business scope.

---

# 49. Common Mistake: Overusing CHECK

Do not create complicated database expressions for every application rule.

Use `CHECK` for stable, local invariants that are easy to understand.

Good:

```sql
CHECK (price >= 0)
```

Good:

```sql
CHECK (end_at > start_at)
```

Be cautious with rules that require:

- external services
- other tables
- complex workflows
- changing business policies

---

# 50. Common Mistake: Relying Only on Application Validation

Application validation can be bypassed by another writer.

If the rule is fundamental:

```text
email unique
price non-negative
quantity non-negative
```

enforce it at the database level too.

---

# 51. Common Mistake: Forgetting Existing Data

A migration can fail because old data violates the new rule.

Before adding:

```sql
UNIQUE
CHECK
NOT NULL
FOREIGN KEY
```

inspect existing data.

Production schema changes are also data-quality changes.

---

# 52. Common Mistake: Treating Constraints as Optional Documentation

A comment saying:

```text
email should be unique
```

does not enforce uniqueness.

A database constraint does:

```sql
UNIQUE (email)
```

When a rule matters for correctness, encode it where possible.

---

# 53. Practical Mental Model

Think of constraints as database guardrails:

```text
PRIMARY KEY
    ↓
every row has identity

UNIQUE
    ↓
duplicates are forbidden

CHECK
    ↓
invalid states are forbidden

NOT NULL
    ↓
required values cannot be absent

FOREIGN KEY
    ↓
references must point to valid parents
```

Together:

```text
application
     ↓
database
     ↓
constraints
     ↓
valid stored state
```

---

# 54. Production Checklist

Before shipping a schema:

- Important unique business fields have `UNIQUE`.
- Required unique fields also use `NOT NULL`.
- Scoped uniqueness uses composite `UNIQUE` where appropriate.
- Case-insensitive uniqueness is modeled intentionally.
- Optional unique fields intentionally allow `NULL`.
- Stable numeric/range invariants use `CHECK`.
- Status domains use `CHECK` when appropriate.
- Cross-column invariants use `CHECK` when simple and stable.
- Application validation is not the only protection for critical invariants.
- Concurrent writes are protected by database constraints.
- `ON CONFLICT` is used where idempotent behavior is appropriate.
- Existing data is checked before adding constraints.
- Large-table migrations are planned operationally.
- Constraint names follow a consistent convention.

---

# 55. Production Takeaways

1. `UNIQUE` protects against duplicates.
2. `CHECK` protects simple data invariants.
3. `NOT NULL` and `UNIQUE` solve different problems.
4. Composite uniqueness is essential for scoped relationships.
5. Unique constraints protect against concurrent duplicate writes.
6. Application checks are useful but should not replace critical database constraints.
7. `NULL` behavior must be understood when designing uniqueness.
8. Case-insensitive uniqueness must be modeled intentionally.
9. Partial unique indexes are powerful for soft-delete and conditional uniqueness patterns.
10. `CHECK` is excellent for stable local invariants, not complex workflows.
11. Existing data must satisfy a new constraint before it can be enforced safely.
12. Constraints are part of production correctness, not just schema decoration.

Core mental model:

> **`UNIQUE` answers "can this value/relationship repeat?"  
> `CHECK` answers "is this state valid?"**
