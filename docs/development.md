# Development

Install the exact Go and Node versions from `.tool-versions` with mise or asdf. Docker Compose and `psql` are also required.

1. Run `make dev-up`.
2. Run `make run`; the service applies embedded, versioned migrations before listening.
3. Run `cd web && npm ci && npm run dev`.
4. Open `http://127.0.0.1:5173` and check `http://localhost:8080/healthz` and `/readyz`.

Set `FORMA_ENABLE_DEMO_API=true` only for local development. The demo routes are intentionally unauthenticated. Before production, add managed migration versioning, authentication, encrypted connection credentials, integration tests, and durable event replay. The database schema remains provisional and will evolve with the adapter contracts.
