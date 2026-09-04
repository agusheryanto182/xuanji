# PostgreSQL Playground

A hands-on playground for learning PostgreSQL from fundamentals to advanced concepts.

## Setup

Start PostgreSQL with Docker Compose:

```bash
docker compose up -d
```

Check running containers:

```bash
docker ps
```

Stop the PostgreSQL container:

```bash
docker compose down
```

Stop the container and remove its volume/data:

```bash
docker compose down -v
```

---

# Docker Commands

## Enter PostgreSQL Container

Open a shell inside the PostgreSQL container:

```bash
docker exec -it xuanji-postgres bash
```

However, you usually don't need to enter the container shell first. You can directly open `psql`:

```bash
docker exec -it xuanji-postgres psql -U xuanji -d xuanji
```

You should see:

```text
xuanji=#
```

This means you are now inside the PostgreSQL CLI.

---

## Check Container

```bash
docker ps
```

Show all containers, including stopped containers:

```bash
docker ps -a
```

View PostgreSQL logs:

```bash
docker logs xuanji-postgres
```

Follow logs in real time:

```bash
docker logs -f xuanji-postgres
```

Restart PostgreSQL:

```bash
docker restart xuanji-postgres
```

---

# psql Commands

These commands are **psql commands**, not SQL commands.

## Show Databases

```sql
\l
```

Example:

```text
                                  List of databases
   Name    | Owner  | Encoding | ...
-----------+--------+----------+-----
 postgres  | xuanji | UTF8     |
 template0 | xuanji | UTF8     |
 template1 | xuanji | UTF8     |
 xuanji    | xuanji | UTF8     |
```

---

## Connect to a Database

```sql
\c xuanji
```

After connecting:

```text
You are now connected to database "xuanji".
```

---

## Show Current Database

```sql
SELECT current_database();
```

---

## Show Current User

```sql
SELECT current_user;
```

---

## Show PostgreSQL Version

```sql
SELECT version();
```

---

## List Tables

```sql
\dt
```

If there are no tables:

```text
Did not find any relations.
```

---

## Describe a Table

After creating a table:

```sql
\d users
```

This shows:

* columns
* data types
* nullable status
* default values
* indexes
* primary keys
* foreign keys

For more detailed information:

```sql
\d+ users
```

---

## List Schemas

```sql
\dn
```

---

## List Users / Roles

```sql
\du
```

---

## Show Current Connection

```sql
\conninfo
```

---

## Clear Screen

```sql
\! clear
```

On Windows:

```sql
\! cls
```

---

## Exit psql

```sql
\q
```

---

# SQL Commands

Unlike `\dt` or `\d`, these are actual SQL statements.

## Create Table

```sql
CREATE TABLE users (
    id BIGSERIAL PRIMARY KEY,
    username TEXT NOT NULL,
    email TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

---

## Insert Data

```sql
INSERT INTO users (username, email)
VALUES ('agus', 'agus@example.com');
```

Insert multiple rows:

```sql
INSERT INTO users (username, email)
VALUES
    ('agus', 'agus@example.com'),
    ('john', 'john@example.com'),
    ('jane', 'jane@example.com');
```

---

## Select Data

Select everything:

```sql
SELECT *
FROM users;
```

Select specific columns:

```sql
SELECT id, username, email
FROM users;
```

---

## Filter Data

```sql
SELECT *
FROM users
WHERE username = 'agus';
```

Multiple conditions:

```sql
SELECT *
FROM users
WHERE username = 'agus'
  AND email = 'agus@example.com';
```

---

## Sort Data

Ascending:

```sql
SELECT *
FROM users
ORDER BY username ASC;
```

Descending:

```sql
SELECT *
FROM users
ORDER BY username DESC;
```

---

## Limit Results

```sql
SELECT *
FROM users
LIMIT 10;
```

---

## Update Data

```sql
UPDATE users
SET username = 'agus-heryanto'
WHERE id = 1;
```

**Always be careful with `UPDATE` without `WHERE`.**

This:

```sql
UPDATE users
SET username = 'test';
```

will update every row.

---

## Delete Data

```sql
DELETE FROM users
WHERE id = 1;
```

Again, be careful with:

```sql
DELETE FROM users;
```

This deletes all rows from the table.

---

# Useful SQL Inspection

## Count Rows

```sql
SELECT COUNT(*)
FROM users;
```

---

## Check Table Structure

```sql
\d users
```

---

## Check Indexes

```sql
\di
```

Or for a specific table:

```sql
\d users
```

---

## Check Constraints

```sql
SELECT
    constraint_name,
    constraint_type
FROM information_schema.table_constraints
WHERE table_name = 'users';
```

---

# Transaction Commands

Start a transaction:

```sql
BEGIN;
```

Make changes:

```sql
UPDATE users
SET username = 'temporary'
WHERE id = 1;
```

Undo:

```sql
ROLLBACK;
```

Or save:

```sql
COMMIT;
```

---

# Useful psql Shortcuts

| Command       | Description                |
| ------------- | -------------------------- |
| `\l`          | List databases             |
| `\c database` | Connect to database        |
| `\dt`         | List tables                |
| `\d table`    | Describe table             |
| `\d+ table`   | Detailed table information |
| `\dn`         | List schemas               |
| `\du`         | List users/roles           |
| `\di`         | List indexes               |
| `\conninfo`   | Show current connection    |
| `\q`          | Exit psql                  |

---

# Typical Workflow

Start PostgreSQL:

```bash
docker compose up -d
```

Check container:

```bash
docker ps
```

Connect directly to PostgreSQL:

```bash
docker exec -it xuanji-postgres psql -U xuanji -d xuanji
```

Inside `psql`:

```sql
\dt
```

Create a table:

```sql
CREATE TABLE users (
    id BIGSERIAL PRIMARY KEY,
    username TEXT NOT NULL,
    email TEXT NOT NULL
);
```

Check the table:

```sql
\dt
```

Inspect the table:

```sql
\d users
```

Insert data:

```sql
INSERT INTO users (username, email)
VALUES ('agus', 'agus@example.com');
```

Read data:

```sql
SELECT *
FROM users;
```

Exit:

```sql
\q
```

Stop PostgreSQL:

```bash
docker compose down
```

---

# Important Distinction

There are two kinds of commands you'll use:

### Docker commands

Run in your terminal:

```bash
docker ps
docker logs xuanji-postgres
docker exec -it xuanji-postgres psql -U xuanji -d xuanji
```

### psql commands

Run after entering PostgreSQL:

```sql
\dt
\d users
\l
\du
\q
```

### SQL commands

Run inside PostgreSQL:

```sql
SELECT *
FROM users;

INSERT INTO users (...);

UPDATE users
SET ...;

DELETE FROM users
WHERE ...;
```