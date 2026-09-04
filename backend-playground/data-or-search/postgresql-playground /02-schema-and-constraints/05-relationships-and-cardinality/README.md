# Relationships and Cardinality

Production-first PostgreSQL material for modeling how records relate to each other.

---

## 1. What This Material Covers

This chapter focuses on:

- relationships between tables
- parent and child tables
- one-to-one
- one-to-many
- many-to-many
- foreign keys
- unique foreign keys
- junction tables
- relationship optionality
- cardinality
- composite relationships
- choosing the correct schema shape
- common production mistakes

Core idea:

> **Cardinality describes how many records on one side can be related to records on the other side.**

---

# 2. Why Relationships Matter

Real applications rarely have isolated tables.

Typical systems contain:

```text
users
orders
products
payments
organizations
roles
permissions
```

These records depend on each other.

For example:

```text
user
  ↓
orders
```

or:

```text
orders
  ↓
order_items
  ↓
products
```

The schema needs to represent these relationships correctly.

---

# 3. Parent and Child

A useful mental model:

```text
parent
  ↓
child
```

The child normally stores a foreign key to the parent.

Example:

```text
users
  id
   ↑
   |
orders
  user_id
```

SQL:

```sql
CREATE TABLE orders (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    user_id BIGINT NOT NULL,

    FOREIGN KEY (user_id)
        REFERENCES users(id)
);
```

---

# 4. Cardinality

Cardinality answers:

> How many records can be related?

The most common relationship types are:

```text
one-to-one
one-to-many
many-to-many
```

There is also an important distinction between:

```text
required
optional
```

These dimensions should be considered separately.

---

# 5. One-to-Many

One-to-many is probably the most common relationship in backend systems.

Example:

```text
one customer
     ↓
many orders
```

Schema:

```sql
CREATE TABLE customers (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    name TEXT NOT NULL
);

CREATE TABLE orders (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    customer_id BIGINT NOT NULL,

    FOREIGN KEY (customer_id)
        REFERENCES customers(id)
);
```

One customer can have many orders.

Each order belongs to one customer.

---

# 6. Why the Foreign Key Goes on the Many Side

Suppose:

```text
customer 1
   ↓
order 1
order 2
order 3
```

The `orders` table can store:

```text
id | customer_id
---+------------
1  | 1
2  | 1
3  | 1
```

The parent does not need a column such as:

```text
order_ids = [1,2,3]
```

Relational databases represent this naturally through rows in the child table.

---

# 7. One-to-Many Example

```sql
CREATE TABLE users (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    name TEXT NOT NULL
);

CREATE TABLE posts (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    user_id BIGINT NOT NULL,
    title TEXT NOT NULL,

    CONSTRAINT fk_posts_user
        FOREIGN KEY (user_id)
        REFERENCES users(id)
);
```

Data:

```text
users

id | name
---+------
1  | Alice
2  | Bob
```

```text
posts

id | user_id | title
---+---------+--------
10 | 1       | Hello
11 | 1       | PostgreSQL
12 | 2       | Go
```

Alice has two posts.

Bob has one.

---

# 8. Required vs Optional One-to-Many

Consider:

```text
post → author
```

If every post must have an author:

```sql
author_id BIGINT NOT NULL
```

If a post may survive without an author:

```sql
author_id BIGINT
```

with:

```sql
FOREIGN KEY (author_id)
REFERENCES users(id)
ON DELETE SET NULL
```

So cardinality is not the same as optionality.

You need to answer both:

```text
How many?
+
Is it required?
```

---

# 9. One-to-One

One-to-one means:

```text
one user
   ↓
at most one profile
```

A normal foreign key alone does not enforce this.

This:

```sql
user_id BIGINT NOT NULL
REFERENCES users(id)
```

still allows:

```text
user_id = 1
user_id = 1
user_id = 1
```

So it is one-to-many.

---

# 10. Enforcing One-to-One

Add `UNIQUE`:

```sql
CREATE TABLE user_profiles (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    user_id BIGINT NOT NULL UNIQUE,

    FOREIGN KEY (user_id)
        REFERENCES users(id)
);
```

Now:

```text
users
  1
  ↓
profile
```

Only one profile can reference user `1`.

The key rule is:

> **A unique foreign key enforces at-most-one child row per parent.**

---

# 11. One-to-One Is Often Really One-to-Zero-or-One

Many systems have:

```text
user
  ↓
0 or 1 profile
```

not:

```text
user
  ↓
exactly 1 profile
```

A unique foreign key naturally enforces:

```text
at most one
```

The existence of the profile row determines whether the optional relationship exists.

---

# 12. When One-to-One Is Useful

Examples:

```text
users → user_profiles
users → user_preferences
users → employee_details
```

It can also separate rarely accessed or sensitive fields from a main table.

For example:

```text
users
  ↓
authentication_details
```

The relationship shape should come from the domain, not from an assumption that every concept deserves a separate table.

---

# 13. One-to-One vs Same Table

Sometimes developers split a table unnecessarily.

Instead of:

```text
users
  +
user_profiles
```

a single table may be simpler:

```sql
users (
    id,
    name,
    avatar_url,
    bio
)
```

A separate table can make sense when:

- fields are optional
- fields are rarely accessed
- lifecycle differs
- ownership differs
- security boundaries matter
- the table would otherwise become unwieldy

Do not split tables merely because one-to-one is possible.

---

# 14. Many-to-Many

Many-to-many means:

```text
many users
    ↕
many roles
```

A user can have many roles.

A role can belong to many users.

Relational databases normally model this with a junction table.

---

# 15. Junction Table

Example:

```sql
CREATE TABLE users (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY
);

CREATE TABLE roles (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY
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

Relationship:

```text
users
   ↕
user_roles
   ↕
roles
```

---

# 16. Why a Junction Table Is Needed

Avoid storing:

```text
user.role_ids = [1,2,3]
```

in a normal relational design when the relationship needs querying, constraints, or additional data.

The junction table gives each relationship its own row:

```text
user_id | role_id
--------+--------
1       | 1
1       | 2
2       | 1
```

Now SQL can efficiently query the relationship.

---

# 17. Preventing Duplicate Relationships

A junction table should usually prevent duplicate pairs.

Use:

```sql
PRIMARY KEY (user_id, role_id)
```

or:

```sql
UNIQUE (user_id, role_id)
```

Then:

```text
user 1 → role 2
```

can exist only once.

This also makes repeated relationship creation safely idempotent when combined with:

```sql
ON CONFLICT DO NOTHING
```

---

# 18. Junction Tables Can Have Their Own Data

A relationship may have attributes.

Example:

```text
user
  ↕
membership
  ↕
organization
```

The membership may contain:

```text
joined_at
role
status
invited_by
```

Schema:

```sql
CREATE TABLE memberships (
    user_id BIGINT NOT NULL,
    organization_id BIGINT NOT NULL,
    role TEXT NOT NULL,
    joined_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    PRIMARY KEY (user_id, organization_id),

    FOREIGN KEY (user_id)
        REFERENCES users(id),

    FOREIGN KEY (organization_id)
        REFERENCES organizations(id)
);
```

The junction table is now a real domain entity.

---

# 19. When a Junction Table Becomes an Entity

Compare:

```text
user ↔ role
```

with:

```text
user ↔ organization
       |
       +-- role
       +-- joined_at
       +-- status
```

The second relationship contains meaningful data.

That often means the junction table deserves its own domain model in application code.

For example:

```text
Membership
```

rather than treating it as an invisible join implementation detail.

---

# 20. Many-to-Many Query

Find all roles for a user:

```sql
SELECT r.id, r.name
FROM roles r
JOIN user_roles ur
    ON ur.role_id = r.id
WHERE ur.user_id = $1;
```

Find all users with a role:

```sql
SELECT u.id, u.name
FROM users u
JOIN user_roles ur
    ON ur.user_id = u.id
