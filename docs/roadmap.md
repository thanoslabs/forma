# Forma roadmap

- **M0 Architecture**: ADRs, product boundaries, canonical model, adapter and asset-pack specifications, acceptance criteria.
- **M1 Core**: Go service, PostgreSQL schema/migrations, API, authentication baseline, durable event log, WebSocket replay, React shell, CI.
- **M2 Agents**: mock harness and Paseo adapter, discovery, status and session inspection, capability-aware actions.
- **M3 Projects**: Overwatch adapter, project/work-item projections, claim/heartbeat integration, task-to-execution correlation.
- **M4 Living office**: PixiJS illustrated rooms, navigation, avatars, interaction hotspots, activity visualization, versioned packs.
- **M5 Orchestration**: dispatch and approvals, execution attempts, retries, auditability, cross-agent handoffs respecting provider ownership.
- **M6 Production**: external adapter protocol/SDK, RBAC/OIDC, observability, asset-pack customization, multi-user collaboration, Helm and scale tests.

Each milestone requires automated tests, docs, and green CI. PRs should be stacked in dependency order and kept mergeable. The first demo should use a mock harness and Postgres, with the visual world and management view reflecting the same event stream.

## M1 acceptance
1. Fresh PostgreSQL boot migrates automatically and repeatably.
2. API reports health/readiness and persists workspace/connection records.
3. Mock agent events survive server restart and can replay to a browser.
4. UI distinguishes connected, stale, and unknown provider state.
5. CI runs unit, integration, migration, and frontend tests.
