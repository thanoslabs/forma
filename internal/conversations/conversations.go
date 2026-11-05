package conversations

import (
	"context"
	"errors"
	"time"
)

var ErrNotFound = errors.New("conversation not found")

type Conversation struct {
	ID          string    `json:"id"`
	WorkspaceID string    `json:"workspaceId"`
	Title       string    `json:"title"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}
type Author struct {
	Kind string `json:"kind"`
	ID   string `json:"id"`
	Name string `json:"name"`
}
type Mention struct {
	Kind         string `json:"kind"`
	ID           string `json:"id"`
	ConnectionID string `json:"connectionId,omitempty"`
	Name         string `json:"name"`
	Text         string `json:"text"`
	Start        int    `json:"start"`
	End          int    `json:"end"`
}
type Message struct {
	ID               string    `json:"id"`
	ConversationID   string    `json:"conversationId"`
	Body             string    `json:"body"`
	Author           Author    `json:"author"`
	Mentions         []Mention `json:"mentions"`
	OccurredAt       time.Time `json:"occurredAt"`
	ReplyToMessageID string    `json:"replyToMessageId,omitempty"`
}
type Repository interface {
	ListConversations(context.Context, string) ([]Conversation, error)
	CreateConversation(context.Context, Conversation) (Conversation, error)
	ListMessages(context.Context, string, string) ([]Message, error)
	CreateMessage(context.Context, string, string, Message) (Message, bool, error)
}
