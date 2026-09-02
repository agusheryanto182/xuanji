# Primary and Foreign Keys

Production-first PostgreSQL material for understanding primary keys and foreign keys.

---

## 1. What This Material Covers

This chapter focuses on the parts of primary and foreign keys that matter most in real backend systems:

- Primary key (PK)
- Foreign key (FK)
- Referential integrity
- ID type consistency
- `ON DELETE`
- `ON UPDATE`
- Nullable foreign keys
- One-to-many relationships
- Preventing orphan records
- Constraints vs application validation
- Indexing foreign keys
- Common production mistakes

The goal is not to memorize SQL syntax.

The goal is to understand:

> **Which table owns the identity, which table references it, and what the database should guarantee.**

---

# 2. Primary Key

A primary key uniquely identifies a row.

Example:

```sql
CREATE TABLE users (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    name TEXT NOT NULL
);
```

The important properties are:

1. Every row has a primary key.
2. The value is unique.
3. The value cannot be `NULL`.
4. The database uses it as the row's identity.

Example:

```text
users

id | name
---+-------
1  | Alice
2  | Bob
3  | Charlie
```

`id = 2` identifies exactly one user.

---

# 3. Why Primary Keys Matter in Production

Application code needs a stable way to refer to records.

For example:

```http
GET /users/123
```

The `123` normally identifies a specific database record.

Other tables can also reference the same identity:

```text
users
  |
  +---- orders
  |
  +---- payments
  |
  +---- sessions
```

Without a reliable primary key, relationships become much harder to enforce.

---

# 4. Primary Key Should Be Stable

A primary key should generally not represent information that users can change.

Bad example:

```sql
PRIMARY KEY (email)
```

Email addresses can change.

Better:

```sql
id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY
```

Then:

```text
id = 123
email = old@example.com
```

can become:

```text
id = 123
email = new@example.com
```

The user's identity remains the same.

---

# 5. Identity Columns

Modern PostgreSQL supports identity columns.

Example:

```sql
CREATE TABLE users (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    name TEXT NOT NULL
);
```

Insert:

```sql
INSERT INTO users (name)
VALUES ('Alice');
```

PostgreSQL generates the ID.

You can retrieve it:

```sql
INSERT INTO users (name)
VALUES ('Alice')
RETURNING id;
```

This is useful in backend code because the application immediately receives the created record's identity.

---

# 6. BIGINT vs INTEGER

For production systems, `BIGINT` is often a reasonable default for generated IDs.

Example:

```sql
id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY
```

The critical rule is consistency.

If the parent uses:

```sql
id BIGINT
```

the foreign key should use:

```sql
user_id BIGINT
```

Avoid mismatched identifier types.

---

# 7. Foreign Key

A foreign key represents a relationship to another table.

Example:

```sql
CREATE TABLE orders (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    user_id BIGINT NOT NULL,
    total NUMERIC(12, 2) NOT NULL,

    CONSTRAINT fk_orders_user
        FOREIGN KEY (user_id)
        REFERENCES users(id)
);
```

Here:

```text
users.id
    ↑
    |
orders.user_id
```

`orders.user_id` references `users.id`.

---

# 8. What the Foreign Key Guarantees

Suppose we have:

```text
users

id
---
1
2
3
```

This is valid:

```sql
INSERT INTO orders (user_id, total)
VALUES (2, 100.00);
```

Because user `2` exists.

This is rejected:

```sql
INSERT INTO orders (user_id, total)
VALUES (999, 100.00);
```

if user `999` does not exist.

The database protects referential integrity.

---

# 9. Referential Integrity

Referential integrity means references between tables remain valid.

Without a foreign key, this can happen:

```text
users

id
---
1
2

orders

id | user_id
---+--------
10 | 1
11 | 999
```

Order `11` points to a user that does not exist.

That is an orphan reference.

A foreign key prevents this.

---

# 10. Database Constraint vs Application Validation

Application validation is useful:

```text
API request
   ↓
validate user_id
   ↓
insert order
```

But validation alone is not enough.

There can be:

- multiple application instances
- background workers
- scripts
- migrations
- admin tools
- concurrent requests

Therefore:

> **Important data integrity rules should also be enforced by the database.**

Use application validation for user experience.

Use database constraints for integrity.

Usually you want both.

---

# 11. One-to-Many Relationship

The most common foreign-key relationship is one-to-many.

Example:

```text
One user
   |
   +---- many orders
```

Schema:

