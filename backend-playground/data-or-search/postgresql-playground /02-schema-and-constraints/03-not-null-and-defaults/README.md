# NOT NULL and Defaults

Production-first PostgreSQL material for designing required fields, optional fields, and database defaults.

---

## 1. What This Material Covers

This chapter focuses on:

- `NOT NULL`
- `DEFAULT`
- required vs optional data
- database-generated values
- application validation vs database constraints
- timestamps
- status defaults
- boolean defaults
- ID defaults
- `NULL` vs default values
- insert behavior
- update behavior
- migration considerations
- common production mistakes

Core idea:

> **`NOT NULL` defines whether a value is required. `DEFAULT` defines what PostgreSQL should use when a value is omitted.**

They solve different problems.

---

# 2. NOT NULL

`NOT NULL` means a column cannot contain `NULL`.

Example:

```sql
CREATE TABLE users (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    name TEXT NOT NULL
);
```

This is valid:

```sql
INSERT INTO users (name)
VALUES ('Alice');
```

This is rejected:

```sql
INSERT INTO users (name)
VALUES (NULL);
```

The database protects the invariant:

> Every user must have a name.

---

# 3. Why NOT NULL Matters

Without `NOT NULL`, the database allows missing values.

Example:

```sql
CREATE TABLE users (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    name TEXT
);
```

Now all of these are possible:

```text
Alice
Bob
NULL
```

If your application assumes every user has a name, the schema does not enforce that assumption.

This creates unnecessary complexity in:

- queries
- API responses
- application code
- reporting
- data migrations

Use `NOT NULL` when the business model requires a value.

---

# 4. Required vs Optional

Ask:

> Can this record logically exist without this value?

If no:

```sql
name TEXT NOT NULL
```

If yes:

```sql
middle_name TEXT
```

Example:

```sql
CREATE TABLE users (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    name TEXT NOT NULL,
    middle_name TEXT
);
```

The distinction should come from the domain.

Do not make everything nullable just because it is easier during development.

---

# 5. NOT NULL Is Not the Same as Validation

Suppose an API requires:

```json
{
  "name": ""
}
```

`NOT NULL` does not reject an empty string.

This:

```sql
name TEXT NOT NULL
```

rejects:

```sql
NULL
```

but allows:

```text
''
```

If empty strings are invalid, application validation or a `CHECK` constraint may also be required.

Example:

```sql
name TEXT NOT NULL
CHECK (length(trim(name)) > 0)
```

Use the appropriate layer for the rule.

---

# 6. DEFAULT

A `DEFAULT` specifies a value PostgreSQL should use when an insert does not provide a value.

Example:

```sql
CREATE TABLE users (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    name TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'active'
);
```

Insert:

```sql
INSERT INTO users (name)
VALUES ('Alice');
```

The database creates:

```text
name   = Alice
status = active
```

---

# 7. DEFAULT Is Used When the Column Is Omitted

This distinction is important.

Given:

```sql
status TEXT DEFAULT 'active'
```

This uses the default:

```sql
INSERT INTO users (name)
VALUES ('Alice');
```

But this explicitly supplies `NULL`:

```sql
INSERT INTO users (name, status)
VALUES ('Alice', NULL);
```

The default is not automatically used.

If the column is also `NOT NULL`, the second statement fails.

---

# 8. DEFAULT + NOT NULL

A common production pattern is:

```sql
status TEXT NOT NULL DEFAULT 'active'
```

This means:

1. the value is required in stored data
2. callers may omit it during insert
3. PostgreSQL fills in the default

Example:

```sql
CREATE TABLE orders (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    status TEXT NOT NULL DEFAULT 'pending'
);
```

Application:

```sql
INSERT INTO orders DEFAULT VALUES;
```

Result:

```text
status = pending
```

This is useful for predictable initial state.

---

# 9. DEFAULT Does Not Mean Immutable

A default only applies when the value is not supplied during insertion.

Example:

```sql
status TEXT NOT NULL DEFAULT 'active'
```

After creation:

```sql
UPDATE users
SET status = 'inactive'
WHERE id = 1;
```

The default does not restore `active`.

Think:

```text
INSERT without value
        ↓
      DEFAULT

UPDATE
        ↓
default is irrelevant
```

---

# 10. Common Status Default

Status columns frequently have a default.

Example:

```sql
CREATE TABLE orders (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    status TEXT NOT NULL DEFAULT 'pending'
);
```

Typical lifecycle:

