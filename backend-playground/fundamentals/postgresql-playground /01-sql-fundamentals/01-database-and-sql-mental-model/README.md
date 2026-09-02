# 01 — Database and SQL Mental Model

## Goal

Build the minimum mental model needed to work with PostgreSQL effectively in production.

This material is intentionally **concise and production-first**.

---

## 1. Database vs DBMS

A **database** stores persistent data and its structure.

A **DBMS (Database Management System)** is the software that manages that data.

Examples:

- PostgreSQL
- MySQL
- MariaDB
- SQL Server
- Oracle

Think:

```text
Database = data
DBMS     = software managing the data
```

---

## 2. PostgreSQL vs SQL

**SQL** is the language used to communicate with relational databases.

**PostgreSQL** is a relational database management system that implements SQL and provides PostgreSQL-specific features.

Example SQL:

```sql
SELECT id, name
FROM users
WHERE status = 'active';
```

Mental model:

```text
SQL
↓
describes what you want

PostgreSQL
↓
decides how to execute it
```

---

## 3. Database Structure

A useful PostgreSQL hierarchy is:

```text
Database
  ↓
Schema
  ↓
Table
  ↓
Rows + Columns
```

Example:

```text
crm
└── public
    ├── users
    ├── orders
    └── products
```

A schema is mainly a namespace for database objects.

For many applications, the default `public` schema is sufficient.

---

## 4. Table, Row, Column

Example:

```text
users

id | name  | email
---+-------+-------------------
1  | Alice | alice@example.com
2  | Bob   | bob@example.com
```

- **Table** → collection of related records
- **Row** → one record
- **Column** → one attribute

A column also has a type:

```sql
id BIGINT
name TEXT
email TEXT
created_at TIMESTAMPTZ
```

### Production point

Choose types that represent the data correctly. Do not store everything as `TEXT` just because it is convenient.

---

## 5. Primary Key

A primary key uniquely identifies a row.

```sql
CREATE TABLE users (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    name TEXT NOT NULL
);
```

A primary key is:

- unique
- not NULL

It is commonly used for:

- lookups
- updates
- deletes
- foreign keys
- joins

### Important

A primary key does **not** automatically mean "newest row."

If chronological ordering matters, use a timestamp such as:

```sql
ORDER BY created_at DESC
```

---

## 6. Foreign Key

A foreign key represents a relationship between tables.

```sql
CREATE TABLE orders (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id),
    total NUMERIC(12,2) NOT NULL
);
```

Relationship:

```text
users.id
   ↑
orders.user_id
```

An order belongs to a user.

### Production point

Foreign keys can prevent invalid references and protect relational integrity.

---

## 7. Constraints

Constraints tell the database which states are invalid.

Common constraints:

```text
PRIMARY KEY
FOREIGN KEY
NOT NULL
UNIQUE
CHECK
```

Example:

```sql
CREATE TABLE products (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    sku TEXT NOT NULL UNIQUE,
    price NUMERIC(12,2) NOT NULL CHECK (price >= 0)
);
```

The database now enforces:

```text
id    → unique identity
sku   → required + unique
price → required + non-negative
```

### Production point

Important persistent invariants should not exist only in application code.

---

## 8. Application Validation vs Database Constraints

You normally want both.

### Application validation

Useful for:

- friendly errors
- input validation
- request-level rules
- early feedback

### Database constraints

Useful for:

- data integrity
- concurrency-safe enforcement
- protecting every writer

Example:

```text
API:
"Is this email formatted correctly?"

Database:
"Is this email unique?"
```

The database should be the final authority for important data invariants.

---

## 9. Relationships

The common relational relationships are:

```text
one-to-one
one-to-many
many-to-many
```

### One-to-many

```text
one user
    ↓
many orders
```

Usually:

```text
users
  id

orders
  id
  user_id
```

### Many-to-many

Usually uses a junction table:

```text
users
roles
user_roles
```

Understanding cardinality is essential when writing joins.

---

## 10. Why Joins Can Multiply Rows

Suppose Alice has three orders.

A join can produce:

```text
Alice | Order 1
Alice | Order 2
Alice | Order 3
```

This is not duplicated user data. It represents three relationships.

### Production habit

Before joining tables, ask:

> How many rows can each side contribute?

This prevents many duplicate-result bugs.

---

## 11. Normalization

Normalization is about structuring relational data to reduce unnecessary duplication and update problems.

Instead of:

```text
orders
--------------------------------
order_id | user_name | user_email
```