```sql
CREATE TABLE users (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    name TEXT NOT NULL
);

CREATE TABLE orders (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    user_id BIGINT NOT NULL,
    total NUMERIC(12, 2) NOT NULL,

    CONSTRAINT fk_orders_user
        FOREIGN KEY (user_id)
        REFERENCES users(id)
);
```

One `users.id` can be referenced by many `orders.user_id` values.

---

# 12. Parent and Child

It helps to think in terms of parent and child.

```text
users
  |
  +---- orders
```

`users` is the parent.

`orders` is the child.

The child stores the foreign key.

```text
orders.user_id → users.id
```

This mental model is useful when deciding deletion behavior.

---

# 13. ON DELETE

The database needs to know what should happen when a referenced parent is deleted.

Example:

```sql
FOREIGN KEY (user_id)
REFERENCES users(id)
ON DELETE RESTRICT
```

Common choices include:

- `RESTRICT`
- `NO ACTION`
- `CASCADE`
- `SET NULL`

Choose based on business semantics.

---

# 14. ON DELETE RESTRICT

`RESTRICT` prevents deleting the parent while dependent rows exist.

Example:

```sql
FOREIGN KEY (user_id)
REFERENCES users(id)
ON DELETE RESTRICT
```

If:

```text
users.id = 10
```

is referenced by orders, deleting user `10` is rejected.

This is often appropriate for important historical data.

For example:

```text
customer
   ↓
orders
```

You may not want deleting a customer to silently delete financial history.

---

# 15. ON DELETE CASCADE

`CASCADE` deletes child rows automatically.

Example:

```sql
FOREIGN KEY (user_id)
REFERENCES users(id)
ON DELETE CASCADE
```

If user `10` is deleted:

```text
users
10
 ↓
orders
101
102
103
```

the dependent orders are also deleted.

Use this when the child has no meaningful existence without the parent.

Good examples can include:

```text
user → temporary session
post → post attachment
order → order line
```

But do not use `CASCADE` simply because it is convenient.

---

# 16. ON DELETE SET NULL

`SET NULL` removes the relationship while keeping the child row.

Example:

```sql
CREATE TABLE posts (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    author_id BIGINT,

    CONSTRAINT fk_posts_author
        FOREIGN KEY (author_id)
        REFERENCES users(id)
        ON DELETE SET NULL
);
```

When the user is deleted:

```text
posts.author_id
       ↓
      NULL
```

The post remains.

This requires the foreign key column to allow `NULL`.

---

# 17. Choosing ON DELETE

Think about the business meaning.

### Use RESTRICT / NO ACTION when:

The child record must not disappear automatically.

Example:

```text
customer → invoice
```

### Use CASCADE when:

The child is owned by the parent and has no independent meaning.

Example:

```text
order → order_item
```

### Use SET NULL when:

The child can survive without the parent.

Example:

```text
post → author
```

where anonymous/orphaned posts are acceptable.

---

# 18. ON UPDATE

You will see:

```sql
ON UPDATE CASCADE
```

less often in typical backend schemas.

If primary keys are stable and never changed, there is usually little reason to update them.

A common production approach is:

> Create stable IDs and treat them as immutable.

Then `ON UPDATE` becomes much less important.

---

# 19. Nullable Foreign Keys

A foreign key does not automatically mean `NOT NULL`.

This is valid:

```sql
author_id BIGINT
```

with:

```sql
FOREIGN KEY (author_id)
REFERENCES users(id)
```

The column may contain:

```text
1
2
NULL
```

If the relationship is required, use:

```sql
author_id BIGINT NOT NULL
```

Example:

```text
Every order must belong to a user
```

Then:

```sql
user_id BIGINT NOT NULL
```

is appropriate.

---

# 20. Optional Relationships

Suppose a support ticket may or may not have an assigned employee.

```sql
assigned_to BIGINT NULL
```

This represents:

```text
ticket
  |
  +---- assigned employee
          |
          +---- optional
```

`NULL` means:

> No employee is currently assigned.

This is different from an invalid user ID.

---

# 21. Required vs Optional Relationship

Ask:

> Can this record logically exist without the referenced record?

If no:

```sql
user_id BIGINT NOT NULL
```

If yes:

```sql
user_id BIGINT NULL
```

This decision should come from the business domain, not from SQL syntax.

---

# 22. Foreign Keys Do Not Mean Automatic Indexes

A foreign key constraint enforces correctness.

It does not mean you should assume the child column has the index you need for application queries.

For example:

```sql
orders.user_id
```

may be queried frequently:

```sql
SELECT id, total
FROM orders
WHERE user_id = $1
ORDER BY id DESC
LIMIT 20;
```

An index may be useful:

```sql
CREATE INDEX idx_orders_user_id
ON orders(user_id);
```

Or, if the query needs ordering:

```sql
CREATE INDEX idx_orders_user_id_id
ON orders(user_id, id DESC);
```

Index based on actual query patterns.

---

# 23. Why Foreign-Key Indexes Matter for Deletes

Consider:

```text
users
   |
   +---- orders
```

When deleting or updating a parent, PostgreSQL must maintain the foreign-key relationship.

Indexes on child foreign-key columns can be important for efficient referential checks, especially on larger tables.

So when creating a foreign key, ask:

> Will this child column also need an index?

Do not blindly add every possible index, but do not ignore FK access patterns either.

---

# 24. Composite Primary Keys

A primary key can contain multiple columns.

Example:

```sql
CREATE TABLE order_items (
    order_id BIGINT NOT NULL,
    product_id BIGINT NOT NULL,
    quantity INTEGER NOT NULL,

    PRIMARY KEY (order_id, product_id)
);
```

This means:

```text
(order_id, product_id)
```

must be unique.

This can be appropriate for junction tables.

---

# 25. Composite Foreign Keys

A foreign key can also reference multiple columns.

Example:

```sql
FOREIGN KEY (a, b)
REFERENCES parent(a, b)
```

Use composite foreign keys when the relationship itself is naturally identified by multiple columns.

Do not introduce composite keys just because they are technically possible.

For many application systems, a single stable ID plus additional unique constraints is simpler.

---

# 26. Many-to-Many Relationships

A many-to-many relationship normally uses a junction table.

Example:

```text
users
  |
  +---- user_roles ----+
                       |
roles -----------------+
```

Schema:

```sql
CREATE TABLE users (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    name TEXT NOT NULL
);

CREATE TABLE roles (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    name TEXT NOT NULL
);

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

---

# 27. Preventing Duplicate Relationships

The junction table should usually prevent the same relationship from being inserted twice.

This:

```text
user_id | role_id
--------+--------
1       | 2
1       | 2
```

is normally undesirable.

A composite primary key solves it:

```sql
PRIMARY KEY (user_id, role_id)
```

Now:

```sql
INSERT INTO user_roles (user_id, role_id)
VALUES (1, 2);
```

followed by another identical insert will violate the constraint.

---

# 28. Foreign Key Naming

Explicit constraint names improve debugging.

Example:

```sql
CONSTRAINT fk_orders_user
    FOREIGN KEY (user_id)
    REFERENCES users(id)
```

Instead of relying entirely on generated names.

This makes database errors easier to understand.

For larger systems, consistent naming conventions are valuable.

---

# 29. Primary Key + Foreign Key Example

A realistic small schema:

```sql
CREATE TABLE customers (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    name TEXT NOT NULL
);

CREATE TABLE orders (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    customer_id BIGINT NOT NULL,
    total NUMERIC(12, 2) NOT NULL,

    CONSTRAINT fk_orders_customer
        FOREIGN KEY (customer_id)
        REFERENCES customers(id)
        ON DELETE RESTRICT
);

CREATE INDEX idx_orders_customer_id
    ON orders(customer_id);
```

The responsibilities are separated:

```text
PRIMARY KEY
    ↓
identifies a row

FOREIGN KEY
    ↓
protects a relationship

INDEX
    ↓
helps query performance
```

Do not confuse these three jobs.

---

# 30. Primary Key vs UNIQUE

A primary key:

```sql
PRIMARY KEY (id)
```

identifies the row.

A unique constraint:

```sql
UNIQUE (email)
```

prevents duplicate values.

Example:

```sql
CREATE TABLE users (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    email TEXT NOT NULL UNIQUE
);
```

Here:

```text
id
```

is the identity.

```text
email
```

is a business attribute that must be unique.

This is a common production pattern.

---

# 31. Foreign Key vs Unique

A foreign key answers:

> Does this value reference a valid parent?

A unique constraint answers:

> Can this value appear more than once?

They solve different problems.

Example:

```sql
user_id BIGINT NOT NULL
    REFERENCES users(id)
