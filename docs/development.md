# Development

Install the exact Go and Node versions from `.tool-versions` with mise or asdf. Docker Compose and `psql` are also required.

1. Run `make dev-up`.
2. Run `make run`; the service applies embedded, versioned migrations before listening.
3. Run `cd web && npm ci && npm run dev`.
4. Open `http://127.0.0.1:5173` and check `http://localhost:8080/healthz` and `/readyz`.

Set `FORMA_ENABLE_DEMO_API=true` only for local development. The demo routes are intentionally unauthenticated. Before production, add managed migration versioning, authentication, encrypted connection credentials, integration tests, and durable event replay. The database schema remains provisional and will evolve with the adapter contracts.

To exercise Paseo, set `FORMA_PASEO_URL` to its `ws://` or `wss://` daemon endpoint and optionally set `FORMA_PASEO_PASSWORD`, then create a workspace agent connection with adapter ID `paseo`. Forma uses Paseo's WebSocket RPC protocol directly.

Set `FORMA_CREDENTIAL_KEY` to a base64-encoded 32-byte key before creating connections with credentials. Credentials use AES-256-GCM and are bound to the connection ID; plaintext is write-only and never returned by the API. Set `FORMA_API_TOKEN` to require a constant-time checked bearer token on every `/api/` route. Health and readiness probes remain unauthenticated. In browser deployments, terminate user authentication at a trusted same-origin proxy that injects the service bearer token; never embed this token in frontend assets.

Production startup is fail-closed. Set `FORMA_ENV=production`, a bearer token of at least 32 bytes, the credential key, and a comma-separated `FORMA_TRUSTED_ORIGINS` list of exact browser origins. The demo API is rejected in production. Production responses include request IDs and browser security headers, cross-origin and WebSocket handshakes are restricted to the allowlist, server request/header/idle timeouts are bounded, and SIGINT/SIGTERM trigger a ten-second graceful shutdown.
