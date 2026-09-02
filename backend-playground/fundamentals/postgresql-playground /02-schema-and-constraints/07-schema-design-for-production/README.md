# Schema Design for Production

Production-first PostgreSQL material for turning database design decisions into a schema that is safe to operate, query, migrate, and evolve.

---

## 1. What This Material Covers

This chapter focuses on:

- production table design
- naming conventions
- primary keys
- foreign keys
- required and optional fields
- audit timestamps
- soft delete
- status columns
- indexes
- constraints
- tenant boundaries
- migrations
- backward-compatible schema changes
- data lifecycle
- schema review
- common production mistakes

Core idea:

> **A production schema is not just a collection of tables. It is a contract for data integrity, application behavior, and future changes.**

---

# 2. Start With the Domain

Before writing SQL, identify the actual entities.

Example e-commerce system:

```text
Customer
Order
OrderItem
Product
Payment
```

Then define what one row represents:

```text
customers
→ one customer

orders
→ one order

order_items
→ one product line inside one order

products
→ one product

payments
→ one payment
```

This prevents vague table designs.

---

# 3. Define Ownership

For each field, ask:

> Which entity owns this fact?

Example:

```text
customers
    email
    name

orders
    status
    total
    customer_id

products
    name
    price
```

Avoid putting customer fields into orders merely because an order query needs customer information.

Use the relationship:

```text
orders.customer_id
    ↓
customers.id
```

unless the copied value has a deliberate historical meaning.

---

# 4. Use Clear Table Names

Choose a consistent naming convention.

A common PostgreSQL style is:

```text
snake_case
```

Examples:

```text
users
user_profiles
order_items
payment_transactions
```

Avoid inconsistent naming such as:

```text
Users
userProfile
ORDER_ITEMS
```

Consistency reduces friction in:

- SQL
- migrations
- ORM mappings
- debugging
- code review

---

# 5. Use Clear Column Names

Prefer names that explain the value.

Good:

```text
created_at
updated_at
deleted_at
customer_id
approved_by
published_at
```

Avoid vague names:

```text
date
type
value
status1
user
```

unless the meaning is obvious from the domain.

A foreign key should usually communicate the relationship:

```text
customer_id
created_by
assigned_to
approved_by
```

---

# 6. Primary Keys

Most application tables should have a stable primary key.

Example:

```sql
CREATE TABLE users (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    name TEXT NOT NULL
);
```

The primary key should generally be:

- unique
- non-null
- stable
- independent of mutable business data

Avoid using:

```text
email
phone_number
username
```

as the primary identity when those values can change.

---

# 7. ID Strategy

A practical PostgreSQL choice:

```sql
id BIGINT GENERATED ALWAYS AS IDENTITY
```

For distributed systems or public identifiers, UUID can also be appropriate:

```sql
id UUID PRIMARY KEY
```

The important production rule is consistency.

If related tables use numeric IDs:

```text
users.id       BIGINT
orders.user_id BIGINT
```

keep the types aligned.

Do not choose an ID type only because it is fashionable.

Choose based on:

- system boundaries
- scale
- public exposure
- generation strategy
- operational requirements

---

# 8. Business Uniqueness

Do not confuse identity with business uniqueness.

Example:

```sql
CREATE TABLE users (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    email TEXT NOT NULL,

    CONSTRAINT uq_users_email
        UNIQUE (email)
);
```

Here:

```text
id
→ identity

email
→ business uniqueness
```

This allows email to change without changing the row's identity.

---

# 9. Required Fields

Use:

```sql
NOT NULL
```

when a value is required for a valid row.

Example:

```sql
name TEXT NOT NULL
```

Avoid making everything nullable.

A schema full of unnecessary `NULL`s creates uncertainty in:

- application code
- queries
- API responses
- reporting

Ask whether the absence of a value has real business meaning.

---

# 10. Optional Fields

Allow `NULL` when the absence of a value is meaningful.

Example:

```sql
assigned_to BIGINT NULL
```

can mean:

```text
not assigned yet
```

This is different from inventing:

```text
assigned_to = 0
```

Use `NULL` when it accurately represents the domain.

---

# 11. Defaults

Use database defaults for sensible initial values.

Example:

```sql
status TEXT NOT NULL DEFAULT 'pending'
```

or:

```sql
created_at TIMESTAMPTZ NOT NULL DEFAULT now()
```

