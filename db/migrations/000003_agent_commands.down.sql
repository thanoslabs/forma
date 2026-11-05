DROP TABLE IF EXISTS agent_commands;
DROP INDEX IF EXISTS conversation_messages_idempotency_unique;
ALTER TABLE conversation_messages DROP COLUMN IF EXISTS idempotency_key;
