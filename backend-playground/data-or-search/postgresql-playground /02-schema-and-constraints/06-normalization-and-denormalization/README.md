# Normalization and Denormalization

Production-first PostgreSQL material for deciding when data should be separated, when duplication is acceptable, and how schema shape affects correctness and performance.

---

## 1. What This Material Covers

This chapter focuses on:

- normalization
- practical table separation
- duplicated data
- update anomalies
- one source of truth
- denormalization
- read performance
- historical snapshots
- reporting and aggregates
- transactional consistency
- deciding when duplication is justified
- production trade-offs

Core idea:

> **Normalize by default for correctness. Denormalize deliberately when a measurable requirement justifies duplicated data.**

---

# 2. What Is Normalization?

Normalization means structuring relational data so that each fact has an appropriate place and unnecessary duplication is reduced.

Example:

Bad design:

```text
orders

id | customer_name | customer_email | total
---+---------------+----------------+------
1  | Alice         | alice@x.com    | 100
2  | Alice         | alice@x.com    | 200
3  | Alice         | alice@x.com    | 300
```

Customer information is repeated.

A normalized design separates the entities:

```text
customers

id | name  | email
---+-------+--------------
1  | Alice | alice@x.com
```

```text
orders

id | customer_id | total
---+-------------+------
1  | 1           | 100
2  | 1           | 200
3  | 1           | 300
```

---

# 3. Why Normalization Matters in Production

Normalization is primarily about keeping data correct and maintainable.

Without it, the same fact may exist in many rows:

```text
Alice
alice@x.com
```

If Alice changes her email, many rows may need updating.

If one row is missed:

```text
alice@x.com
alice@example.com
```

the database now contains conflicting representations of the same fact.

Normalization gives you:

```text
one fact
   ↓
one authoritative location
```

---

# 4. Source of Truth

A useful production question is:

> Where is the authoritative value stored?

For example:

```text
customers.email
```

should usually be the source of truth for the customer's current email.

Other tables can reference:

```text
customer_id
```

instead of copying the current email everywhere.

This reduces synchronization problems.

---

# 5. Practical Normalization Rule

Do not memorize every normal-form definition first.

For backend schema design, start with:

> **Store each independently changing fact in one appropriate place.**

For example:

```text
customer
    ↓
customer.email

order
    ↓
order.customer_id
```

If the customer email changes, the order does not need to be updated merely because the current email changed.

---

# 6. Update Anomaly

Duplicated data can create update anomalies.

Example:

```text
orders

id | customer_name | customer_email
---+---------------+----------------
1  | Alice         | old@example.com
2  | Alice         | old@example.com
3  | Alice         | old@example.com
```

Alice changes her email.

If only two rows are updated:

```text
new@example.com
new@example.com
old@example.com
```

the database now contains conflicting information.

With normalization:

```text
customers.email
```

is updated once.

---

# 7. Insert Anomaly

Poor schema design can also make inserting data awkward.

Imagine customer data can only exist inside an order:

```text
orders
customer_name
customer_email
```

What if a customer signs up but has not placed an order?

You may have to create fake order data just to store the customer.

Separating:

```text
customers
orders
```

allows each entity to have its own lifecycle.

---

# 8. Delete Anomaly

Duplication can also cause accidental information loss.

Suppose:

```text
customer information
```

exists only in:

```text
orders
```

If the customer's last order is deleted, customer information may disappear too.

With:

```text
customers
orders
```

the customer can exist independently from orders.

---

# 9. Entity Lifecycle

A strong reason to separate tables is different lifecycles.

Example:

```text
customer
   ↓
exists before first order

order
   ↓
created later
```

These are different entities with different lifecycles.

If two concepts:

- have independent identity
- change independently
- can exist independently

they often deserve separate tables.

---

# 10. Normalization Is Not "One Table Per Field"

Over-normalization is also a problem.

Bad extreme:

```text
user_name
user_email
user_phone
```

each stored in separate tables without a real domain reason.

