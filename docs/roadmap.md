# Forma roadmap

- **M0 Architecture**: ADRs, product boundaries, canonical model, adapter and asset-pack specifications, acceptance criteria.
- **M1 Core**: Go service, PostgreSQL schema/migrations, API, authentication baseline, durable event log, WebSocket replay, React shell, CI.
- **M2 Conversations & Agents**: canonical conversations, mock harness and Paseo adapter, discovery, status, session inspection and capability-aware user-to-agent messages.
- **M3 Projects**: Overwatch adapter, project/work-item projections, claim/heartbeat integration, task-to-execution correlation.
- **M4 Living office**: PixiJS illustrated rooms, navigation, avatars, interaction hotspots, activity visualization, versioned packs.
- **M5 Chat & Orchestration**: Slack proof of concept, chat delivery reconciliation, dispatch and approvals, execution attempts, retries, auditability, and cross-agent handoffs respecting provider ownership.
- **M6 Production**: external adapter protocol/SDK, RBAC/OIDC, observability, asset-pack customization, multi-user collaboration, Helm and scale tests.

Each milestone requires automated tests, docs, and green CI. PRs should be stacked in dependency order and kept mergeable. The first demo should use a mock harness and Postgres, with the visual world and management view reflecting the same event stream.

## M1 acceptance
1. Fresh PostgreSQL boot migrates automatically and repeatably.
2. API reports health/readiness and persists workspace/connection records.
3. Mock agent events survive server restart and can replay to a browser.
4. UI distinguishes connected, stale, and unknown provider state.
5. CI runs unit, integration, migration, and frontend tests.

## Current increment

- Startup applies embedded, versioned migrations under an advisory lock.
- Workspace, connection, conversation, and message APIs persist to PostgreSQL.
- Conversation WebSockets replay durable history before live in-process updates.
- The browser bootstraps a shared workspace and falls back locally when the API is unavailable.

The next increment routes structured agent mentions through configured harness connections with idempotent agent turns, then projects resulting state into the office.
