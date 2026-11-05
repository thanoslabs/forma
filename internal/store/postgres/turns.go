package postgres

import (
	"context"
	"encoding/json"
	"time"

	"github.com/thanoslabs/forma/internal/conversations"
	"github.com/thanoslabs/forma/internal/turns"
)

func (s *Store) StartTurn(ctx context.Context, turn turns.Turn) (turns.Turn, bool, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return turns.Turn{}, false, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if turn.ID == "" {
		turn.ID = newUUID()
	}
	err = tx.QueryRow(ctx, `INSERT INTO conversation_participants(id,conversation_id,participant_type,display_name,connection_id,provider_participant_id) VALUES($1,$2,'agent',$3,$4,$5) ON CONFLICT (conversation_id,connection_id,provider_participant_id) WHERE connection_id IS NOT NULL AND provider_participant_id IS NOT NULL DO UPDATE SET display_name=EXCLUDED.display_name RETURNING id`, newUUID(), turn.ConversationID, turn.AgentName, turn.ConnectionID, turn.AgentID).Scan(&turn.AgentParticipantID)
	if err != nil {
		return turns.Turn{}, false, err
	}
	tag, err := tx.Exec(ctx, `INSERT INTO agent_turns(id,conversation_id,request_message_id,agent_participant_id,idempotency_key,status) VALUES($1,$2,$3,$4,$5,'pending') ON CONFLICT (conversation_id,idempotency_key) DO NOTHING`, turn.ID, turn.ConversationID, turn.RequestMessageID, turn.AgentParticipantID, turn.IdempotencyKey)
	if err != nil {
		return turns.Turn{}, false, err
	}
	if tag.RowsAffected() == 0 {
		return turn, false, tx.Commit(ctx)
	}
	_, err = tx.Exec(ctx, `INSERT INTO agent_commands(id,workspace_id,connection_id,agent_turn_id,capability,idempotency_key,status) VALUES($1,$2,$3,$4,'input',$5,'pending')`, newUUID(), turn.WorkspaceID, turn.ConnectionID, turn.ID, turn.IdempotencyKey)
	if err != nil {
		return turns.Turn{}, false, err
	}
	if err := tx.Commit(ctx); err != nil {
		return turns.Turn{}, false, err
	}
	return turn, true, nil
}

func (s *Store) CompleteTurn(ctx context.Context, turn turns.Turn, body string, occurredAt time.Time) (conversations.Message, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return conversations.Message{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	message := conversations.Message{ID: newUUID(), ConversationID: turn.ConversationID, Body: body, Author: conversations.Author{Kind: "agent", ID: turn.AgentID, Name: turn.AgentName}, Mentions: []conversations.Mention{}, OccurredAt: occurredAt, ReplyToMessageID: turn.RequestMessageID}
	mentions, _ := json.Marshal(message.Mentions)
	if _, err := tx.Exec(ctx, `INSERT INTO conversation_messages(id,conversation_id,author_participant_id,reply_to_message_id,body,mentions,occurred_at) VALUES($1,$2,$3,$4,$5,$6::jsonb,$7)`, message.ID, message.ConversationID, turn.AgentParticipantID, turn.RequestMessageID, message.Body, string(mentions), message.OccurredAt); err != nil {
		return conversations.Message{}, err
	}
	if _, err := tx.Exec(ctx, `UPDATE agent_turns SET response_message_id=$1,status='completed',updated_at=$2 WHERE id=$3`, message.ID, occurredAt, turn.ID); err != nil {
		return conversations.Message{}, err
	}
	providerResponse, _ := json.Marshal(map[string]string{"messageId": message.ID})
	if _, err := tx.Exec(ctx, `UPDATE agent_commands SET status='accepted',provider_response=$1::jsonb,updated_at=$2 WHERE agent_turn_id=$3`, string(providerResponse), occurredAt, turn.ID); err != nil {
		return conversations.Message{}, err
	}
	if _, err := tx.Exec(ctx, `UPDATE conversations SET updated_at=$1 WHERE id=$2`, occurredAt, turn.ConversationID); err != nil {
		return conversations.Message{}, err
	}
	return message, tx.Commit(ctx)
}

func (s *Store) FailTurn(ctx context.Context, turn turns.Turn, cause error) error {
	_, err := s.pool.Exec(ctx, `WITH failed_turn AS (UPDATE agent_turns SET status='failed',error_message=$1,updated_at=now() WHERE id=$2) UPDATE agent_commands SET status='failed',error_message=$1,updated_at=now() WHERE agent_turn_id=$2`, cause.Error(), turn.ID)
	return err
}