This creates unnecessary joins and complexity.

Normalization does not mean:

> Split everything.

It means:

> Structure data according to entities, dependencies, and ownership.

---

# 11. Normalization and Relationships

Consider:

```text
customers
   |
   +---- orders
```

The relationship is represented by:

```sql
orders.customer_id
```

Customer attributes stay in:

```text
customers
```

Order attributes stay in:

```text
orders
```

This gives each table a clear responsibility.

---

# 12. Functional Dependency Intuition

You do not need to memorize formal theory to use this concept.

Ask:

> If I know this column, what other value does it determine?

Example:

```text
customer_id → customer_email
```

If one customer ID identifies exactly one customer, then the current email belongs to the customer, not to every order.

Therefore:

```text
customers.customer_id
    ↓
customers.email
```

rather than:

```text
orders.customer_id
orders.customer_email
```

for the current email.

---

# 13. Repeating Groups

A common relational design problem is storing multiple values in one column.

Bad:

```text
user
role_ids = '1,2,3'
```

Better:

```text
users
user_roles
roles
```

The relationship becomes rows:

```text
user_id | role_id
--------+--------
1       | 1
1       | 2
1       | 3
```

This supports normal foreign keys, uniqueness, and indexes.

---

# 14. Normalization and JSON

JSONB is useful when the data is genuinely document-like or flexible.

But avoid using JSON just to hide relational structure.

Bad candidate:

```json
{
  "customer_id": 1,
  "product_ids": [10, 11, 12]
}
```

when those values represent relationships that need:

- foreign keys
- joins
- uniqueness
- filtering
- indexing

Use relational tables when the data is fundamentally relational.

---

# 15. Normalization and Current Data

Suppose:

```text
orders.customer_id
```

references:

```text
customers.id
```

and:

```text
customers.email
```

changes.

A query can retrieve the current email:

```sql
SELECT
    o.id,
    c.email
FROM orders o
JOIN customers c
    ON c.id = o.customer_id;
```

This is normalized current-state data.

The order does not need a duplicated current email.

---

# 16. But Historical Data Can Be Different

Sometimes you intentionally need the value at a specific point in time.

Example:

```text
invoice
```

should preserve:

```text
billing_name
billing_address
```

as they were when the invoice was issued.

If the customer later changes their address, the invoice should not suddenly display the new address.

That duplication is intentional.

It is not necessarily bad normalization.

---

# 17. Snapshot Data

A snapshot is a deliberate copy of data representing a historical state.

Example:

```sql
CREATE TABLE invoices (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    customer_id BIGINT NOT NULL,
    billing_name TEXT NOT NULL,
    billing_address TEXT NOT NULL,
    issued_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    FOREIGN KEY (customer_id)
        REFERENCES customers(id)
);
```

Here:

```text
customer_id
    ↓
current relationship

billing_name
billing_address
    ↓
historical snapshot
```

Both can legitimately coexist.

---

# 18. Current State vs Historical Fact

Ask:

> Do I need the current value or the value at the time of the event?

### Current value

Use a relationship:

```text
order.customer_id
```

then join to:

```text
customers
```

### Historical value

Store a snapshot:

```text
invoice.billing_name
invoice.billing_address
```

This distinction prevents many schema mistakes.

---

# 19. Denormalization

Denormalization intentionally duplicates or precomputes data.

Example:

```text
orders
    ↓
customer_name
```

even though:

```text
orders.customer_id
    ↓
customers.name
```

already exists.

Why would you do this?

Potential reasons include:

- expensive repeated joins
- read-heavy workloads
- reporting
- search/indexing requirements
- historical snapshots
- precomputed aggregates
- external system requirements

Denormalization should have a reason.

---

# 20. Denormalization Has a Cost

Duplicated data creates another responsibility:

> Keep the copies consistent when they are supposed to represent the same fact.

Example:

```text
customers.name
orders.customer_name
```

If they represent the same current value, every update needs synchronization.

That creates:

