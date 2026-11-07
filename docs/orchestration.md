---
title: Independent orchestration
tags: [architecture, orchestration, herdr, overwatch]
date: 2026-11-06
---

# Independent orchestration

Overwatch and Herdr form an independently functioning work system. Forma is its visual operations console, not a required intermediary.

```text
Overwatch ── work, claims, leases, outcomes ── Bridge ── agents, helpers, terminals ── Herdr
     │                                           │                                  │
     └────────────────── observations and typed operator commands ──────────────────┘
                                                 │
                                               Forma
```

## Authority

| Concern | Authority |
| --- | --- |
| Work queue, priority, claims, leases, attempts, retry policy and workflow state | Overwatch |
| Terminals, processes, machines, worktrees, parent/helper hierarchy and runtime state | Herdr |
| Cross-system execution correlation, recovery and orchestration commands | Standalone bridge |
| Rooms, avatars, conversations, projections and operator experience | Forma |

Forma sends intent to an authoritative API. It does not independently claim a work item and start a terminal, and it never treats avatar movement or a settled terminal as proof that project work completed.

## Execution envelope

Every attempt has a stable execution ID and records the Overwatch work item and claim, the Herdr parent terminal and helper terminals, attempt number, lifecycle timestamps, result links and last reconciled provider states. Provider IDs remain scoped to their connection or installation.

## Implementation sequence

1. Define and validate the versioned execution/event envelope.
2. Build an independently runnable bridge with typed Overwatch and Herdr ports.
3. Implement claim, heartbeat, release and restart recovery with durable state.
4. Publish a replayable observation feed without requiring Forma.
5. Connect Forma read-only and render work-to-agent and parent-to-helper relationships.
6. Add explicitly authorized operator commands such as steer, stop and approve.

## Acceptance

> [!important]
> Stop Forma, submit eligible work to Overwatch, verify that the bridge claims it and completes it through Herdr, restart Forma, and reconstruct the entire execution from authoritative state and replayed observations.

- A bridge crash must not create a second active attempt for the same claim.
- A lost heartbeat must reconcile against Overwatch before retrying or releasing.
- Herdr `done` means the runtime settled; Overwatch alone determines work completion.
- Forma downtime must not pause claims, heartbeats, agents or outcome reporting.
- Read-only observations and privileged runtime controls use distinct capabilities.
- Raw Herdr sockets, unrestricted shell access and unbounded terminal output are never exposed to Forma.