prefer:

```text
users
----------------
id | name | email

orders
----------------
id | user_id
```

Now user information has one natural home.

### Production mental model

> Store each fact in an appropriate place and represent relationships explicitly.

Normalization does not mean data can never be duplicated. Intentional denormalization can be useful for performance, reporting, search, or historical snapshots.

---

## 12. Transactions

A transaction groups related database operations into one logical unit.

Example:

```text
create order
+
decrease inventory
+
create payment record
```

Conceptually:

```sql
BEGIN;

-- related operations

COMMIT;
```

If something fails:

```sql
ROLLBACK;
```

The goal is to keep operations that must succeed together atomic.

### Production question

Ask:

> Which operations must succeed or fail together?

That usually defines the transaction boundary.

---

## 13. Keep Transactions Focused

Transactions should generally be as short as practical.

Avoid unnecessarily doing:

```text
BEGIN
↓
database update
↓
external API call
↓
wait
↓
another API call
↓
COMMIT
```

Long transactions can increase:

- lock contention
- resource usage
- latency

Keep the atomic database work focused.

---

## 14. Indexes

An index is an additional data structure that can help PostgreSQL find or order data efficiently.

Example:

```sql
CREATE INDEX users_email_idx
ON users(email);
```

For:

```sql
SELECT id, name
FROM users
WHERE email = $1;
```

an index on `email` may provide an efficient access path.

Think:

```text
without index
→ search more data

with appropriate index
→ find relevant data more efficiently
```

### Production point

Indexes should support real query patterns.

---

## 15. Indexes Are Not Free

Indexes consume:

- storage
- write work
- maintenance

Too many indexes can make writes more expensive.

Do not create an index on every column.

Think about:

```text
query pattern
data size
filtering
ordering
write frequency
```

Then measure.

---

## 16. Query Planner

PostgreSQL roughly follows:

```text
SQL
 ↓
parse
 ↓
plan
 ↓
execute
```

The planner chooses an execution strategy.

Examples include:

```text
Sequential Scan
Index Scan
Bitmap Scan
Nested Loop
Hash Join
Merge Join
```

You do not need to memorize these yet.

The important idea is:

> SQL describes the operation; PostgreSQL chooses the execution strategy.

---

## 17. EXPLAIN

PostgreSQL provides:

```sql
EXPLAIN
```

to inspect a query plan.

Example:

```sql
EXPLAIN
SELECT id, name
FROM users
WHERE email = 'alice@example.com';
```

This connects SQL with actual database execution.

Later, `EXPLAIN ANALYZE` can be used to compare estimates with actual execution.

Remember:

> `EXPLAIN ANALYZE` executes the query, so be careful with modifying statements.

---

## 18. Application ↔ Database

A typical Go backend looks like:

```text
HTTP request
     ↓
handler
     ↓
service
     ↓
repository
     ↓
database driver
     ↓
connection pool
     ↓
PostgreSQL
```

The application sends SQL and parameters.

PostgreSQL returns:

```text
rows
or
error
```

The application then maps the result to its domain/API response.

---

## 19. Connection Pool

Production applications normally reuse database connections through a pool.

```text
Application
     ↓
Connection Pool
   ├── connection
   ├── connection
   ├── connection
   └── connection
     ↓
PostgreSQL
```

A pool helps:

- reuse connections
- control concurrent database access
- avoid creating a connection for every request

### Production point

More connections are not automatically better.

Too many application instances with oversized pools can overwhelm PostgreSQL.

---

## 20. Parameterized Queries

Application values should be passed as parameters.

Good:

```sql
SELECT id, name
FROM users
WHERE email = $1;
```

Avoid constructing SQL by concatenating raw user input.

Parameterized queries help:

- prevent SQL injection
- handle values correctly
- separate SQL structure from data

Production rule:

> Keep SQL structure separate from user-controlled values.

---

## 21. ORM vs SQL

An ORM can abstract database operations.

For example:

```text
User.find(...)
```

may eventually become SQL.

But PostgreSQL still executes SQL.

ORMs do not remove problems such as:

- N+1 queries
- bad joins
- missing indexes
- huge result sets
- slow pagination
- poor transaction boundaries

Therefore:

> Learn SQL even when the application uses an ORM.

The ORM should be an abstraction you understand, not a black box.

---

## 22. N+1 Queries

Suppose the application loads:

```text
100 users
```

and then performs one query for each user's orders.

That can become:

```text
1 query
+
100 queries
=
101 queries
```

