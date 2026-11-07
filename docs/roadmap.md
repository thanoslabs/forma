# Forma roadmap

- **M0 Architecture**: ADRs, product boundaries, canonical model, adapter and asset-pack specifications, acceptance criteria.
- **M1 Core**: Go service, PostgreSQL schema/migrations, API, authentication baseline, durable event log, WebSocket replay, React shell, CI.
- **M2 Conversations & Agents**: canonical conversations, mock harness and Paseo adapter, discovery, status, session inspection and capability-aware user-to-agent messages.
- **M3 Projects**: Overwatch adapter, project/work-item projections, claim/heartbeat integration, task-to-execution correlation.
- **M4 Runtime fleet (Herdr)**: herdr-core adapter, authenticated local bridge, parent/helper hierarchy, streamed lifecycle and terminal metadata, safe steering controls, and correlation between Herdr sessions and Overwatch work items.
- **M5 Living office**: PixiJS illustrated rooms, navigation, avatars, interaction hotspots, activity visualization, versioned packs, and spatial projections of agent teams and delegation.
- **M6 Chat & Orchestration**: Slack proof of concept, chat delivery reconciliation, dispatch and approvals, execution attempts, retries, auditability, and cross-agent handoffs respecting provider ownership.
- **M7 Production**: external adapter protocol/SDK, RBAC/OIDC, observability, asset-pack customization, multi-user collaboration, Helm and scale tests.

Each milestone requires automated tests, docs, and green CI. PRs should be stacked in dependency order and kept mergeable. The first demo should use a mock harness and Postgres, with the visual world and management view reflecting the same event stream.

## M1 acceptance
1. Fresh PostgreSQL boot migrates automatically and repeatably.
2. API reports health/readiness and persists workspace/connection records.
3. Mock agent events survive server restart and can replay to a browser.
4. UI distinguishes connected, stale, and unknown provider state.
5. CI runs unit, integration, migration, and frontend tests.

## M4 acceptance: Herdr runtime fleet

1. A local connector discovers herdr-core terminals without exposing its raw control socket to the Forma server.
2. Forma renders stable parent/helper relationships and lifecycle states reported by Herdr rather than inferring them from names or screen layout.
3. Read-only status is distinct from privileged start, stop, steer, attach, and output-access capabilities.
4. Terminal output is bounded and sanitized by default; credentials and unrestricted shell access never cross the connector boundary.
5. An execution can correlate one Overwatch work item with a Herdr parent agent and its helper sessions without treating runtime completion as project completion.
6. Disconnects preserve the last observation with an explicit stale state, and reconnect reconciliation is idempotent.

## Current increment

- Startup applies embedded, versioned migrations under an advisory lock.
- Workspace, connection, conversation, and message APIs persist to PostgreSQL.
- Conversation WebSockets replay durable history before live in-process updates.
- Workspace agent projections stream to the browser, with provider capabilities and agent details available on selection.
- Overwatch repositories and issues project into the Projects panel without moving claim authority into Forma.
- Workspace desk layouts persist per floor, while the browser retains a local fallback when the API is unavailable.

The next runtime increment introduces Herdr as a second agent harness. Start with a read-only herdr-core connector and hierarchy projection, then add capability-gated steering and Overwatch execution correlation. Slack and Discord remain later delivery surfaces over Forma conversations rather than runtime authorities.