```text
pending
   ↓
processing
   ↓
completed
```

The database can guarantee that a newly inserted order starts with a valid default state.

If only specific statuses are allowed, combine this with a constraint.

Example:

```sql
status TEXT NOT NULL DEFAULT 'pending'
CHECK (status IN ('pending', 'processing', 'completed', 'cancelled'))
```

---

# 11. Boolean Defaults

Boolean flags often have clear defaults.

Example:

```sql
is_active BOOLEAN NOT NULL DEFAULT TRUE
```

This avoids unnecessary three-state behavior.

Instead of:

```text
true
false
NULL
```

you get:

```text
true
false
```

This makes application logic simpler when `NULL` has no meaningful business meaning.

---

# 12. Nullable Boolean

A nullable boolean can represent three states:

```text
TRUE
FALSE
NULL
```

This is sometimes useful.

For example:

```text
email_verified
```

could mean:

```text
TRUE  = verified
FALSE = explicitly not verified
NULL  = verification status unknown
```

But if there is no meaningful third state, prefer:

```sql
BOOLEAN NOT NULL DEFAULT FALSE
```

Do not use nullable booleans accidentally.

---

# 13. Timestamp Defaults

A common production pattern:

```sql
created_at TIMESTAMPTZ NOT NULL DEFAULT now()
```

Example:

```sql
CREATE TABLE users (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    name TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
```

Now the application does not need to provide `created_at` for every insert.

PostgreSQL generates it.

---

# 14. created_at and updated_at

A common schema:

```sql
CREATE TABLE users (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    name TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
```

Important:

> `DEFAULT now()` does not automatically update `updated_at` whenever the row changes.

It only supplies the initial value.

Updating `updated_at` requires application logic, a trigger, or another deliberate database mechanism.

Do not assume the default provides automatic update behavior.

---

# 15. Database-Generated IDs

Identity columns are a form of database-generated default behavior.

Example:

```sql
id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY
```

The application can insert:

```sql
INSERT INTO users (name)
VALUES ('Alice')
RETURNING id;
```

PostgreSQL generates the ID.

This is preferable to manually coordinating numeric IDs in application code.

---

# 16. GENERATED ALWAYS vs BY DEFAULT

PostgreSQL supports:

```sql
GENERATED ALWAYS AS IDENTITY
```

and:

```sql
GENERATED BY DEFAULT AS IDENTITY
```

For most application code, the important production concept is:

> Let PostgreSQL own generated identity values unless there is a deliberate reason to override them.

`GENERATED ALWAYS` is stricter about manually supplying values.

You do not need to memorize every override syntax to use identity columns effectively.

---

# 17. DEFAULT and NULL

Remember the distinction:

```text
DEFAULT
  ↓
value used when column is omitted

NULL
  ↓
explicit absence of a value
```

Example:

```sql
status TEXT DEFAULT 'active'
```

Then:

```sql
INSERT INTO users DEFAULT VALUES;
```

produces:

```text
active
```

while:

```sql
INSERT INTO users (status)
VALUES (NULL);
```

produces `NULL` unless another constraint rejects it.

---

# 18. DEFAULT and Empty String

A default does not convert empty strings.

Example:

```sql
name TEXT NOT NULL DEFAULT 'unknown'
```

This:

```sql
INSERT INTO users (name)
VALUES ('');
```

stores:

```text
''
```

not:

```text
unknown
```

If empty strings are invalid, define that rule explicitly.

---

# 19. DEFAULT and Zero

Likewise:

```sql
quantity INTEGER NOT NULL DEFAULT 0
```

does not mean every missing or invalid value becomes zero.

It means:

```text
column omitted
      ↓
use 0
```

This:

```sql
INSERT INTO products (quantity)
VALUES (NULL);
```

still supplies `NULL`.

If `NOT NULL` exists, the insert fails.

---

# 20. DEFAULT Expressions

Defaults can be expressions.

Examples:

```sql
created_at TIMESTAMPTZ DEFAULT now()
```

```sql
is_active BOOLEAN DEFAULT TRUE
```

```sql
status TEXT DEFAULT 'pending'
```

The expression is evaluated by PostgreSQL when a row is inserted without an explicit value.

This is useful when the default should come from the database environment or a deterministic expression.

---

# 21. Where Should a Default Live?

Ask:

> Is this the database's natural invariant or purely an API concern?

Good database default:

```sql
created_at TIMESTAMPTZ NOT NULL DEFAULT now()
```

Good database default:

