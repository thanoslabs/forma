ALTER TABLE connections DROP CONSTRAINT connections_provider_type_check;
ALTER TABLE connections ADD CONSTRAINT connections_provider_type_check
 CHECK (provider_type IN ('agent', 'project', 'chat'));

CREATE TABLE conversations (
 id UUID PRIMARY KEY,
 workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
 title TEXT NOT NULL,
 project_connection_id UUID REFERENCES connections(id) ON DELETE SET NULL,
 provider_project_id TEXT,
 provider_work_item_id TEXT,
 created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
 updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
 CHECK ((provider_project_id IS NULL AND provider_work_item_id IS NULL) OR project_connection_id IS NOT NULL)
);
CREATE INDEX conversations_workspace_updated_idx
 ON conversations(workspace_id, updated_at DESC, id);

CREATE TABLE conversation_participants (
 id UUID PRIMARY KEY,
 conversation_id UUID NOT NULL REFERENCES conversations(id) ON DELETE CASCADE,
 participant_type TEXT NOT NULL CHECK (participant_type IN ('user', 'agent', 'external')),
 display_name TEXT NOT NULL,
 user_id TEXT,
 connection_id UUID REFERENCES connections(id) ON DELETE CASCADE,
 provider_participant_id TEXT,
 created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
 CHECK (
  (participant_type = 'user' AND user_id IS NOT NULL AND connection_id IS NULL AND provider_participant_id IS NULL)
  OR (participant_type IN ('agent', 'external') AND user_id IS NULL AND connection_id IS NOT NULL AND provider_participant_id IS NOT NULL)
 )
);
CREATE UNIQUE INDEX conversation_participants_user_unique
 ON conversation_participants(conversation_id, user_id)
 WHERE user_id IS NOT NULL;
CREATE UNIQUE INDEX conversation_participants_provider_unique
 ON conversation_participants(conversation_id, connection_id, provider_participant_id)
 WHERE connection_id IS NOT NULL AND provider_participant_id IS NOT NULL;

CREATE TABLE conversation_messages (
 id UUID PRIMARY KEY,
 conversation_id UUID NOT NULL REFERENCES conversations(id) ON DELETE CASCADE,
 author_participant_id UUID REFERENCES conversation_participants(id) ON DELETE SET NULL,
 reply_to_message_id UUID REFERENCES conversation_messages(id) ON DELETE SET NULL,
 body TEXT NOT NULL CHECK (length(body) BETWEEN 1 AND 20000),
 mentions JSONB NOT NULL DEFAULT '[]'::jsonb,
 native_metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
 occurred_at TIMESTAMPTZ NOT NULL,
 received_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX conversation_messages_timeline_idx
 ON conversation_messages(conversation_id, occurred_at, id);

CREATE TABLE message_deliveries (
 id UUID PRIMARY KEY,
 message_id UUID NOT NULL REFERENCES conversation_messages(id) ON DELETE CASCADE,
 connection_id UUID NOT NULL REFERENCES connections(id) ON DELETE CASCADE,
 provider_channel_id TEXT NOT NULL,
 provider_thread_id TEXT,
 provider_message_id TEXT,
 idempotency_key TEXT NOT NULL,
 status TEXT NOT NULL CHECK (status IN ('pending', 'delivered', 'failed', 'uncertain')),
 error_message TEXT,
 attempted_at TIMESTAMPTZ,
 delivered_at TIMESTAMPTZ,
 created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
 UNIQUE (connection_id, idempotency_key)
);
CREATE UNIQUE INDEX message_deliveries_provider_message_unique
 ON message_deliveries(connection_id, provider_message_id)
 WHERE provider_message_id IS NOT NULL;

CREATE TABLE agent_turns (
 id UUID PRIMARY KEY,
 conversation_id UUID NOT NULL REFERENCES conversations(id) ON DELETE CASCADE,
 request_message_id UUID NOT NULL REFERENCES conversation_messages(id) ON DELETE RESTRICT,
 agent_participant_id UUID NOT NULL REFERENCES conversation_participants(id) ON DELETE RESTRICT,
 response_message_id UUID REFERENCES conversation_messages(id) ON DELETE SET NULL,
 provider_session_id TEXT,
 idempotency_key TEXT NOT NULL,
 status TEXT NOT NULL CHECK (status IN ('pending', 'accepted', 'running', 'completed', 'failed', 'cancelled')),
 error_message TEXT,
 created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
 updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
 UNIQUE (conversation_id, idempotency_key)
);
CREATE INDEX agent_turns_conversation_created_idx
 ON agent_turns(conversation_id, created_at, id);
