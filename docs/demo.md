# Local demo API

The demo API is disabled by default. For local development only:

1. `make dev-up`
2. `FORMA_ENABLE_DEMO_API=true make run` (startup applies pending migrations)
3. `cd web && npm ci && npm run dev`
4. Open the Vite URL and select Chat. Forma creates a default persisted workspace and conversation when the database is empty.

The Vite development server proxies `/api` to the local Go backend. Chat history is stored in PostgreSQL and replayed before live WebSocket updates. If the API is unavailable, the web client clearly switches to a browser-session-only fallback. Agent mentions are persisted as structured references; production harness routing remains a separate capability-aware step.

The demo API also exposes fake agent and project data plus an intentionally unauthenticated status mutation route. **Never enable FORMA_ENABLE_DEMO_API on a publicly accessible deployment.** Production providers will use authenticated, workspace-scoped routes and persisted projections instead.
