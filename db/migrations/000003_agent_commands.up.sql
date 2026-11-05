ALTER TABLE conversation_messages ADD COLUMN IF NOT EXISTS idempotency_key TEXT;
CREATE UNIQUE INDEX IF NOT EXISTS conversation_messages_idempotency_unique
 ON conversation_messages(conversation_id, idempotency_key) WHERE idempotency_key IS NOT NULL;

CREATE TABLE IF NOT EXISTS agent_commands (
 id UUID PRIMARY KEY,
 workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
 connection_id UUID NOT NULL REFERENCES connections(id) ON DELETE RESTRICT,
 agent_turn_id UUID NOT NULL REFERENCES agent_turns(id) ON DELETE CASCADE,
 capability TEXT NOT NULL,
 idempotency_key TEXT NOT NULL,
 status TEXT NOT NULL CHECK (status IN ('pending', 'accepted', 'failed', 'uncertain')),
 provider_response JSONB NOT NULL DEFAULT '{}'::jsonb,
 error_message TEXT,
 created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
 updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
 UNIQUE (workspace_id, idempotency_key)
);
CREATE INDEX IF NOT EXISTS agent_commands_turn_idx ON agent_commands(agent_turn_id, created_at, id);
