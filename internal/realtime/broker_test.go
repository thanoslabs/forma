package realtime

import (
	"testing"
	"time"

	"github.com/thanoslabs/forma/internal/conversations"
)

func TestBrokerPublishesByConversation(t *testing.T) {
	broker := New()
	messages, cancel := broker.Subscribe("conversation-1")
	defer cancel()
	broker.Publish(conversations.Message{ID: "message-1", ConversationID: "conversation-1"})
	select {
	case message := <-messages:
		if message.ID != "message-1" {
			t.Fatalf("message: %#v", message)
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for message")
	}
}
