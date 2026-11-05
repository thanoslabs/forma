DROP TABLE IF EXISTS agent_turns;
DROP TABLE IF EXISTS message_deliveries;
DROP TABLE IF EXISTS conversation_messages;
DROP TABLE IF EXISTS conversation_participants;
DROP TABLE IF EXISTS conversations;

ALTER TABLE connections DROP CONSTRAINT connections_provider_type_check;
ALTER TABLE connections ADD CONSTRAINT connections_provider_type_check
 CHECK (provider_type IN ('agent', 'project'));