```

allows many orders for one user.

Adding:

```sql
UNIQUE (user_id)
```

would change the relationship to one-to-one.

---

# 32. One-to-One Relationships

A one-to-one relationship can be enforced with a unique foreign key.

Example:

```sql
CREATE TABLE user_profiles (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    user_id BIGINT NOT NULL UNIQUE,

    FOREIGN KEY (user_id)
        REFERENCES users(id)
        ON DELETE CASCADE
);
```

Because `user_id` is unique:

```text
one user → at most one profile
```

Without `UNIQUE`, it would be one-to-many.

---

# 33. Database-Enforced Cardinality

The database can enforce important relationship rules.

### One-to-many

```sql
FOREIGN KEY (user_id)
REFERENCES users(id)
```

### One-to-one

```sql
FOREIGN KEY (user_id)
REFERENCES users(id)

UNIQUE (user_id)
```

### Many-to-many

```sql
PRIMARY KEY (user_id, role_id)
```

This is why schema design matters.

The database is not just storing data.

It is enforcing the shape of the data.

---

# 34. Avoid Application-Only Relationship Checks

Avoid relying on:

```text
SELECT user
if user exists:
    INSERT order
```

as the only protection.

Two requests can race.

Example:

```text
Request A                  Request B

check user 10              check user 10
     ↓                          ↓
user exists                user exists
     ↓                          ↓
insert order               insert order
```

That is fine if the user still exists.

But if other operations can delete or modify the parent, application checks alone do not provide database-level integrity.

The foreign key remains the final authority.

---

# 35. Transactions and Foreign Keys

Foreign-key operations often participate naturally in transactions.

Example:

```sql
BEGIN;

INSERT INTO users (name)
VALUES ('Alice')
RETURNING id;

INSERT INTO orders (user_id, total)
VALUES (1, 100.00);

COMMIT;
```

If a constraint fails:

```sql
ROLLBACK;
```

The database keeps the transaction consistent.

In backend code, transaction boundaries should cover the related writes that must succeed or fail together.

---

# 36. Deleting Parent Records in Production

Before deleting a parent, understand its dependencies.

Example:

```text
customer
  ├── orders
  ├── invoices
  ├── payments
  └── support_tickets
```

A simple:

```sql
DELETE FROM customers
WHERE id = $1;
```

may fail because of foreign keys.

That failure is useful.

It tells you:

> The schema is protecting related data.

Do not remove the constraint just to make the delete succeed.

Instead, decide the correct business behavior.

---

# 37. Soft Delete and Foreign Keys

Production systems often avoid physically deleting important records.

Instead:

```sql
deleted_at TIMESTAMPTZ NULL
```

can mark a record as deleted.

Then:

```text
customer
   |
   +---- orders
```

can remain intact.

This can be useful for:

- auditability
- historical records
- recovery
- compliance requirements
- preserving relationships

Soft delete is a business decision, not a universal rule.

---

# 38. IDs Across Services

In a single PostgreSQL database:

```text
users.id
orders.user_id
```

can use database-enforced foreign keys.

Across separate databases/services, a normal PostgreSQL foreign key cannot enforce the relationship across databases.

For example:

```text
User Service DB
    users.id

Order Service DB
    orders.user_id
```

The application must handle cross-service consistency.

This is an important architectural boundary:

> **Foreign keys are strongest inside the same database boundary.**

---

# 39. Migration Ordering

When adding a parent and child table, create the parent first.

Example:

```sql
CREATE TABLE users (...);

CREATE TABLE orders (
    ...
    FOREIGN KEY (user_id)
        REFERENCES users(id)
);
```

When removing them, dependencies matter.

You cannot casually remove the parent while child constraints still depend on it.

Migration order should respect the dependency graph.

---

# 40. Adding a Foreign Key to Existing Data

This is a common production migration problem.

Suppose existing data contains:

```text
orders.user_id = 999
```

but user `999` does not exist.

Adding the foreign key will fail.

Before adding the constraint, find invalid references:

```sql
SELECT o.id, o.user_id
FROM orders o
LEFT JOIN users u
    ON u.id = o.user_id
WHERE u.id IS NULL;
```

Then decide how to clean or repair the data.

Do not assume historical data already satisfies the new constraint.

---

# 41. Adding NOT NULL to an Existing Foreign Key

This also requires care.

First find:

```sql
SELECT COUNT(*)
FROM orders
WHERE user_id IS NULL;
```

If rows exist, you need a migration strategy.

Possible approach:

```text
1. identify invalid rows
2. repair/backfill them
3. verify
4. add NOT NULL
```

Production schema changes should consider existing data, not just the desired final schema.

---

# 42. Common Mistake: Missing Foreign Key

Bad:

```sql
CREATE TABLE orders (
    id BIGINT PRIMARY KEY,
    user_id BIGINT
);
```

Nothing guarantees that `user_id` exists in `users`.

Better:

```sql
CREATE TABLE orders (
    id BIGINT PRIMARY KEY,
    user_id BIGINT NOT NULL,

    FOREIGN KEY (user_id)
        REFERENCES users(id)
);
```

---

# 43. Common Mistake: Wrong Delete Behavior

Bad reasoning:

> `CASCADE` is easier, so use it everywhere.

This can cause accidental data loss.

Instead ask:

```text
If the parent disappears,
should the child:

