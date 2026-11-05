# Local demo API

The demo API is disabled by default. For local development only:

1. `make dev-up && make migrate-up`
2. `FORMA_ENABLE_DEMO_API=true make run`
3. `cd web && npm ci && npm run dev`
4. Open the Vite URL, select Chat, and send `hello @Atlas`.

The Vite development server proxies `/api` to the local Go backend. The conversation endpoint keeps an in-memory history and routes agent mentions through the deterministic mock harness. Requests require an idempotency key, so retrying the same request does not duplicate messages. If the API is unavailable, the web client clearly switches to a browser-session-only fallback.

The demo API also exposes fake agent and project data plus an intentionally unauthenticated status mutation route. **Never enable FORMA_ENABLE_DEMO_API on a publicly accessible deployment.** Production providers will use authenticated, workspace-scoped routes and persisted projections instead.