```text
correctness cost
maintenance cost
write complexity
```

Do not optimize reads by ignoring these costs.

---

# 21. Performance Is Not Automatically a Reason to Denormalize

Before duplicating data, check whether the real problem is:

```text
missing index
bad query
N+1 queries
unnecessary columns
bad join strategy
poor pagination
```

Often the correct fix is query/index optimization.

A normalized schema with good indexes can handle very large workloads.

---

# 22. Measure Before Denormalizing

A production decision should be based on evidence.

Useful tools include:

```sql
EXPLAIN
```

and:

```sql
EXPLAIN ANALYZE
```

Measure:

- query latency
- row counts
- execution plan
- frequency
- database load
- read/write ratio

Do not denormalize because:

> "Joins are slow."

Measure the actual query first.

---

# 23. JOINs Are Not Automatically Bad

A normalized schema often requires joins.

Example:

```sql
SELECT
    o.id,
    o.total,
    c.name
FROM orders o
JOIN customers c
    ON c.id = o.customer_id
WHERE o.id = $1;
```

This is a normal relational operation.

With appropriate indexes and a sensible query, a join can be efficient.

Avoid the false rule:

> Never use joins in production.

---

# 24. N+1 Is a Different Problem

Suppose the application does:

```text
query customers
   ↓
for each customer:
    query orders
```

For 100 customers:

```text
1 + 100 queries
```

That is an N+1 problem.

The solution may be:

```text
JOIN
```

or:

```text
batch query
```

not denormalization.

Always identify the actual bottleneck.

---

# 25. Denormalized Aggregates

A common legitimate use of denormalization is a precomputed counter.

Example:

```sql
CREATE TABLE posts (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    comment_count BIGINT NOT NULL DEFAULT 0
);
```

Instead of counting every comment on every request:

```sql
SELECT COUNT(*)
FROM comments
WHERE post_id = $1;
```

the application can read:

```text
posts.comment_count
```

This can improve read performance.

But now the counter must remain correct.

---

# 26. Counter Consistency

If you maintain:

```text
comment_count
```

you must define how it changes.

For example:

```text
create comment
    ↓
increment count

delete comment
    ↓
decrement count
```

If one path forgets to update the counter:

```text
actual comments = 10
comment_count = 9
```

The denormalized value is wrong.

This is the trade-off.

---

# 27. Denormalized Data as a Cache

Sometimes duplicated database data behaves like a cache.

Example:

```text
users
  ↓
current data

user_search_documents
  ↓
search-optimized representation
```

The second copy may be intentionally derived.

Once you think of it as a cache, you should ask:

```text
What is the source of truth?
How is it updated?
Can it be stale?
How is it rebuilt?
```

These questions are critical in production.

---

# 28. Source of Truth for Denormalized Data

Suppose:

```text
customers.name
orders.customer_name
```

If `customers.name` is authoritative:

```text
customers.name
      ↓
source of truth

orders.customer_name
      ↓
derived copy
```

Document this relationship.

Otherwise future developers may update the wrong field.

---

# 29. Transactional Denormalization

If a duplicated value must always change together with its source, updating both inside one transaction can help.

Example:

```text
customers.name
orders.customer_name
```

If the design truly requires both to remain synchronized, the related updates may need a transaction.

But this can still become expensive and complicated.

Prefer not to duplicate current-state facts unless there is a strong reason.

---

# 30. Asynchronous Denormalization

Sometimes eventual consistency is acceptable.

Example:

```text
customers
    ↓
event
    ↓
search index / read model
```

The derived copy may update asynchronously.

Then:

```text
source of truth
      ↓
event
      ↓
derived representation
```

This is a valid production pattern, but the application must explicitly tolerate stale data.

---

# 31. Read Models

For complex dashboards, a normalized transactional schema may require many joins and aggregations.

A separate read model can be useful:

```text
transactional tables
        ↓
aggregation/update process
        ↓
read-optimized table
        ↓
dashboard
```

This is denormalization for a specific workload.