WHERE ur.role_id = $1;
```

The junction table makes both directions natural.

---

# 21. Indexing Junction Tables

The primary key:

```sql
PRIMARY KEY (user_id, role_id)
```

creates an index beginning with:

```text
user_id
```

That helps queries filtering by `user_id`.

But the reverse query:

```sql
WHERE role_id = $1
```

may benefit from another index:

```sql
CREATE INDEX idx_user_roles_role_id
ON user_roles(role_id);
```

For relationship tables, think about both traversal directions.

---

# 22. Order of Composite Index Columns

Suppose:

```sql
CREATE INDEX idx_memberships_org_user
ON memberships(organization_id, user_id);
```

This is naturally useful for queries starting with:

```sql
WHERE organization_id = $1
```

The column order matters.

For a composite key:

```text
(user_id, organization_id)
```

and:

```text
(organization_id, user_id)
```

are not interchangeable for every query pattern.

Design indexes around access patterns.

---

# 23. Relationship Optionality

A relationship can be:

```text
required
optional
```

Example:

```text
order → customer
```

Usually required:

```sql
customer_id BIGINT NOT NULL
```

Example:

```text
ticket → assigned_agent
```

Potentially optional:

```sql
agent_id BIGINT NULL
```

Optionality is represented by whether the FK can be `NULL`.

---

# 24. Zero-to-Many

One-to-many does not mean every parent must have at least one child.

Usually:

```text
customer
   ↓
0..N orders
```

A new customer may have no orders yet.

The database does not need a special setting for this.

The child rows simply do not exist.

---

# 25. One-to-Zero-or-One

A unique nullable relationship can represent:

```text
0..1
```

Example:

```sql
CREATE TABLE user_profiles (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    user_id BIGINT NOT NULL UNIQUE
        REFERENCES users(id)
);
```

A user can have:

```text
zero profiles
```

or:

```text
one profile
```

but never two.

---

# 26. Exactly One Is Different

A database schema with a unique foreign key usually guarantees:

```text
0..1
```

not:

```text
exactly 1
```

If every user must have a profile, the application may need to create both within the same transaction.

For example:

```text
BEGIN
  create user
  create profile
COMMIT
```

The exact design depends on the lifecycle.

Do not confuse:

```text
at most one
```

with:

```text
exactly one
```

---

# 27. Foreign Key Direction

For:

```text
customer → orders
```

the foreign key normally belongs on:

```text
orders.customer_id
```

Think:

```text
many side stores reference to one side
```

This rule makes one-to-many modeling much easier.

---

# 28. Self-Referencing Relationship

A table can reference itself.

Example:

```text
employees
   |
   +---- manager_id → employees.id
```

Schema:

```sql
CREATE TABLE employees (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    name TEXT NOT NULL,
    manager_id BIGINT,

    FOREIGN KEY (manager_id)
        REFERENCES employees(id)
);
```

This represents:

```text
employee
   ↓
manager
   ↓