or:

```sql
is_active BOOLEAN NOT NULL DEFAULT TRUE
```

A default should represent a legitimate initial state.

Do not create fake defaults merely to avoid handling missing data.

---

# 12. Timestamps

A common production pattern:

```sql
created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
```

`TIMESTAMPTZ` is generally a good choice for event timestamps in applications that operate across time zones.

Remember:

```sql
DEFAULT now()
```

sets the initial value.

It does not automatically update `updated_at` on every update.

Implement update behavior deliberately.

---

# 13. Audit Columns

Common audit fields include:

```text
created_at
updated_at
deleted_at
```

Depending on the domain, you may also need:

```text
created_by
updated_by
deleted_by
```

Do not add every audit field automatically.

Ask what operational or business requirement it serves.

---

# 14. Soft Delete

For records that should remain in the database after logical deletion:

```sql
deleted_at TIMESTAMPTZ NULL
```

Example:

```sql
UPDATE users
SET deleted_at = now()
WHERE id = $1;
```

Active rows:

```sql
WHERE deleted_at IS NULL
```

Soft delete can support:

- recovery
- audit history
- historical references
- business reporting

But it introduces query complexity.

---

# 15. Soft Delete Is Not Free

Once a table uses:

```text
deleted_at
```

developers must remember that:

```sql
SELECT *
FROM users;
```

may include deleted records.

Typical application queries need:

```sql
WHERE deleted_at IS NULL
```

This affects:

- indexes
- uniqueness
- joins
- counts
- authorization
- reporting

Use soft delete only when the lifecycle requires it.

---

# 16. Soft Delete and Unique Values

Suppose:

```sql
email TEXT NOT NULL UNIQUE
deleted_at TIMESTAMPTZ
```

A deleted user's email remains reserved.

If the business allows reusing it after deletion, use a partial unique index:

```sql
CREATE UNIQUE INDEX uq_users_active_email
ON users (email)
WHERE deleted_at IS NULL;
```

Now uniqueness applies only to active rows.

The correct choice depends on business requirements.

---

# 17. Foreign Keys

Relationships should normally be represented with foreign keys.

Example:

```sql
customer_id BIGINT NOT NULL,

CONSTRAINT fk_orders_customer
    FOREIGN KEY (customer_id)
    REFERENCES customers(id)
```

This prevents invalid references.

The database should protect fundamental relationship integrity.

---

# 18. Delete Behavior

For every foreign key, consider:

```text
RESTRICT / NO ACTION
CASCADE
SET NULL
```

Example:

```text
order → order_item
```

may reasonably use:

```sql
ON DELETE CASCADE
```

because an order item has no independent meaning without its order.

But:

```text
customer → invoice
```

may need deletion to be restricted.

Choose based on lifecycle and business meaning.

---

# 19. Constraints

Production schemas should encode important invariants.

Examples:

```sql
price NUMERIC(12, 2)
    CHECK (price >= 0)
```

```sql
quantity INTEGER
    CHECK (quantity >= 0)
```

```sql
email TEXT NOT NULL UNIQUE
```

Constraints protect the database even when data comes from:

- APIs
- workers
- scripts
- imports
- admin tools

---

# 20. Keep Application Validation Too

Database constraints do not replace application validation.

Use application validation for:

- request format
- user-friendly errors
- authorization
- workflow rules
- external business logic

Use database constraints for:

- uniqueness
- requiredness
- referential integrity
- simple invariants

A production system often needs both.

---

# 21. Status Columns

Statuses are common:

```sql
status TEXT NOT NULL DEFAULT 'pending'
```

If the allowed values are finite, consider a `CHECK`:

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

Keep the status vocabulary explicit.

Avoid arbitrary strings when invalid states would cause serious problems.

---

# 22. Status Does Not Define the Whole Workflow

A constraint such as:

```sql
CHECK (status IN (...))
```

only defines valid values.

It does not automatically enforce:

```text
pending → processing
processing → completed
```

Workflow transitions belong in application logic or another deliberate mechanism.

Keep:

```text
valid states
```

separate from:

```text
allowed transitions
```

---

# 23. Indexes

Indexes are part of production schema design.

A common relationship index:

```sql
CREATE INDEX idx_orders_customer_id
ON orders(customer_id);
```

A common time-based query index:

```sql
CREATE INDEX idx_orders_created_at
ON orders(created_at);
```