The important part is that the read model has a clear source and update strategy.

---

# 32. Reporting Tables

Reporting workloads can justify different schema shapes.

Operational schema:

```text
orders
customers
products
payments
```

Reporting schema may contain:

```text
daily_sales_summary
```

with:

```text
date
organization_id
order_count
gross_sales
refund_total
```

This avoids recomputing expensive aggregations for every dashboard request.

The summary becomes derived data.

---

# 33. When Normalization Is Usually Better

Prefer normalized data when:

- data changes frequently
- correctness is critical
- relationships are important
- writes are common
- duplication would be difficult to synchronize
- queries are manageable with normal joins and indexes
- there is no measured performance problem

This is a strong default for transactional backend systems.

---

# 34. When Denormalization May Be Justified

Consider denormalization when:

- a read path is demonstrably expensive
- the same expensive computation happens frequently
- a read-heavy workload needs precomputed data
- a historical snapshot is required
- a reporting/read model has different needs
- an external integration requires a snapshot
- search requires a derived representation

The key word is:

> **Justified.**

---

# 35. Decision Process

Use this sequence:

```text
1. Model the data correctly
        ↓
2. Write the normal query
        ↓
3. Add appropriate indexes
        ↓
4. Measure with EXPLAIN / EXPLAIN ANALYZE
        ↓
5. Identify the real bottleneck
        ↓
6. Optimize query/application behavior
        ↓
7. Denormalize only if needed
        ↓
8. Define consistency strategy
```

This avoids premature denormalization.

---

# 36. Example: Customer Name

Normalized:

```text
orders.customer_id
customers.name
```

Query:

```sql
SELECT o.id, c.name
FROM orders o
JOIN customers c
    ON c.id = o.customer_id;
```

Usually this is the correct starting point.

Do not immediately add:

```text
orders.customer_name
```

just to avoid the join.

---

# 37. Example: Invoice Snapshot

Here duplication is intentional:

```text
invoices.customer_id
invoices.billing_name
invoices.billing_address
```

Why?

Because the invoice represents a historical document.

The duplicated fields are not trying to be the current customer profile.

They represent:

```text
customer information at invoice creation time
```

This is a valid denormalization/snapshot pattern.

---

# 38. Example: Comment Count

Normalized:

```text
posts
comments
```

Query:

```sql
SELECT COUNT(*)
FROM comments
WHERE post_id = $1;
```

If this becomes a hot and expensive operation, a derived:

```text
posts.comment_count
```

may be justified.

But then define:

```text
who updates it?
what happens on delete?
what happens after failed transactions?
how is it repaired?
```

---

# 39. Repairability Matters

Derived data can become incorrect.

A production denormalized design should ideally have a way to rebuild or reconcile it.

Example:

```sql
SELECT post_id, COUNT(*)
FROM comments
GROUP BY post_id;
```

can be used to compare actual counts with:

```text
posts.comment_count
```

A useful principle:

> **If derived data cannot be rebuilt, maintaining it safely becomes much harder.**

---

# 40. Denormalization and Transactions

Suppose a transaction:

```text
create comment
increment post.comment_count
```

must either both happen or neither happen.

Use a transaction when the consistency requirement is immediate.

Conceptually:

```text
BEGIN
    INSERT comment
    UPDATE post counter
COMMIT
```

If something fails:

```text
ROLLBACK
```

This reduces inconsistent intermediate state.

---

# 41. Denormalization and Concurrency

Counters introduce concurrency concerns.

Two requests may both do:

```text
read count
add 1
write count
```

This can lose updates.

Prefer an atomic update:

```sql
UPDATE posts
SET comment_count = comment_count + 1
WHERE id = $1;
```

The database performs the increment atomically.

Denormalization often creates concurrency considerations that normalized designs avoid.

---

# 42. Duplicate Current Data vs Historical Data

This distinction is extremely important.

### Usually avoid:

```text
customers.email
orders.customer_email
```

if both mean:

> current customer email.

### Often valid:

```text
customers.email
invoices.billing_email
```

if the second means:

> email captured when the invoice was issued.

Same physical concept.

Different business meaning.

---

# 43. Normalization and API Design

A normalized database does not mean the API must expose normalized responses.

Database:

```text
orders
customers
```

API response:

```json
{
  "id": 123,
  "customer": {
    "id": 10,
    "name": "Alice"
  },
  "total": 100
}
```

The backend can join and shape data for the API.

Do not denormalize the database merely because the API response is nested.

---

# 44. ORM Considerations

ORMs can make normalized relationships easy to load.

But careless relationship loading can cause:

```text
N+1 queries
```

Before denormalizing, inspect:

- generated SQL
- number of queries
- selected columns
- indexes
- execution plans

The ORM is not the schema.

The database design should still be based on data semantics and production workload.

---

# 45. Normalization and Indexes

A normalized schema may need indexes to make relationships efficient.

Example:

```sql
CREATE INDEX idx_orders_customer_id
ON orders(customer_id);
```

Then:

```sql
SELECT *
FROM orders
WHERE customer_id = $1;
```

can efficiently find the child rows.

Do not denormalize simply because an index was missing.

---

# 46. Normalization and Query Performance

Performance depends on more than table count.

Consider:

```text
query shape
indexes
row count
selectivity
join strategy
data size
cache behavior
```

A schema with several normalized tables can still perform well.

A denormalized table can also become slow if:

```text
rows are huge
indexes are excessive
writes are expensive
queries scan too much data
```

Normalization and performance are related, but not opposites.

---

# 47. Write Amplification

Denormalization can increase writes.

If one fact is stored in ten places:

```text
one logical update
    ↓
ten physical updates
```

This can increase:

- transaction work
- lock contention
- storage
- index maintenance
- failure scenarios

Read optimization can create write costs.

Evaluate both sides.

---

# 48. Storage Duplication

Duplicated data consumes storage.

For small values:

```text
name
status
email
```

this may not matter initially.

At very large scale, duplicated data can significantly increase:

- table size
- index size
- cache pressure
- backup size
- replication traffic

The important question is not:

> Is duplication always bad?

It is:

> Is the operational benefit worth the additional cost?

---

# 49. Denormalization and Consistency Models

A derived copy may have:

```text
strong consistency
```

or:

```text
eventual consistency
```

Strong:

```text
source updated
   ↓
derived value updated in same transaction
```

Eventual:

```text
source updated
   ↓
event emitted
   ↓
worker updates derived data later
```

Choose intentionally.

Do not accidentally introduce eventual consistency by using asynchronous updates for data that must be immediately correct.

---

# 50. Production Design Questions

Before duplicating a field, ask:

1. What is the source of truth?
2. Why is the copy needed?
3. Is the copy current or historical?
4. Can it become stale?
5. Is stale data acceptable?
6. How is it updated?
7. Can it be rebuilt?
8. What happens if the update fails?
9. Does it increase write cost?
10. Did measurement show a real performance problem?

If these questions do not have clear answers, do not denormalize yet.

---

# 51. Practical Comparison

| Approach             | Main Benefit                        | Main Cost                  |
| -------------------- | ----------------------------------- | -------------------------- |
| Normalized           | correctness, single source of truth | more joins                 |
| Denormalized         | faster/simpler hot reads            | synchronization            |
| Snapshot             | preserves historical state          | duplicated data            |
| Aggregate/read model | efficient reporting                 | refresh/rebuild complexity |

Use the simplest design that satisfies the actual requirements.

---

# 52. Production Checklist

Before shipping a schema:

### Normalization

- Each entity has a clear table.
- Independently changing facts have an appropriate owner.
- Current data is not duplicated without a reason.
- Relationships use foreign keys.

### Denormalization

- There is a documented reason.
- The source of truth is clear.
- Consistency behavior is defined.
- Update ownership is clear.
- Rebuild/reconciliation is possible when appropriate.
- Read performance has been measured.
- Write and storage costs are understood.

