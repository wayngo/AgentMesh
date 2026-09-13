# PostgreSQL migrations

Phase 2 introduces the `runs` table through `001_runs.sql`. The gateway also applies the same idempotent schema on startup when `DATABASE_URL` is configured.

Example:

```text
DATABASE_URL=postgres://agentmesh:agentmesh@localhost:5432/agentmesh?sslmode=disable
```