employee
```

It is useful for hierarchical data.

---

# 29. Self-Referencing Example

Data:

```text
id | name    | manager_id
---+---------+-----------
1  | Alice   | NULL
2  | Bob     | 1
3  | Charlie | 1
```

Meaning:

```text
Alice
├── Bob
└── Charlie
```

The root row has:

```text
manager_id = NULL
```

while child rows reference another employee.

---

# 30. Self-Referencing Production Considerations

Self-references are useful for:

- employee hierarchies
- categories
- folders
- comments/replies
- organizational structures

But they introduce application-level questions:

- Can a row reference itself?
- Can cycles exist?
- How deep can the hierarchy become?
- How should deletion work?

A simple foreign key prevents references to nonexistent rows, but it does not automatically enforce every hierarchy rule.

---

# 31. Categories Example

```sql
CREATE TABLE categories (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    parent_id BIGINT,

    CONSTRAINT fk_categories_parent
        FOREIGN KEY (parent_id)
        REFERENCES categories(id)
);
```

Data:

```text
Electronics
├── Phones
│   ├── Android
│   └── iPhone
└── Laptops
```

This is a self-referencing one-to-many relationship.

---

# 32. Multiple Foreign Keys to the Same Table

A table can reference the same parent table multiple times.

Example:

```sql
CREATE TABLE orders (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,

    created_by BIGINT NOT NULL,
    approved_by BIGINT,

    FOREIGN KEY (created_by)
        REFERENCES users(id),

    FOREIGN KEY (approved_by)
        REFERENCES users(id)
);
```

Both point to:

```text
users.id
```

but represent different relationships.

Name columns based on their business meaning.

---

# 33. Multiple Relationships Need Clear Names

Avoid vague names:

```text
user_id
```

when the table actually has:

```text
created_by
approved_by
assigned_to
cancelled_by
```

Clear relationship names improve:

- queries
- API models
- debugging
- migrations
- code review

The foreign key tells you what is valid.

The column name tells you what the relationship means.

---

# 34. Relationship vs Embedded Data

Suppose an order contains:

```text
customer_id
```

That is a relationship.

Suppose the order also stores:

```text
customer_name
```

That is duplicated data.

Sometimes duplication is intentional, such as preserving a historical snapshot.

For example:

```text
invoice.customer_name
```

may represent the name at invoice creation time rather than the customer's current name.

Do not assume every repeated value is an accidental normalization problem.

Understand the business meaning.

---

# 35. Historical Relationships

Consider:

```text
orders.customer_id
```

If the customer is renamed, old orders still point to the same customer.

That is usually desirable.

But a historical document may also need a snapshot:

```text
billing_name
billing_address
```

The key question is:

> Do we need the current related data, or the value at the time of the event?

That decision affects schema design.

---

# 36. Deletion and Relationships

When designing a relationship, decide what happens when the parent is deleted.

Possible behaviors:

```text
RESTRICT
CASCADE
SET NULL
```

Example:

```text
user → session
```

may reasonably use:

```sql
ON DELETE CASCADE
```

Example:

```text
customer → invoice
```

may need:

```sql
ON DELETE RESTRICT
```

Relationship modeling includes lifecycle behavior, not just cardinality.

---

# 37. Relationship and Soft Delete

If parent records are soft-deleted:

```sql
deleted_at TIMESTAMPTZ
```

the foreign key can continue referencing the row.

This preserves the relationship while allowing the application to hide deleted records.

Be careful with queries:

```sql
JOIN users u
    ON u.id = orders.user_id
```

may still return a soft-deleted user.

If only active users should appear, the query needs an explicit condition.

---

# 38. Tenant Relationships

In multi-tenant systems, relationships often need tenant boundaries.

Example:

```text
organizations
users
projects
```

A project may belong to an organization:

```sql
organization_id BIGINT NOT NULL
    REFERENCES organizations(id)
```

But application authorization must also ensure that related records belong to the same tenant.

A foreign key to each table does not automatically prove:

```text
project.organization_id
=
user.organization_id
```

if both are independently stored.

Cross-column or composite designs may be required when the database itself must enforce tenant consistency.

---

# 39. Composite Relationship for Tenant Safety

One database-level approach is to make the tenant part of the referenced key.

Conceptually:

```text
organizations
    ↓
(organization_id, user_id)
```

and:

```text
projects
    ↓
(organization_id, user_id)
```

Then a composite foreign key can enforce that both values belong to the same tenant context.

This is more advanced and should be introduced only when the integrity requirement justifies the added schema complexity.

---

# 40. Relationship Queries

For a one-to-many relationship:

```sql
SELECT o.id, o.total
FROM orders o
WHERE o.customer_id = $1;
```

For parent plus children:

```sql
SELECT
    c.id,
    c.name,
    o.id AS order_id,
    o.total
FROM customers c
LEFT JOIN orders o
    ON o.customer_id = c.id
WHERE c.id = $1;
```

Use `LEFT JOIN` when the parent should still appear even when it has zero children.

---

# 41. LEFT JOIN and Zero Children

Suppose:

```text
customer 1 → 3 orders
customer 2 → 0 orders
```

Using:

```sql
LEFT JOIN orders
```

returns customer `2` with `NULL` order columns.

Using:

```sql
INNER JOIN orders
```

would remove customer `2` from the result.

The join type is part of relationship reasoning.

---

# 42. EXISTS for Relationship Checks

If you only need to know whether a relationship exists, `EXISTS` is often clearer.

Example:

```sql
SELECT c.id, c.name
FROM customers c
WHERE EXISTS (
    SELECT 1
    FROM orders o
    WHERE o.customer_id = c.id
);
```

This asks:

> Which customers have at least one order?

You do not need to retrieve every order row.

---

# 43. Avoiding Duplicate Parents

A common mistake:

```sql
SELECT c.*
FROM customers c
JOIN orders o
    ON o.customer_id = c.id;
