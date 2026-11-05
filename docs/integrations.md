# Integration contracts (M0 proposal)

## Adapter categories
AgentHarnessAdapter: discover agents/sessions, observe health/state, subscribe to events; optional create session, dispatch, send input, cancel, resume, inspect logs/artifacts.
ProjectProviderAdapter: discover projects/work items, read metadata and status, subscribe or poll; optional create, edit, native assign, claim/lease, update outcome, link back to provider.

ChatProviderAdapter: discover channels and messages, subscribe to inbound events; optional send/edit messages, threads, reactions and attachments. Forma owns canonical conversations and deliveries; chat providers remain authoritative for native channel, thread, user and message identities. See [[conversations]].

Adapters declare a versioned manifest with category, adapter ID, protocol version, config schema, authentication type and supported capabilities. Unsupported actions must be hidden, not emulated silently. Provider IDs are opaque and scoped to a connection. API payloads use a canonical envelope plus a namespaced native metadata field.

## Chat providers

Chat ingestion verifies provider signatures before accepting an event, stores a durable cursor, and deduplicates retries by `(connection_id, provider_message_id)`. Outbound delivery uses stable idempotency keys and records uncertain outcomes for reconciliation. A chat adapter never invokes an agent directly; Forma's conversation router resolves permissions and delegates agent input to an AgentHarnessAdapter.

## Overwatch first
Overwatch already exposes /api/v1, /openapi.yaml, scoped bearer tokens, and MCP at /mcp. Its documented agent workflow supports claim-next, heartbeat, update-task, and release-task. Overwatch owns work claims, 15-minute default leases and retry policies; Forma must not create competing leases. Use the REST API for strongly typed integration initially, reserving MCP as an optional transport. Validate precise endpoints and response schemas against Overwatch's OpenAPI before implementation.

## Harnesses
Paseo is the first target harness, subject to inspecting its real session and event APIs. Start with a deterministic mock harness to test the UI and adapter conformance. Do not assume every harness supports session creation or dispatch.

## Failure behavior
If a provider is offline, show last-observed state with a stale/unknown indicator. Queue no mutation by default while disconnected. Write operations need idempotency and reconciliation after uncertain responses. Correlate tasks to executions without overwriting native provider assignees.

## Conformance tests
Discover empty and populated sources; missing optional capabilities; duplicate/out-of-order events; reconnect from cursor; expired authentication; idempotent command replay; provider timeout; permission denial; stale-state rendering; native ID collision across connections.
