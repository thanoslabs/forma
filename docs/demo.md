# Local demo API

The demo API is disabled by default. For local development only:

1. `make dev-up && make migrate-up`
2. `FORMA_ENABLE_DEMO_API=true make run`
3. `cd web && npm install && npm run dev`
4. Open the Vite URL and select Agents or Projects.

The Vite development server proxies `/api` to the local Go backend. The demo API exposes fake data and an intentionally unauthenticated status mutation route. **Never enable FORMA_ENABLE_DEMO_API on a publicly accessible deployment.** Production providers will use authenticated, workspace-scoped routes and persisted projections instead.