```

If a customer has five orders, that customer appears five times.

If the requirement is:

```text
customers who have orders
```

prefer:

```sql
SELECT c.*
FROM customers c
WHERE EXISTS (
    SELECT 1
    FROM orders o
    WHERE o.customer_id = c.id
);
```

or use appropriate aggregation/distinct logic when needed.

Think about result grain.

---

# 44. Relationship Counts

To count children:

```sql
SELECT
    c.id,
    c.name,
    COUNT(o.id) AS order_count
FROM customers c
LEFT JOIN orders o
    ON o.customer_id = c.id
GROUP BY c.id, c.name;
```

Important:

```sql
COUNT(o.id)
```

returns `0` for customers with no matching order rows.

This works because the joined child ID is `NULL` when there is no child.

---

# 45. Relationship Table Grain

A useful question:

> What does one row in this table represent?

Examples:

```text
users
→ one user

orders
→ one order

order_items
→ one item inside one order

user_roles
→ one user-role relationship
```

Correct relationship modeling starts with correct row grain.

---

# 46. Relationship Modeling Workflow

When designing a new feature:

### Step 1

Identify entities.

```text
User
Order
Product
```

### Step 2

Define what one row represents.

```text
one order
one product
```

### Step 3

Ask how entities relate.

```text
User → Order
Order → Product
```

### Step 4

Determine cardinality.

```text
User 1 → N Orders
Order 1 → N OrderItems
Product 1 → N OrderItems
```

### Step 5

Choose the FK location.

```text
orders.user_id
order_items.order_id
order_items.product_id
```

### Step 6

Add constraints and indexes.

---

# 47. Example: E-Commerce Relationship Model

```text
customers
   |
   | 1:N
   ↓
orders
   |
   | 1:N
   ↓
order_items
   ↑
   | N:1
   |
products
```

Tables:

```text
customers
orders
order_items
products
```

Foreign keys:

```text
orders.customer_id
order_items.order_id
order_items.product_id
```

This is a common production schema shape.

---

# 48. Example: SaaS Relationship Model

```text
organizations
   |
   +---- users
   |
   +---- projects
            |
            +---- tasks
```

Potential relationships:

```text
organization 1:N users
organization 1:N projects
project      1:N tasks
```

For users belonging to multiple organizations:

```text
users
  ↕
memberships
  ↕
organizations
```

The relationship determines whether you need a direct FK or a junction table.

---

# 49. Example: Permissions

A common authorization relationship:

```text
users
  ↕
user_roles
  ↕
roles
  ↕
role_permissions
  ↕
