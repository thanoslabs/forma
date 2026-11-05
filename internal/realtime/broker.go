package realtime

import (
	"sync"

	"github.com/thanoslabs/forma/internal/conversations"
)

type Broker struct {
	mu          sync.Mutex
	subscribers map[string]map[chan conversations.Message]struct{}
}

func New() *Broker {
	return &Broker{subscribers: make(map[string]map[chan conversations.Message]struct{})}
}

func (b *Broker) Subscribe(conversationID string) (<-chan conversations.Message, func()) {
	channel := make(chan conversations.Message, 16)
	b.mu.Lock()
	if b.subscribers[conversationID] == nil {
		b.subscribers[conversationID] = make(map[chan conversations.Message]struct{})
	}
	b.subscribers[conversationID][channel] = struct{}{}
	b.mu.Unlock()
	var once sync.Once
	return channel, func() {
		once.Do(func() {
			b.mu.Lock()
			if _, ok := b.subscribers[conversationID][channel]; ok {
				delete(b.subscribers[conversationID], channel)
				close(channel)
			}
			if len(b.subscribers[conversationID]) == 0 {
				delete(b.subscribers, conversationID)
			}
			b.mu.Unlock()
		})
	}
}

func (b *Broker) Publish(message conversations.Message) {
	b.mu.Lock()
	defer b.mu.Unlock()
	for channel := range b.subscribers[message.ConversationID] {
		select {
		case channel <- message:
		default:
			delete(b.subscribers[message.ConversationID], channel)
			close(channel)
		}
	}
}
