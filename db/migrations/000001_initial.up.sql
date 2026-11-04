CREATE TABLE IF NOT EXISTS workspaces (
 id UUID PRIMARY KEY,
 name TEXT NOT NULL,
 created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS connections (
 id UUID PRIMARY KEY,
 workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
 adapter_id TEXT NOT NULL,
 provider_type TEXT NOT NULL CHECK (provider_type IN ('agent','project')),
 display_name TEXT NOT NULL,
 config JSONB NOT NULL DEFAULT '{}'::jsonb,
 created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS activity_events (
 id UUID PRIMARY KEY,
 workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
 connection_id UUID REFERENCES connections(id) ON DELETE SET NULL,
 provider_event_id TEXT,
 event_type TEXT NOT NULL,
 schema_version INT NOT NULL DEFAULT 1,
 payload JSONB NOT NULL,
 occurred_at TIMESTAMPTZ,
 received_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE UNIQUE INDEX IF NOT EXISTS activity_provider_event_unique
 ON activity_events(connection_id, provider_event_id)
 WHERE connection_id IS NOT NULL AND provider_event_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS activity_events_workspace_received_idx
 ON activity_events(workspace_id, received_at, id);
