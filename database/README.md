# Database

Schema and seed data for the PostgreSQL-backed todo app.

- `schema.sql` creates the `todos` table
- `seed.sql` inserts a couple of sample records

The Go app also bootstraps the table at startup, so these files are mainly for
explicit setup or future migration tooling.