1. disappear?
2. remain?
3. lose the relationship?
4. be impossible to delete?
```

Then choose:

```text
CASCADE
RESTRICT / NO ACTION
SET NULL
```

accordingly.

---

# 44. Common Mistake: Foreign Key Without NOT NULL

If every order must belong to a customer:

Bad:

```sql
customer_id BIGINT
```

Better:

```sql
customer_id BIGINT NOT NULL
```

The foreign key protects validity.

`NOT NULL` protects requiredness.

They solve different problems.

---

# 45. Common Mistake: Assuming FK Means Index

This:

```sql
FOREIGN KEY (user_id)
REFERENCES users(id)
```

does not mean every query involving `user_id` is automatically optimal.

Check real queries:

```sql
SELECT *
FROM orders
WHERE user_id = $1;
```

If this is frequent and the table is large, an index may be necessary.

---

# 46. Common Mistake: Using Mutable Business Data as Identity

Avoid:

```sql
PRIMARY KEY (email)
```

when email can change.

Prefer:

```sql
id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
email TEXT NOT NULL UNIQUE
```

Identity and business uniqueness are separate concepts.

---

# 47. Common Mistake: Inconsistent ID Types

Avoid:

```text
users.id       INTEGER
orders.user_id BIGINT
```

Prefer:

```text
users.id       BIGINT
orders.user_id BIGINT
```

Use a consistent ID strategy across related tables.

---

# 48. Common Mistake: Ignoring Existing Data During Migration

A migration may look correct:

```sql
ALTER TABLE orders
ADD CONSTRAINT fk_orders_user
FOREIGN KEY (user_id)
REFERENCES users(id);
```

But existing invalid rows can make it fail.

Always consider:

```text
existing rows
      ↓
data cleanup
      ↓
constraint
```

---

# 49. Production Design Checklist

For every important relationship, ask:

### Identity

- What uniquely identifies the row?
- Is the primary key stable?
- Is the ID type consistent?

### Relationship

- Which table is the parent?
- Which table stores the foreign key?
- Is the relationship required or optional?

### Integrity

- Should the FK be `NOT NULL`?
- Should the database enforce uniqueness?
- Can invalid references ever exist?

### Deletion

- `RESTRICT`?
- `CASCADE`?
- `SET NULL`?
- Soft delete instead?

### Performance

- Is the FK column queried frequently?
- Does it need an index?
- What do the real queries look like?

### Migration

- Does existing data satisfy the new constraint?
- Can the migration run safely in production?
- Is backfill required?

---

# 50. Practical Mental Model

Think of a schema like this:

```text
                PRIMARY KEY
                    ↓
              identifies row
                    |
                    |
users.id  <---------+--------- orders.user_id
   ↑                              ↑
 parent                         foreign key
                                  |
                                  ↓
                         references parent
```

Then add the responsibilities:

```text
PRIMARY KEY
    = who is this row?

FOREIGN KEY
    = which parent does this row reference?

NOT NULL
    = is the relationship required?

UNIQUE
    = can this relationship/value repeat?

ON DELETE
    = what happens when the parent disappears?

INDEX
    = how efficiently can we find related rows?
```

---

# 51. Production Takeaways

1. Every important table should have a clear primary key.
2. Prefer stable identifiers over mutable business fields.
3. Use foreign keys to enforce relationships inside the database.
4. Use `NOT NULL` when a relationship is mandatory.
5. Treat `ON DELETE` as a business decision, not a convenience setting.
6. Use `CASCADE` only when child data should genuinely disappear with the parent.
7. Use `SET NULL` only for relationships that are truly optional.
8. Use `UNIQUE` to enforce one-to-one relationships when appropriate.
9. Consider indexes on foreign-key columns based on query patterns.
10. Database constraints protect against concurrency and multiple writers.
11. Plan migrations around existing data.
12. Keep identity, integrity, and performance concerns conceptually separate.

The core mental model:

> **Primary keys identify. Foreign keys relate. Constraints protect. Indexes accelerate.**