But do not index every column.

Indexes have costs:

- storage
- insert/update overhead
- vacuum work
- maintenance
- cache usage

Design them around actual query patterns.

---

# 24. Composite Indexes

For:

```sql
SELECT id, total
FROM orders
WHERE customer_id = $1
ORDER BY created_at DESC
LIMIT 20;
```

a composite index may be useful:

```sql
CREATE INDEX idx_orders_customer_created
ON orders(customer_id, created_at DESC);
```

Index design should follow:

```text
WHERE
+
ORDER BY
+
LIMIT
```

and the real workload.

---

# 25. Index Foreign Keys Based on Queries

A foreign key protects integrity.

An index improves access.

Example:

```sql
FOREIGN KEY (customer_id)
REFERENCES customers(id)
```

and separately:

```sql
CREATE INDEX idx_orders_customer_id
ON orders(customer_id);
```

Whether the index is necessary depends on query patterns and table size.

Do not assume FK and index are the same thing.

---

# 26. Avoid Indexing Everything

Suppose a table has:

```text
20 columns
```

and you create:

```text
20 indexes
```

Writes may become more expensive.

A better process:

```text
identify important queries
        ↓
inspect filters/sorts/joins
        ↓
design indexes
        ↓
measure
```

Index for workload, not for column count.

---

# 27. Table Grain

Every production table should have a clear row meaning.

Examples:

```text
orders
→ one order

order_items
→ one item within one order

payments
→ one payment transaction

user_roles
→ one user-role relationship
```

If you cannot explain what one row represents, the schema probably needs clarification.

---

# 28. Avoid Mixed Grain

Bad design:

```text
orders
```

contains both:

```text
one order
```

and:

```text
multiple product rows
```

inside the same structure.

Better:

```text
orders
order_items
```

This avoids repeated order-level data and makes aggregation more reliable.

---

# 29. Naming Relationships

Prefer:

```text
customer_id
created_by
approved_by
assigned_to
```

over generic:

```text
user_id
```

when multiple relationships point to the same table.

Example:

```sql
created_by BIGINT NOT NULL
    REFERENCES users(id),

approved_by BIGINT
    REFERENCES users(id)
```

The column name communicates business meaning.

---

# 30. Many-to-Many

Use a junction table.

Example:

```sql
CREATE TABLE user_roles (
    user_id BIGINT NOT NULL,
    role_id BIGINT NOT NULL,

    PRIMARY KEY (user_id, role_id),

    FOREIGN KEY (user_id)
        REFERENCES users(id)
        ON DELETE CASCADE,

    FOREIGN KEY (role_id)
        REFERENCES roles(id)
        ON DELETE CASCADE
);
```

Avoid storing:

```text
role_ids = [1,2,3]
```

when the relationship needs normal relational querying and integrity.

---

# 31. Junction Table Indexing

If the primary key is:

```sql
PRIMARY KEY (user_id, role_id)
```

queries starting with:

```sql
WHERE user_id = $1
```

are naturally supported.

If you frequently query:

```sql
WHERE role_id = $1
```

add an index such as:

```sql
CREATE INDEX idx_user_roles_role_id
ON user_roles(role_id);
```

Design for both relationship directions when needed.

---

# 32. Multi-Tenant Schema

In a SaaS system, many tables may contain:

```text
organization_id
tenant_id
```

Example:

```sql
CREATE TABLE projects (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    organization_id BIGINT NOT NULL,
    name TEXT NOT NULL,

    FOREIGN KEY (organization_id)
        REFERENCES organizations(id)
);
```

Tenant ownership should be explicit.

---

# 33. Tenant Scoping

Application queries should normally scope tenant-owned data.

Example:

```sql
SELECT id, name
FROM projects
WHERE organization_id = $1
  AND id = $2;
```

Do not assume:

```text
id = $2
```

alone is enough when IDs are globally accessible but authorization is tenant-scoped.

Schema design and authorization design must work together.

---

# 34. Tenant-Specific Uniqueness

Suppose usernames are unique inside an organization.

Use:

```sql
UNIQUE (organization_id, username)
```

rather than:

```sql
UNIQUE (username)
```

if global uniqueness is not required.

This is a schema-level representation of the business rule:

```text
same username
    ↓
allowed in different tenants

same tenant + username
    ↓
not allowed
```

---