### Historical data

- Snapshots are explicitly treated as historical facts.
- Current and historical values are not confused.

### Performance

- Queries are measured.
- Indexes have been considered.
- N+1 has been ruled out.
- `EXPLAIN` / `EXPLAIN ANALYZE` has been used where needed.

---

# 53. Common Mistake: Denormalizing Too Early

Bad process:

```text
join exists
   ↓
joins must be slow
   ↓
duplicate everything
```

Better:

```text
join exists
   ↓
measure
   ↓
check indexes
   ↓
inspect query
   ↓
optimize
   ↓
denormalize only if justified
```

---

# 54. Common Mistake: Treating Every Duplicate as Bad

This is also wrong.

Historical snapshots are intentionally duplicated.

Example:

```text
invoice.billing_address
```

may need to remain unchanged even when:

```text
customers.address
```

changes.

The duplication has business meaning.

---

# 55. Common Mistake: No Source of Truth

Bad:

```text
users.name
orders.customer_name
payments.customer_name
tickets.customer_name
```

with no clear rule about which value is authoritative.

If these all represent current customer name, synchronization becomes difficult.

Define:

```text
customers.name
    ↓
source of truth
```

and avoid unnecessary copies.

---

# 56. Common Mistake: Denormalized Counter Without Repair Strategy

If you maintain:

```text
comment_count
```

but cannot reconcile it with:

```text
comments
```

eventual bugs become difficult to fix.

Derived data should ideally have:

```text
source
↓
calculation
↓
rebuild/reconciliation path
```

---

# 57. Common Mistake: Fixing N+1 with Duplication

If the problem is:

```text
101 queries
```

do not immediately duplicate data.

First consider:

```text
JOIN
batch query
preload
eager loading
```

The right solution is often query optimization.

---

# 58. Common Mistake: Ignoring Writes

A read optimization may increase write complexity.

Example:

```text
normalized:
1 customer update

denormalized:
1 customer update
+
1000 related copies
```

The system may become faster for reads but much more expensive for writes.

Evaluate the workload direction.

---

# 59. Common Mistake: Confusing API Shape with DB Shape

A nested API response does not require nested JSON storage.

Example API:

```json
{
  "customer": {
    "id": 1,
    "name": "Alice"
  }
}
```

can come from:

```text
customers
orders
```

with a normal SQL join.

Keep storage design and API representation conceptually separate.

---

# 60. Decision Rule

A useful default:

```text
Start normalized.
      ↓
Measure.
      ↓
Optimize queries/indexes.
      ↓
If a real workload still requires it,
denormalize deliberately.
```

This is usually safer than starting with duplicated data.

---

# 61. Production Mental Model

Think of normalization as:

```text
one fact
   ↓
one authoritative place
```

Think of denormalization as:

```text
source of truth
      ↓
derived copy
      ↓
faster/specialized access
```

The moment you create the second copy, you create a consistency responsibility.

---

# 62. Production Takeaways

1. Normalize by default.
2. Give each independently changing fact an appropriate owner.
3. Use relationships instead of duplicating current data unnecessarily.
4. Normalization reduces update, insert, and delete anomalies.
5. Normalization does not mean splitting every field into a separate table.
6. Joins are normal and are not automatically a performance problem.
7. Check indexes and query shape before denormalizing.
8. Use `EXPLAIN` and `EXPLAIN ANALYZE` to measure real performance.
9. Historical snapshots are a legitimate form of intentional duplication.
10. Read models and aggregates can justify denormalization.
11. Denormalized data needs a clear source of truth and consistency strategy.
12. Derived data should ideally be rebuildable or reconcilable.
13. Denormalization can increase write cost, storage, and concurrency complexity.
14. API response shape does not determine database normalization.
15. Start with the simplest correct model, then optimize based on evidence.

Core mental model:

> **Normalize for correctness.  
> Denormalize for a proven workload.  
> If you duplicate data, own the consistency problem.**
