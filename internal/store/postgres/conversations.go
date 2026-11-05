package postgres

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/thanoslabs/forma/internal/conversations"
)

func (s *Store) ListConversations(ctx context.Context, workspaceID string) ([]conversations.Conversation, error) {
	rows, err := s.pool.Query(ctx, `SELECT id,workspace_id,title,created_at,updated_at FROM conversations WHERE workspace_id=$1 ORDER BY updated_at DESC,id`, workspaceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := make([]conversations.Conversation, 0)
	for rows.Next() {
		var item conversations.Conversation
		if err := rows.Scan(&item.ID, &item.WorkspaceID, &item.Title, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, err
		}
		result = append(result, item)
	}
	return result, rows.Err()
}
func (s *Store) CreateConversation(ctx context.Context, item conversations.Conversation) (conversations.Conversation, error) {
	err := s.pool.QueryRow(ctx, `INSERT INTO conversations(id,workspace_id,title) VALUES($1,$2,$3) RETURNING created_at,updated_at`, item.ID, item.WorkspaceID, item.Title).Scan(&item.CreatedAt, &item.UpdatedAt)
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23503" {
		return conversations.Conversation{}, conversations.ErrNotFound
	}
	return item, err
}
func (s *Store) ListMessages(ctx context.Context, workspaceID, conversationID string) ([]conversations.Message, error) {
	rows, err := s.pool.Query(ctx, `SELECT m.id,m.conversation_id,m.body,COALESCE(p.participant_type,'system'),COALESCE(p.user_id,p.provider_participant_id,''),COALESCE(p.display_name,'Forma'),m.mentions,m.occurred_at,COALESCE(m.reply_to_message_id::text,'') FROM conversation_messages m JOIN conversations c ON c.id=m.conversation_id LEFT JOIN conversation_participants p ON p.id=m.author_participant_id WHERE c.workspace_id=$1 AND c.id=$2 ORDER BY m.occurred_at,m.id`, workspaceID, conversationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := make([]conversations.Message, 0)
	for rows.Next() {
		message, err := scanMessage(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, message)
	}
	return result, rows.Err()
}
func (s *Store) CreateMessage(ctx context.Context, workspaceID, idempotencyKey string, message conversations.Message) (conversations.Message, bool, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return conversations.Message{}, false, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	var exists bool
	if err := tx.QueryRow(ctx, "SELECT EXISTS (SELECT 1 FROM conversations WHERE id=$1 AND workspace_id=$2)", message.ConversationID, workspaceID).Scan(&exists); err != nil {
		return conversations.Message{}, false, err
	}
	if !exists {
		return conversations.Message{}, false, conversations.ErrNotFound
	}
	var participantID string
	err = tx.QueryRow(ctx, `INSERT INTO conversation_participants(id,conversation_id,participant_type,display_name,user_id) VALUES($1,$2,'user',$3,$4) ON CONFLICT (conversation_id,user_id) WHERE user_id IS NOT NULL DO UPDATE SET display_name=EXCLUDED.display_name RETURNING id`, newUUID(), message.ConversationID, message.Author.Name, message.Author.ID).Scan(&participantID)
	if err != nil {
		return conversations.Message{}, false, err
	}
	mentions, err := json.Marshal(message.Mentions)
	if err != nil {
		return conversations.Message{}, false, err
	}
	tag, err := tx.Exec(ctx, `INSERT INTO conversation_messages(id,conversation_id,author_participant_id,reply_to_message_id,body,mentions,occurred_at,idempotency_key) VALUES($1,$2,$3,NULLIF($4,'')::uuid,$5,$6::jsonb,$7,$8) ON CONFLICT (conversation_id,idempotency_key) WHERE idempotency_key IS NOT NULL DO NOTHING`, message.ID, message.ConversationID, participantID, message.ReplyToMessageID, message.Body, string(mentions), message.OccurredAt, idempotencyKey)
	if err != nil {
		return conversations.Message{}, false, err
	}
	if tag.RowsAffected() == 0 {
		existing, err := scanMessage(tx.QueryRow(ctx, `SELECT m.id,m.conversation_id,m.body,COALESCE(p.participant_type,'system'),COALESCE(p.user_id,p.provider_participant_id,''),COALESCE(p.display_name,'Forma'),m.mentions,m.occurred_at,COALESCE(m.reply_to_message_id::text,'') FROM conversation_messages m LEFT JOIN conversation_participants p ON p.id=m.author_participant_id WHERE m.conversation_id=$1 AND m.idempotency_key=$2`, message.ConversationID, idempotencyKey))
		if err != nil {
			return conversations.Message{}, false, err
		}
		if err := tx.Commit(ctx); err != nil {
			return conversations.Message{}, false, err
		}
		return existing, false, nil
	}
	if _, err := tx.Exec(ctx, "UPDATE conversations SET updated_at=$1 WHERE id=$2", message.OccurredAt, message.ConversationID); err != nil {
		return conversations.Message{}, false, err
	}
	if err := tx.Commit(ctx); err != nil {
		return conversations.Message{}, false, err
	}
	return message, true, nil
}

type rowScanner interface{ Scan(...any) error }

func scanMessage(row rowScanner) (conversations.Message, error) {
	var message conversations.Message
	var mentions []byte
	err := row.Scan(&message.ID, &message.ConversationID, &message.Body, &message.Author.Kind, &message.Author.ID, &message.Author.Name, &mentions, &message.OccurredAt, &message.ReplyToMessageID)
	if err != nil {
		return conversations.Message{}, err
	}
	if err := json.Unmarshal(mentions, &message.Mentions); err != nil {
		return conversations.Message{}, err
	}
	if message.Mentions == nil {
		message.Mentions = []conversations.Mention{}
	}
	return message, nil
}