# 35. Data Types

Choose data types based on meaning.

Examples:

```text
count       → INTEGER/BIGINT
money       → NUMERIC
timestamp   → TIMESTAMPTZ
date only   → DATE
true/false  → BOOLEAN
identifier  → appropriate ID type
flexible doc → JSONB
```

Avoid storing structured values as `TEXT` when a stronger type exists.

---

# 36. Money

For exact monetary values, prefer:

```sql
NUMERIC(12, 2)
```

over floating-point types.

Example:

```sql
price NUMERIC(12, 2) NOT NULL
    CHECK (price >= 0)
```

The exact precision and scale should match the business domain.

Do not use floating point when exact decimal arithmetic is required.

---

# 37. JSONB

JSONB is useful for genuinely flexible or document-like data.

Example:

```sql
metadata JSONB
```

Good use cases may include:

- optional integration payloads
- flexible metadata
- provider-specific attributes

Avoid using JSONB to hide relationships that should have foreign keys.

Ask:

> Will this data need relational constraints and frequent relational queries?

If yes, a normal table/column may be better.

---

# 38. Schema Evolution

Production schemas change.

Examples:

```text
add column
remove column
rename column
add constraint
change data type
split table
merge table
```

Design migrations with application compatibility in mind.

Do not treat migrations as isolated SQL scripts.

They are part of application deployment.

---

# 39. Backward-Compatible Changes

A safer deployment often follows:

```text
database first
    ↓
application supports old + new
    ↓
backfill
    ↓
switch reads/writes
    ↓
remove old structure later
```

For example, adding a new column can be safer than immediately renaming/removing an old column used by currently running application instances.

This matters when multiple application versions coexist during deployment.

---

# 40. Expand and Contract

A common migration strategy:

```text
EXPAND
  ↓
add new structure
  ↓
deploy compatible application
  ↓
migrate/backfill data
  ↓
switch application
  ↓
CONTRACT
  ↓
remove old structure
```

This reduces the risk of a deployment where application code and database schema temporarily disagree.

---

# 41. Adding a Required Column

Suppose an existing large table needs:

```sql
new_status TEXT NOT NULL
```

Existing rows have no value.

A staged approach may be:

```text
1. add column
2. deploy code that understands it
3. backfill existing rows
4. add default if appropriate
5. enforce NOT NULL
```

Exact ordering depends on PostgreSQL behavior, table size, and deployment strategy.

---

# 42. Adding Constraints to Existing Data

Before adding:

```sql
UNIQUE
CHECK
NOT NULL
FOREIGN KEY
```

inspect existing rows.

Examples:

```sql
SELECT email, COUNT(*)
FROM users
GROUP BY email
HAVING COUNT(*) > 1;
```

```sql
SELECT *
FROM products
WHERE price < 0;
```

```sql
SELECT *
FROM orders o
LEFT JOIN customers c
    ON c.id = o.customer_id
WHERE c.id IS NULL;
```

New constraints must be compatible with existing data.

---

# 43. Migration Safety

Before a production migration, consider:

```text
table size
lock behavior
existing data
application compatibility
rollback strategy
deployment order
query impact
index build time
```

A migration that works on a local database with 100 rows may behave very differently with millions of rows.

---

# 44. Avoid Destructive Changes in One Step

Risky:

```text
drop old column
deploy new code
```

if old application instances may still reference it.

Safer:

```text
add new column
deploy compatibility
migrate data
switch code
verify
remove old column later
```

This is especially important in rolling or zero-downtime deployments.

---

# 45. Schema as an Application Contract

Application code depends on assumptions such as:

```text
user.email is never NULL
order.customer_id always exists
status is one of four values
created_at always exists
```

The database schema should enforce as many fundamental assumptions as practical.

Then:

```text
application assumptions
        ↕
database constraints
```

stay aligned.

---

# 46. Production Example

A practical users table:

```sql
CREATE TABLE users (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,

    email TEXT NOT NULL,

    name TEXT NOT NULL,

    is_active BOOLEAN NOT NULL DEFAULT TRUE,

    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    deleted_at TIMESTAMPTZ NULL,

    CONSTRAINT uq_users_email
        UNIQUE (email),

    CONSTRAINT ck_users_name_nonempty
        CHECK (length(trim(name)) > 0)
);
```

Possible responsibilities:

```text
id
→ stable identity

email
→ required + unique

name
→ required + non-empty

is_active
→ known initial state

created_at
→ creation timestamp

updated_at
→ initial modification timestamp

deleted_at
→ logical deletion state
```

---

# 47. Production Example: Orders

```sql
CREATE TABLE orders (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,

    customer_id BIGINT NOT NULL,

    status TEXT NOT NULL DEFAULT 'pending',

    total NUMERIC(12, 2) NOT NULL,

    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT fk_orders_customer
        FOREIGN KEY (customer_id)
        REFERENCES customers(id)
        ON DELETE RESTRICT,

    CONSTRAINT ck_orders_total
        CHECK (total >= 0),

    CONSTRAINT ck_orders_status
        CHECK (
            status IN (
                'pending',
                'processing',
                'completed',
                'cancelled'
            )
        )
);
```

Then, based on query patterns:

```sql
CREATE INDEX idx_orders_customer_created
ON orders(customer_id, created_at DESC);
```

---

# 48. Production Example: Order Items

```sql
CREATE TABLE order_items (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,

    order_id BIGINT NOT NULL,

    product_id BIGINT NOT NULL,

    quantity INTEGER NOT NULL,

    unit_price NUMERIC(12, 2) NOT NULL,

    CONSTRAINT fk_order_items_order
        FOREIGN KEY (order_id)
        REFERENCES orders(id)
        ON DELETE CASCADE,

    CONSTRAINT fk_order_items_product
        FOREIGN KEY (product_id)
        REFERENCES products(id)
        ON DELETE RESTRICT,

    CONSTRAINT ck_order_items_quantity
        CHECK (quantity > 0),

    CONSTRAINT ck_order_items_unit_price
        CHECK (unit_price >= 0)
);
```

Notice:

```text
order_id
→ required relationship

product_id
→ required relationship

quantity
→ positive

unit_price
→ non-negative

order deletion
→ item deletion
```

---

# 49. Production Review: Correctness

Ask:

- Does every table have a clear primary key?
- Are required values `NOT NULL`?
- Are relationships protected by foreign keys?
- Are unique business rules enforced?
- Are simple invariants protected with `CHECK`?
- Are defaults meaningful?
- Are delete behaviors intentional?

---

# 50. Production Review: Query Performance

Ask:

- What are the most important queries?
- Which columns are filtered?
- Which columns are used for joins?
- Which columns are sorted?
- Which queries use pagination?
- Are composite indexes needed?
- Are indexes too numerous?
- Have important queries been checked with `EXPLAIN`?

Do not design indexes in isolation from queries.

---

# 51. Production Review: Lifecycle

Ask:

- Can records be physically deleted?
- Is soft delete needed?
- What happens to children?
- What happens to historical records?
- Which values are immutable?
- Which values can change?
- Are snapshots required?

Schema design should reflect data lifecycle.

---

# 52. Production Review: Deployment

Ask:

- Can the migration run safely on production-sized data?
- Can old and new application versions coexist?
- Does the migration require a backfill?
- Could it create long locks?
- Is the change reversible?
- Can the old schema be removed later?

A correct schema change can still be an unsafe production deployment if the rollout strategy is wrong.

---

# 53. Production Review: Security and Authorization

The schema cannot replace authorization.

For example:

```sql
SELECT *
FROM projects
WHERE id = $1;
```

may find a project belonging to another organization.

Application logic should enforce tenant/user authorization.

Where appropriate, database-level mechanisms such as Row-Level Security can provide additional protection, but they should be introduced deliberately.

---

# 54. Schema Review Checklist

Before shipping a table:

### Identity

- Clear primary key
- Stable ID
- Consistent ID type

### Data

- Correct data types
- Required fields use `NOT NULL`
- Optional fields have meaningful `NULL`
- Defaults represent valid initial states

### Integrity

- Foreign keys
- Unique constraints
- Check constraints
- Explicit delete behavior

### Performance

- Important query paths identified
- Foreign-key access considered
- Composite indexes considered
- No unnecessary indexes

### Lifecycle

- Creation/update/deletion behavior defined
- Historical data requirements understood
- Soft delete used only when justified

### Deployment

- Existing data considered
- Migration compatibility considered
- Backfill strategy defined
- Large-table impact considered

---

# 55. Common Mistake: Designing Only for Today

Bad:

```text
works for current API
```

without considering:

