# Forma architecture (M0)

Forma is a self-hostable, browser-based visual control plane for agents and projects. It is **not** a harness, LLM provider, or canonical project tracker.

## Decisions
- Backend: Go; frontend: React/TypeScript; world renderer: PixiJS.
- PostgreSQL from the first migration; no SQLite fallback. External systems remain authoritative for their own resources.
- REST/OpenAPI for commands and queries; WebSockets for live events.
- Agent harnesses and project management systems integrate through independent, capability-negotiated adapter contracts.
- Support both in-process adapters and external adapters via a versioned transport contract; external protocol details are an M1 design task.
- The visual world is a projection of observed state, never the authority for dispatch, leases, or execution.
- Office look: original illustrated 2D social world inspired by the approachable Club Penguin interaction style, without reusing copyrighted assets.
- Asset packs are versioned, validated, and replaceable; gameplay geometry and interaction hotspots are independent of textures.

## Ownership boundaries
| Entity | Authority |
| --- | --- |
| Agent identity, session, runtime health | Agent harness |
| Project, work item, native assignment, workflow state | Project provider |
| Connection configuration, local workspace, conversations, visual layout, execution correlation, event history | Forma |
| Native channel, thread, user and message identity | Chat provider |
| Work-item claim and lease when using Overwatch | Overwatch |

Forma may cache external resources as projections with provider IDs, sync timestamps and staleness markers. Disconnection must not imply idle state.

## Domain model
Workspace contains Connections, Rooms and Conversations. A Connection identifies a provider and stores server-side credentials. AgentRef = (connection_id, provider_agent_id); SessionRef = (connection_id, provider_session_id). WorkItemRef = (connection_id, provider_item_id). ChatMessageRef = (connection_id, provider_message_id). Execution records correlate one or more sessions to a work item or agent turn, with explicit attempt numbers and idempotency keys. Events are append-only and carry source, provider event ID, observed timestamp, received timestamp, schema version, and correlation IDs. Avatar and room placement belong to Forma and do not mutate provider state.

## Commands and reconciliation
Commands have idempotency keys, authorized actor, target capability, pending/accepted/failed outcome and traceable provider response. Never treat an accepted dispatch as a successful execution. Adapter streams/webhooks are preferred, polling is a fallback. Durable cursors, periodic reconciliation, duplicate suppression and bounded retries are required. PostgreSQL event log is durable; LISTEN/NOTIFY may wake subscribers but must not be used as the sole event store.

Conversation WebSockets replay the persisted message timeline before emitting in-process updates. PostgreSQL remains authoritative; the in-process broker is deliberately bounded and slow subscribers reconnect for replay. Multi-replica deployments will add LISTEN/NOTIFY as a wake-up mechanism without replacing durable history.

Persisted message writes require an `Idempotency-Key`. An agent mention is dispatchable only when it includes both the provider agent ID and its Forma connection ID; a display-name mention alone is inert. Forma records an `agent_turn` and `agent_command` before invoking an input-capable adapter, then persists the correlated reply or failure. Replaying the same message key returns the original message without invoking the adapter again.

The workspace agent roster is a live, connection-scoped projection from configured agent adapters. Provider IDs are always paired with a connection ID, and the browser periodically refreshes this projection. When no agent connection is configured, the local office simulation remains available as a deliberate demo fallback.

World movement uses a logical tile grid independent of viewport scaling. Players move between adjacent walkable cells, autonomous actors use deterministic A* paths, and furniture contributes blocked and interaction cells. Pixi interpolates between committed tile positions so navigation remains visually smooth without introducing free-coordinate drift.

The browser models each workspace as an office floor. A scene manifest supplies its stable ID, floor number, spawn, and elevator cell; the elevator switches the active workspace conversation and agent projection without reloading the page. A shared street-level scene provides the building entrance, sidewalks, park space, crosswalk, and ambient traffic without changing the active workspace. Player positions are remembered independently per scene.

The street background and the initial Forma character atlas were generated with OpenAI's built-in image generation tool from original Forma-specific prompts. Runtime geometry, collision cells, entrances, cars, traffic signals, labels, and interaction state remain code-defined so the artwork can be replaced without changing world behavior.

## Security
Credentials are encrypted at rest with authenticated encryption, bound to their connection ID, and never sent to the browser. API bearer checks use constant-time comparison; health and readiness probes remain separately accessible. Provider permissions and Forma RBAC are checked before dispatch. External adapters are authenticated and constrained by workspace/provider scope. No arbitrary uploaded code runs in the Forma server. Asset packs are declarative, subject to size/type/path validation, and served under restrictive content security policies. Default deployment is private, with OIDC and explicit trusted origins before multi-user access.

Production mode validates its security boundary before opening a listener: database, API token, credential key, and trusted browser origins are mandatory, while demo routes are prohibited. Exact origin checks cover HTTP and WebSocket requests. Responses carry request correlation and hardened browser headers; server timeouts and graceful termination bound resource use and deployment disruption.

## Initial stack
Go + pgx + SQL migrations; React + TypeScript + Vite; PixiJS; PostgreSQL; Docker Compose for development; OCI and Helm for deployment. CI validates Go, frontend, migrations and adapter conformance fixtures. Avoid an extra queue or distributed scheduler until there is a measured need.

## First vertical slice
Start Postgres, open the browser, load a room from a versioned asset manifest, connect a mock harness, observe one agent transition idle -> running -> blocked -> completed, and inspect the same state in a management panel. No real task is dispatched by avatar movement.
