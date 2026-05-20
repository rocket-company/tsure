# Database

Schema and seed data for the PostgreSQL-backed `tsure` app.

- `schema.sql` creates the `service_orders` table and a legacy-compatible
  operational schema derived from the Access files in `database-access/`
- `seed.sql` inserts sample service orders aligned with the leasing domain

The Go app also bootstraps the table at startup, so these files are mainly for
explicit setup or future migration tooling.
