# Development

Requires Go 1.24+, Docker Compose and psql.

1. Run `make dev-up`.
2. Run `make migrate-up` (uses local development credentials; never use these in production).
3. Run `make run`.
4. Check `http://localhost:8080/healthz` and `/readyz`.

This is an initial backend foundation, not a finished M1. Before production, add managed migration versioning, authentication, encrypted connection credentials, integration tests, event replay and a frontend. The database schema is provisional and will evolve with adapter contracts.