```sql
is_active BOOLEAN NOT NULL DEFAULT TRUE
```

Good database default:

```sql
status TEXT NOT NULL DEFAULT 'pending'
```

These describe valid database state.

A UI-specific default may belong in the application instead.

---

# 22. Database Defaults Improve Safety

Suppose five different code paths create orders:

```text
HTTP API
Worker
Admin tool
Import job
Migration
```

If every path must remember:

```text
status = pending
```

one path may forget.

A database default provides a final safety net.

```text
multiple writers
      ↓
database
      ↓
consistent default
```

This is one reason defaults matter in production.

---

# 23. Application Validation Still Matters

A database default does not replace request validation.

For example:

```sql
status TEXT NOT NULL DEFAULT 'pending'
```

does not mean the API should accept arbitrary statuses.

The application may validate:

```text
status ∈ {pending, processing, completed, cancelled}
```

The database can also enforce the invariant:

```sql
CHECK (
    status IN (
        'pending',
        'processing',
        'completed',
        'cancelled'
    )
)
```

Application validation improves UX.

Database constraints protect stored data.

---

# 24. Defaults and API DTOs

Suppose the API request is:

```json
{
  "name": "Alice"
}
```

The database can provide:

```text
status = active
created_at = current timestamp
```

The API does not necessarily need to expose those fields as required input.

This creates a clean separation:

```text
client input
    ↓
application validation
    ↓
database defaults
    ↓
stored row
```

The API should document what clients may omit.

---

# 25. Explicit INSERT Columns

Even when defaults exist, use explicit column lists.

Prefer:

```sql
INSERT INTO users (name)
VALUES ($1);
```

Avoid:

```sql
INSERT INTO users
VALUES ($1);
```

Explicit columns make the query resilient to schema changes and make it clear which fields are intentionally omitted and therefore may receive defaults.

---

# 26. DEFAULT in INSERT

You can explicitly request a default:

```sql
INSERT INTO users (name, status)
VALUES ('Alice', DEFAULT);
```

This is useful in generated or complex SQL.

But in normal application code, simply omitting a column is often clearer:

```sql
INSERT INTO users (name)
VALUES ('Alice');
```

---

# 27. RETURNING After INSERT

When the database generates values, `RETURNING` is useful.

Example:

```sql
INSERT INTO users (name)
VALUES ($1)
RETURNING id, status, created_at;
```

The application immediately receives:

```text
id
status
created_at
```

This avoids an unnecessary follow-up query in many cases.

---

# 28. Defaults and Upsert

Defaults interact with insert behavior.

Example:

```sql
INSERT INTO users (email)
VALUES ($1)
ON CONFLICT (email)
DO UPDATE
SET updated_at = now()
RETURNING *;
```

The default applies to a newly inserted row when the column is omitted.

The `DO UPDATE` branch follows update semantics instead.

Do not assume an upsert will reapply insert defaults during conflict updates.

---

# 29. Defaults Are Not a Substitute for Constraints

Consider:

```sql
status TEXT NOT NULL DEFAULT 'pending'
```

This guarantees:

- no `NULL`
- omitted value becomes `pending`

But it does not necessarily guarantee:

```text
status is one of the allowed statuses
```

Add:

```sql
CHECK (
    status IN ('pending', 'processing', 'completed', 'cancelled')
)
```

Different rules require different constraints.

---

# 30. Choosing NULL vs Default

This is an important design decision.

Use `NULL` when:

> The absence of a value has meaningful business meaning.

Example:

```text
assigned_to = NULL
```

could mean:

> Nobody has been assigned yet.

Use a default when:

> A new row should start in a known state.

Example:

```text
status = pending
```

Do not replace meaningful `NULL` with arbitrary placeholder values such as:

```text
''
0
false
unknown
```

unless those values genuinely represent the domain.

---

# 31. Placeholder Values Can Be Dangerous

Suppose:

```sql
assigned_user_id BIGINT DEFAULT 0
```

and `0` does not represent a real user.

Now the database contains fake relationships.

Better:

```sql
assigned_user_id BIGINT NULL
```

where `NULL` means:

```text
not assigned
```

Use real domain values, not technical placeholders.

---

# 32. Defaults for Counters

Some counters naturally start at zero.

Example:

```sql
CREATE TABLE posts (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    view_count BIGINT NOT NULL DEFAULT 0
);
```

This is reasonable if:

```text
0 = no views yet
```

The value is meaningful and unambiguous.

---

# 33. Defaults for State