```text
future migration
historical data
new application versions
additional query patterns
```

You do not need to predict everything.

But the schema should avoid unnecessary choices that make normal evolution difficult.

---

# 56. Common Mistake: No Constraints

A schema with:

```text
id
email
status
user_id
```

but no constraints may rely entirely on application code.

This is fragile.

At minimum, consider:

```text
PRIMARY KEY
NOT NULL
UNIQUE
FOREIGN KEY
CHECK
```

where the domain requires them.

---

# 57. Common Mistake: Too Many Constraints Without Meaning

Constraints should represent real invariants.

Do not add arbitrary rules such as:

```text
name must have exactly 10 characters
```

unless the business actually requires it.

The schema should be strict about correctness, not arbitrary.

---

# 58. Common Mistake: Overusing Soft Delete

Soft delete can make every query more complicated.

Before using it, ask:

```text
Do we need recovery?
Do historical references need the row?
Do audit requirements require retention?
Can records actually be deleted?
```

If not, physical deletion may be simpler.

---

# 59. Common Mistake: Over-Indexing

Indexes improve reads but add write and storage costs.

Avoid:

```text
index every column
```

Instead:

```text
important query
    ↓
filter/join/order analysis
    ↓
index
    ↓
EXPLAIN
```

---

# 60. Common Mistake: Premature Denormalization

Do not duplicate fields just because a join exists.

Start with:

```text
correct normalized model
```

then:

```text
measure
→ optimize
→ denormalize only when justified
```

If duplicated data is introduced, document the source of truth.

---

# 61. Common Mistake: Unsafe Migration

Avoid assuming:

```text
ALTER TABLE
```

is always instant or harmless.

Production tables may contain:

```text
millions of rows
active traffic
long-running transactions
replicas
multiple application versions
```

Migration design is part of schema design.

---

# 62. Common Mistake: Schema and Code Disagree

Example:

Application assumes:

```text
status is never NULL
```

Database allows:

```text
status NULL
```

or:

Application assumes:

```text
email is unique
```

Database allows duplicates.

These mismatches eventually become production bugs.

Keep schema constraints and application assumptions aligned.

---

# 63. Practical Schema Design Workflow

Use this sequence:

```text
1. Identify entities
        ↓
2. Define row grain
        ↓
3. Define ownership
        ↓
4. Choose data types
        ↓
5. Add primary keys
        ↓
6. Add relationships
        ↓
7. Add NOT NULL / DEFAULT
        ↓
8. Add UNIQUE / CHECK
        ↓
9. Decide delete behavior
        ↓
10. Identify real query patterns
        ↓
11. Add indexes
        ↓
12. Plan migrations
        ↓
13. Review lifecycle
        ↓
14. Review production rollout
```

This gives you a repeatable design process.

---

# 64. Practical Mental Model

Think of a production table as a contract:

```text
IDENTITY
   ↓
What row is this?

DATA
   ↓
What does the row represent?

CONSTRAINTS
   ↓
What states are valid?

RELATIONSHIPS
   ↓
What other entities does it depend on?

INDEXES
   ↓
How will the data be accessed?

LIFECYCLE
   ↓
How is the row created, changed, and deleted?

MIGRATION
   ↓
How can this contract evolve safely?
```

---

# 65. Production Takeaways

1. Design from the domain, not from the API response alone.
2. Give every table a clear row grain.
3. Use stable primary keys.
4. Keep related ID types consistent.
5. Use `NOT NULL` for genuinely required values.
6. Use `DEFAULT` for meaningful initial states.
7. Use foreign keys for important relationships.
8. Use `UNIQUE` and `CHECK` for database-level invariants.
9. Treat delete behavior as part of the data lifecycle.
10. Use soft delete only when the business actually needs it.
11. Design indexes from real query patterns.
12. Avoid over-indexing.
13. Normalize by default and denormalize deliberately.
14. Treat migrations as production operations, not just SQL files.
15. Prefer backward-compatible schema changes when multiple application versions may coexist.
16. Consider existing data before adding constraints.
17. Keep application assumptions and database constraints aligned.
18. Make the source of truth explicit whenever data is duplicated.
19. Review correctness, performance, lifecycle, security, and deployment separately.
20. Prefer the simplest schema that correctly represents the domain and can evolve safely.

Core mental model:

> **A production schema must protect data today and make change tomorrow safe.**