The important lesson is:

> Know how many database queries your application actually executes.

This matters whether you use raw SQL or an ORM.

---

## 23. Explicit Columns

This is convenient during exploration:

```sql
SELECT *
FROM users;
```

For production queries, explicit columns are often preferable:

```sql
SELECT id, name, email
FROM users;
```

Benefits:

- clearer intent
- less data transferred
- less application memory
- less accidental exposure

---

## 24. Result Size

A query can be correct but operationally expensive.

This:

```sql
SELECT *
FROM orders;
```

could return millions of rows.

Production APIs commonly use:

```text
filters
LIMIT
pagination
specific columns
```

Think about the amount of data the query can return, not only whether the SQL is logically correct.

---

## 25. NULL

`NULL` represents an absent/unknown value.

It is different from:

```text
0
''
FALSE
```

For example:

```text
deleted_at = NULL
```

may represent:

```text
not deleted
```

To test for NULL:

```sql
WHERE deleted_at IS NULL
```

not:

```sql
WHERE deleted_at = NULL
```

NULL has special comparison and boolean behavior. It will be covered in detail in the dedicated NULL material.

---

## 26. Migrations

Production schema changes should be versioned through migrations.

Examples:

```text
create table
add column
add constraint
create index
remove obsolete column
```

Conceptually:

```text
migration 001
    ↓
migration 002
    ↓
migration 003
    ↓
current schema
```

This keeps environments reproducible:

```text
local
development
staging
production
```

---

## 27. Schema Changes Are Deployment Changes

A database change can affect:

- application code
- queries
- indexes
- locks
- existing data
- deployment compatibility

Production migrations should therefore be designed together with application deployment.

For important systems, backward-compatible changes are often safer than making an old application immediately incompatible with the new schema.

---

## 28. Production Data Modeling Checklist

When creating a table, ask:

```text
What entity is this?

What is its primary key?

Which fields are required?

Which fields can be NULL?

Which values must be unique?

What relationships exist?

Which invariants need constraints?

Are the data types correct?

What timestamps are needed?

How will this data be queried?

Which indexes support those queries?
```

You do not need a complicated schema.

You need an intentional schema.

---

## 29. Production Query Checklist

Before shipping a query, ask:

```text
What rows should this return?

Is the filtering correct?

Is NULL handled correctly?

Does ordering matter?

Can joins multiply rows?

Is the result bounded?

Are only required columns selected?

Are values parameterized?

Does the query have an appropriate index?

What happens at production data size?
```

---

## 30. Common Mistakes

### Relying only on application validation

Use database constraints for important persistent invariants.

### Adding indexes everywhere

Indexes have storage and write costs.

### Assuming ORM means no SQL knowledge is needed

The database still executes SQL.

### Using `SELECT *` everywhere

Explicit columns are often clearer and safer.

### Ignoring relationship cardinality

Joins can multiply rows.

### Keeping transactions open too long

This can increase contention.

### Building SQL with string concatenation

Use parameters.

### Treating IDs as timestamps

Use actual timestamps for chronological requirements.

### Assuming a short query is fast

Execution cost depends on the data and plan.

### Ignoring production-scale data

A query that works on 1,000 rows may behave very differently on 10 million rows.

---

## 31. Production Mental Model

Think about database work in five steps:

```text
1. MODEL
   What data and relationships exist?

2. INTEGRITY
   What states are allowed?

3. QUERY
   What data do I need?

4. EXECUTION
   How will PostgreSQL execute it?

5. PRODUCTION
   What happens under concurrency and scale?
```

This is the core mental model for the rest of the SQL roadmap.

---

## 32. What to Remember

1. PostgreSQL is more than storage.
2. SQL describes what you want; PostgreSQL chooses how to execute it.
3. Primary keys identify rows.
4. Foreign keys represent relationships.
5. Constraints protect important invariants.
6. Application validation and database constraints complement each other.
7. Indexes support specific query patterns and have costs.
8. Transactions make related operations atomic.
9. ORMs do not replace SQL knowledge.
10. Production database work starts with correctness and uses measurement for performance.

---

## Next

The next material is:

```text
02-select-and-filtering
```

Focus:

```text
SELECT
WHERE
comparison operators
AND / OR / NOT
IN / NOT IN
BETWEEN
LIKE / ILIKE
parameterized filtering
production query patterns
common mistakes
```

The goal is not to memorize SQL syntax.

The goal is to become comfortable translating backend requirements into precise SQL.