permissions
```

This is multiple many-to-many relationships.

Each junction table can enforce:

```sql
PRIMARY KEY (left_id, right_id)
```

to prevent duplicate relationships.

---

# 50. Relationship Integrity Checklist

For every relationship, ask:

### Entity

- What are the two entities?
- What does one row represent?

### Cardinality

- one-to-one?
- one-to-many?
- many-to-many?

### Optionality

- required?
- optional?
- can the parent have zero children?

### Foreign key

- Which table stores the FK?
- Is the FK `NOT NULL`?

### Uniqueness

- Does the FK need `UNIQUE`?
- Does the relationship need composite uniqueness?

### Lifecycle

- What happens when the parent is deleted?
- `CASCADE`, `RESTRICT`, `SET NULL`, or soft delete?

### Performance

- How will the relationship be queried?
- Does the FK need an index?
- Does a junction table need indexes in both directions?

---

# 51. Common Mistake: Storing Lists of IDs

Avoid:

```text
role_ids = [1,2,3]
```

in a normal relational column when those IDs represent relationships that need constraints and queries.

Prefer:

```text
user_roles
```

with:

```text
user_id
role_id
```

A junction table provides:

- foreign keys
- uniqueness
- indexes
- efficient joins
- relationship-level metadata

---

# 52. Common Mistake: Missing UNIQUE on One-to-One

This:

```sql
user_id BIGINT REFERENCES users(id)
```

does not create one-to-one.

It creates:

```text
one user → many rows
```

Add:

```sql
UNIQUE (user_id)
```

if the domain requires at most one child.

---

# 53. Common Mistake: Wrong FK Side

For:

```text
customer 1 → N orders
```

do not store an array of order IDs in `customers`.

Use:

```text
orders.customer_id
```

The many side stores the reference to the one side.

---

# 54. Common Mistake: Confusing Cardinality with Optionality

These are different:

```text
1:N
```

describes quantity.

```text
required/optional
```

describes whether the relationship must exist.

For example:

```text
customer → orders
```

can be:

```text
1 : 0..N
```

A customer can have zero or many orders.

---

# 55. Common Mistake: Ignoring Delete Semantics

A relationship is not fully designed until you know what happens on deletion.

Ask:

```text
What happens to children?
```

Do not add:

```sql
ON DELETE CASCADE
```

everywhere without considering data loss.

---

# 56. Common Mistake: Missing Junction Uniqueness

Bad:

```text
user_id | role_id
--------+--------
1       | 2
1       | 2
```

Better:

```sql
PRIMARY KEY (user_id, role_id)
```

This turns:

```text
duplicate relationship
```

into:

```text
database constraint violation
```

---

# 57. Common Mistake: Ignoring Query Direction

A junction table may be queried from either side.

Example:

```text
user → roles
role → users
```

Design indexes for the actual access patterns.

Do not assume one composite index automatically optimizes both directions.

---

# 58. Common Mistake: Over-Normalizing Simple Data

Not every field needs a separate entity.

If:

```text
users.avatar_url
```

is simple data owned by the user, a separate:

```text
user_avatars
```

table may add unnecessary complexity.

Create relationships because the domain needs them, not because normalization can be taken to an extreme.

---

# 59. Common Mistake: Using JSON for Relational Relationships

A JSON field can store:

```json
{
  "role_ids": [1, 2, 3]
}
```

but this weakens normal relational guarantees.

You lose the straightforward ability to enforce:

```text
role_id must exist
duplicate relationship forbidden
relationship indexed naturally
```

If the data is fundamentally relational and frequently queried, model it relationally.

---

# 60. Production Mental Model

Think of relationships as:

```text
ONE
  ↓
owns/has
  ↓
MANY
```

For one-to-many:

```text
parent
   ↑
   |
child.fk
```

For one-to-one:

```text
child.fk
   +
UNIQUE
```

For many-to-many:

```text
A
↕
junction table
↕
B
```

Then add:

```text
NOT NULL
→ required relationship

FOREIGN KEY
→ valid reference

UNIQUE
→ cardinality protection

INDEX
→ efficient traversal

ON DELETE
→ lifecycle behavior
```

---

# 61. Production Takeaways

1. Start relationship design by identifying the row grain of each table.
2. One-to-many is usually modeled with a foreign key on the many side.
3. A foreign key alone does not create one-to-one.
4. Use a unique foreign key for at-most-one child rows.
5. Many-to-many relationships normally use junction tables.
6. Junction tables should usually prevent duplicate pairs.
7. Junction tables can become full domain entities when the relationship has its own data.
8. Required relationships use `NOT NULL`.
9. Optional relationships can use nullable foreign keys.
10. `LEFT JOIN` is important when parents with zero children must remain in results.
11. `EXISTS` is useful when you only need to test relationship existence.
12. Foreign-key indexes should follow actual query directions.
13. Delete behavior is part of relationship design.
14. Self-references are useful for hierarchical data but need additional business rules when cycles matter.
15. Do not use JSON arrays for relationships that need normal relational integrity and querying.
16. Avoid over-normalizing simple data.
17. Model the relationship according to the business domain, not just what SQL allows.

Core mental model:

> **Cardinality tells you how many.  
> Foreign keys tell you what relates to what.  
> UNIQUE enforces relationship limits.  
> NOT NULL defines requiredness.  
> Indexes make traversal fast.**
