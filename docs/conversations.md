---
title: Conversations and chat
tags: [architecture, integrations, chat]
---

# Conversations and chat

Forma owns a canonical conversation history so users can interact with agents in the office even when no external chat provider is connected. Slack, Discord, and similar systems are transports and remain authoritative for their native channels, threads, users, and delivery state.

## Canonical model

- A **Conversation** belongs to one workspace and may optionally link to a project or work item.
- A **Participant** references a Forma user, an agent through `(connection_id, provider_agent_id)`, or an external chat identity.
- A **Message** has an immutable author, body, timestamp, reply relationship, and optional structured mentions or attachments.
- A **Delivery** maps one canonical message to one external provider channel and records pending, delivered, or failed state.
- An **Agent turn** links a user message to a harness session, command, execution attempt, and one or more agent replies.

Provider identifiers are opaque and only unique within a connection. External messages are deduplicated by connection and provider message ID.

## Routing

```text
Forma UI or chat provider
  -> persist canonical message
  -> resolve mentions and permissions
  -> create an idempotent agent command
  -> send input or dispatch through the harness
  -> persist session events and replies
  -> stream replies to Forma and optional external deliveries
```

Mentioning an agent does not silently imply execution. The router must verify that the harness exposes the required capability, that the actor may invoke the agent, and whether approval is required. Offline providers do not cause writes to be queued by default.

## Delivery rules

- Forma persists a message before attempting external delivery.
- Retries reuse the same idempotency key and delivery record.
- An accepted harness command is not considered a completed agent turn.
- Messages imported from providers retain their original observed timestamp and a separate received timestamp.
- Provider-specific metadata remains in a namespaced native field and never changes canonical semantics.

The first implementation uses the deterministic mock harness. A Slack adapter should be the first external proof of the contract; Discord can then validate that the abstraction is not Slack-specific.
