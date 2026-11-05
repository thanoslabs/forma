package httpapi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type ConversationAuthor struct {
	Kind string `json:"kind"`
	ID   string `json:"id"`
	Name string `json:"name"`
}
type ConversationMention struct {
	Kind  string `json:"kind"`
	ID    string `json:"id"`
	Name  string `json:"name"`
	Text  string `json:"text"`
	Start int    `json:"start"`
	End   int    `json:"end"`
}
type ConversationMessage struct {
	ID               string                `json:"id"`
	Body             string                `json:"body"`
	Author           ConversationAuthor    `json:"author"`
	Mentions         []ConversationMention `json:"mentions"`
	OccurredAt       time.Time             `json:"occurredAt"`
	ReplyToMessageID string                `json:"replyToMessageId,omitempty"`
}
type sendConversationMessageRequest struct {
	Body     string                `json:"body"`
	Mentions []ConversationMention `json:"mentions"`
}
type sendConversationMessageResponse struct {
	Messages []ConversationMessage `json:"messages"`
}

func (a *MockAPI) registerConversations(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/v1/demo/conversations/default/messages", func(w http.ResponseWriter, _ *http.Request) {
		a.mu.Lock()
		messages := append([]ConversationMessage(nil), a.messages...)
		a.mu.Unlock()
		respond(w, messages)
	})
	mux.HandleFunc("POST /api/v1/demo/conversations/default/messages", a.sendConversationMessage)
}

func (a *MockAPI) sendConversationMessage(w http.ResponseWriter, r *http.Request) {
	if !strings.HasPrefix(r.Header.Get("Content-Type"), "application/json") {
		http.Error(w, "expected application/json", http.StatusUnsupportedMediaType)
		return
	}
	idempotencyKey := strings.TrimSpace(r.Header.Get("Idempotency-Key"))
	if idempotencyKey == "" {
		http.Error(w, "Idempotency-Key is required", http.StatusBadRequest)
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, 24*1024)
	var request sendConversationMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	request.Body = strings.TrimSpace(request.Body)
	if len(request.Body) == 0 || len(request.Body) > 20000 {
		http.Error(w, "body must contain between 1 and 20000 bytes", http.StatusBadRequest)
		return
	}
	for _, mention := range request.Mentions {
		if mention.Start < 0 || mention.End <= mention.Start || mention.End > len(request.Body) {
			http.Error(w, "invalid mention range", http.StatusBadRequest)
			return
		}
	}

	a.mu.Lock()
	defer a.mu.Unlock()
	if previous, ok := a.idempotency[idempotencyKey]; ok {
		respond(w, sendConversationMessageResponse{Messages: previous})
		return
	}
	now := time.Now().UTC()
	userMessage := ConversationMessage{ID: a.messageID(), Body: request.Body, Author: ConversationAuthor{Kind: "user", ID: "demo-user", Name: "You"}, Mentions: request.Mentions, OccurredAt: now}
	created := []ConversationMessage{userMessage}
	seenAgents := make(map[string]bool)
	for _, mention := range request.Mentions {
		if mention.Kind != "agent" || seenAgents[mention.ID] {
			continue
		}
		seenAgents[mention.ID] = true
		reply, err := a.agents.SendInput(r.Context(), mention.ID, request.Body, idempotencyKey+":"+mention.ID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnprocessableEntity)
			return
		}
		created = append(created, ConversationMessage{ID: a.messageID(), Body: reply, Author: ConversationAuthor{Kind: "agent", ID: mention.ID, Name: mention.Name}, Mentions: []ConversationMention{}, OccurredAt: now.Add(time.Millisecond), ReplyToMessageID: userMessage.ID})
	}
	a.messages = append(a.messages, created...)
	a.idempotency[idempotencyKey] = append([]ConversationMessage(nil), created...)
	respond(w, sendConversationMessageResponse{Messages: created})
}
func (a *MockAPI) messageID() string {
	id := fmt.Sprintf("message-%d", a.nextMessage)
	a.nextMessage++
	return id
}