State fields are also common:

```sql
status TEXT NOT NULL DEFAULT 'draft'
```

Example:

```text
new article
    ↓
draft
```

The default establishes the initial state.

State transitions should still be controlled by application/business logic and, where appropriate, database constraints.

---

# 34. Defaults for Configuration

Some configuration values have safe defaults.

Example:

```sql
CREATE TABLE notification_settings (
    user_id BIGINT PRIMARY KEY,
    email_enabled BOOLEAN NOT NULL DEFAULT TRUE,
    push_enabled BOOLEAN NOT NULL DEFAULT TRUE
);
```

This makes newly created settings predictable.

But verify that the default is actually safe for every user and every business context.

---

# 35. Adding NOT NULL to Existing Tables

This is where production migration thinking matters.

Suppose:

```sql
ALTER TABLE users
ALTER COLUMN name SET NOT NULL;
```

This can fail if existing rows contain:

```text
name = NULL
```

Before changing the constraint, inspect:

```sql
SELECT COUNT(*)
FROM users
WHERE name IS NULL;
```

Then decide how to repair the existing data.

---

# 36. Safe Migration Pattern

A common approach:

```text
1. inspect existing data
2. backfill valid values
3. verify no invalid rows remain
4. add NOT NULL
```

Example:

```sql
UPDATE users
SET name = 'Unknown'
WHERE name IS NULL;
```

Only use a backfill like this if `'Unknown'` is genuinely valid business data.

Do not destroy semantic information merely to satisfy a migration.

---

# 37. Adding a DEFAULT to Existing Data

Adding a default does not automatically mean every historical row receives that value.

Think separately about:

```text
future inserts
```

and:

```text
existing rows
```

If you need existing rows updated, perform an explicit data migration.

Example:

```sql
UPDATE users
SET status = 'active'
WHERE status IS NULL;
```

Then add:

```sql
ALTER TABLE users
ALTER COLUMN status SET DEFAULT 'active';
```

And, if appropriate:

```sql
ALTER TABLE users
ALTER COLUMN status SET NOT NULL;
```

---

# 38. Schema Evolution

When adding a new required column to a large production table, be careful.

Conceptually:

```text
new column
   ↓
existing rows need a valid value
   ↓
application versions may differ
   ↓
migration must be compatible
```

A safer rollout may be:

```text
1. add nullable column
2. deploy application support
3. backfill
4. verify
5. add default
6. add NOT NULL if appropriate
```

Exact rollout depends on table size, deployment strategy, and PostgreSQL version.

---

# 39. Avoid Overusing Defaults

Not every column needs a default.

Example:

```sql
email TEXT NOT NULL
```

There may be no sensible default email.

Do not invent:

```text
unknown@example.com
```

just to avoid handling missing data.

A default should represent a legitimate initial value.

---

# 40. Common Mistake: Thinking NOT NULL Rejects Bad Data

`NOT NULL` only protects against `NULL`.

It does not validate:

```text
empty strings
negative numbers
invalid statuses
malformed business identifiers
```

Use:

```text
NOT NULL
CHECK
UNIQUE
FK
application validation
```

according to the rule being enforced.

---

# 41. Common Mistake: Thinking DEFAULT Handles NULL

Given:

```sql
status TEXT NOT NULL DEFAULT 'active'
```

this does not work:

```sql
INSERT INTO users (status)
VALUES (NULL);
```

The default is not a replacement for explicitly supplied `NULL`.

The insert violates `NOT NULL`.

---

# 42. Common Mistake: Assuming DEFAULT Runs on UPDATE

Given:

```sql
updated_at TIMESTAMPTZ DEFAULT now()
```

this does not automatically update:

```text
updated_at
```

after:

```sql
UPDATE users
SET name = 'Bob'
WHERE id = 1;
```

If automatic update behavior is required, implement it deliberately.

---

# 43. Common Mistake: Using Fake Defaults

Bad:

```sql
age INTEGER NOT NULL DEFAULT 0
```

if `0` actually means:

```text
age unknown
```

Better:

```sql
age INTEGER NULL
```

if unknown is a legitimate state.

The default must have real domain meaning.

---

# 44. Common Mistake: Everything Nullable

A schema like:

```sql
name TEXT
email TEXT
status TEXT
created_at TIMESTAMPTZ
```

may push too much uncertainty into application code.

If these values are required:

```sql
name TEXT NOT NULL,
email TEXT NOT NULL,
status TEXT NOT NULL DEFAULT 'active',
created_at TIMESTAMPTZ NOT NULL DEFAULT now()
```

The schema becomes much more explicit.

---

# 45. Common Mistake: Everything Has a Default

The opposite is also bad.

Forcing defaults onto meaningful required inputs can hide missing data.

For example:

```sql
customer_name TEXT DEFAULT 'Unknown'
```

may allow an incomplete order to look valid.

If the value is required from the business perspective, prefer:

```sql
customer_name TEXT NOT NULL
```

and require the application to provide it.

---

# 46. Production Pattern: User Table

```sql
CREATE TABLE users (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    email TEXT NOT NULL UNIQUE,
    name TEXT NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
```

Responsibilities:

```text
id
  → identity

email
  → required + unique

name
  → required

is_active
  → known initial state

created_at
  → database-generated creation time

updated_at
  → initial timestamp; update behavior must be implemented separately
```

---

# 47. Production Pattern: Order Table

```sql
CREATE TABLE orders (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    customer_id BIGINT NOT NULL,
    status TEXT NOT NULL DEFAULT 'pending',
    total NUMERIC(12, 2) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    CHECK (
        status IN (
            'pending',
            'processing',
            'completed',
            'cancelled'
        )
    ),

    FOREIGN KEY (customer_id)
        REFERENCES customers(id)
);
```

The schema expresses:

```text
customer is required
status always has a valid initial value
total is required
created_at is generated
status has an allowed domain
customer relationship is valid
```

---

# 48. Go Application Perspective

A Go repository may execute:

```go
row := db.QueryRowContext(ctx, `
    INSERT INTO users (email, name)
    VALUES ($1, $2)
    RETURNING id, status, created_at
`, email, name)
```

The application does not need to manually provide every database-generated value.

PostgreSQL handles:

```text
id
status
created_at
```

and returns the final values.

This keeps database responsibilities explicit.

---

# 49. Production Review Questions

For each column, ask:

### Requiredness

- Can this ever legitimately be `NULL`?
- If not, is it `NOT NULL`?

### Initial value

- Should a new row have a known value?
- Is there a meaningful default?

### Validation

- Does `NOT NULL` cover the whole rule?
- Is `CHECK` needed?
- Is application validation also needed?

### Generation

- Should PostgreSQL generate the value?
- Should the application generate it?

### API behavior

- Can clients omit this field?
- What happens when they omit it?

### Migration

- What happens to existing rows?
- Can the constraint be added safely?

---

# 50. Practical Mental Model

Think of the two keywords this way:

```text
NOT NULL
   ↓
"Stored data must have a value."

DEFAULT
   ↓
"If insert omits this column,
 PostgreSQL supplies a value."
```

Together:

```sql
status TEXT NOT NULL DEFAULT 'pending'
```

means:

```text
INSERT without status
        ↓
   pending

INSERT with valid status
        ↓
   provided value

INSERT with NULL
        ↓
   rejected
```

---

# 51. Production Checklist

Before shipping a table, check:

- Required business fields use `NOT NULL`.
- Optional fields intentionally allow `NULL`.
- Defaults represent real domain values.
- Generated timestamps use an appropriate PostgreSQL type.
- Identity columns are used for database-generated IDs where appropriate.
- Boolean columns are not nullable unless three states are meaningful.
- Status defaults represent valid initial states.
- `CHECK` constraints protect finite domains when appropriate.
- Application validation still handles request-level rules.
- `RETURNING` is used when generated values are needed immediately.
- Existing rows are considered before adding `NOT NULL`.
- Existing rows are considered separately from future insert defaults.
- `updated_at` behavior is implemented deliberately rather than assumed from `DEFAULT now()`.

---

# 52. Production Takeaways

1. `NOT NULL` protects requiredness.
2. `DEFAULT` supplies omitted insert values.
3. `DEFAULT` does not replace explicit `NULL`.
4. `DEFAULT` does not automatically run on updates.
5. Use `NULL` when absence has real business meaning.
6. Do not use fake placeholder defaults to hide missing data.
7. Use defaults for predictable initial states.
8. Use database constraints for database integrity.
9. Use application validation for request-level behavior and UX.
10. Treat schema migrations as data migrations too.
11. Database-generated values can reduce duplicated application logic.
12. Keep the schema explicit about what is required, optional, and automatically generated.

Core mental model:

> **`NOT NULL` answers "must this exist?"  
> `DEFAULT` answers "what should happen if it is omitted?"**
