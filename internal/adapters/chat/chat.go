// Package chat defines provider-neutral contracts for external chat systems.
// Forma conversations remain canonical; adapters transport messages to and from
// systems such as Slack and Discord.
package chat

import (
	"context"
	"errors"
	"time"
)

var ErrUnsupported = errors.New("chat capability unsupported")

type Capability string

const (
	ReadChannels Capability = "read_channels"
	ReadMessages Capability = "read_messages"
	SendMessages Capability = "send_messages"
	EditMessages Capability = "edit_messages"
	Threads      Capability = "threads"
	Reactions    Capability = "reactions"
	Attachments  Capability = "attachments"
	Subscribe    Capability = "subscribe"
)

type Channel struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	URL  string `json:"url,omitempty"`
}

type Author struct {
	ID          string `json:"id"`
	DisplayName string `json:"displayName"`
	IsBot       bool   `json:"isBot"`
}

type Message struct {
	ID        string         `json:"id"`
	ChannelID string         `json:"channelId"`
	ThreadID  string         `json:"threadId,omitempty"`
	Author    Author         `json:"author"`
	Body      string         `json:"body"`
	SentAt    time.Time      `json:"sentAt"`
	Native    map[string]any `json:"native,omitempty"`
}

type Event struct {
	ID         string    `json:"id"`
	Message    Message   `json:"message"`
	OccurredAt time.Time `json:"occurredAt"`
}

type SendRequest struct {
	ChannelID      string
	ThreadID       string
	Body           string
	IdempotencyKey string
}

type Adapter interface {
	ID() string
	Capabilities() []Capability
	ListChannels(ctx context.Context) ([]Channel, error)
	ListMessages(ctx context.Context, channelID, cursor string) ([]Message, string, error)
}

type Sender interface {
	SendMessage(ctx context.Context, request SendRequest) (Message, error)
}

type Subscriber interface {
	Subscribe(ctx context.Context, cursor string) (<-chan Event, error)
}
