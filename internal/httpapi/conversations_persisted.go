package httpapi

import (
	"encoding/json"
	"errors"
	"github.com/thanoslabs/forma/internal/conversations"
	"net/http"
	"strings"
	"time"
)

func (a *API) registerConversations(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/v1/workspaces/{workspaceID}/conversations", a.listConversations)
	mux.HandleFunc("POST /api/v1/workspaces/{workspaceID}/conversations", a.createConversation)
	mux.HandleFunc("GET /api/v1/workspaces/{workspaceID}/conversations/{conversationID}/messages", a.listConversationMessages)
	mux.HandleFunc("POST /api/v1/workspaces/{workspaceID}/conversations/{conversationID}/messages", a.createConversationMessage)
	mux.HandleFunc("GET /api/v1/workspaces/{workspaceID}/conversations/{conversationID}/messages/ws", a.streamConversationMessages)
}
func (a *API) listConversations(w http.ResponseWriter, r *http.Request) {
	items, err := a.conversations.ListConversations(r.Context(), r.PathValue("workspaceID"))
	if err != nil {
		http.Error(w, "unable to list conversations", 500)
		return
	}
	respond(w, items)
}
func (a *API) createConversation(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Title string `json:"title"`
	}
	if !decodeJSON(w, r, &request) {
		return
	}
	request.Title = strings.TrimSpace(request.Title)
	if request.Title == "" || len(request.Title) > 160 {
		http.Error(w, "title must contain between 1 and 160 bytes", 400)
		return
	}
	item, err := a.conversations.CreateConversation(r.Context(), conversations.Conversation{ID: newID(), WorkspaceID: r.PathValue("workspaceID"), Title: request.Title})
	if errors.Is(err, conversations.ErrNotFound) {
		http.NotFound(w, r)
		return
	}
	if err != nil {
		http.Error(w, "unable to create conversation", 500)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Location", "/api/v1/workspaces/"+item.WorkspaceID+"/conversations/"+item.ID)
	w.WriteHeader(201)
	_ = json.NewEncoder(w).Encode(item)
}
func (a *API) listConversationMessages(w http.ResponseWriter, r *http.Request) {
	items, err := a.conversations.ListMessages(r.Context(), r.PathValue("workspaceID"), r.PathValue("conversationID"))
	if err != nil {
		http.Error(w, "unable to list messages", 500)
		return
	}
	respond(w, items)
}
func (a *API) createConversationMessage(w http.ResponseWriter, r *http.Request) {
	idempotencyKey := strings.TrimSpace(r.Header.Get("Idempotency-Key"))
	if idempotencyKey == "" || len(idempotencyKey) > 200 {
		http.Error(w, "Idempotency-Key must contain between 1 and 200 bytes", http.StatusBadRequest)
		return
	}
	var request struct {
		Body     string                  `json:"body"`
		Mentions []conversations.Mention `json:"mentions"`
	}
	if !decodeJSON(w, r, &request) {
		return
	}
	request.Body = strings.TrimSpace(request.Body)
	if request.Body == "" || len(request.Body) > 20000 || !validMentions(request.Body, request.Mentions) {
		http.Error(w, "invalid message", 400)
		return
	}
	workspaceID := r.PathValue("workspaceID")
	message, created, err := a.conversations.CreateMessage(r.Context(), workspaceID, idempotencyKey, conversations.Message{ID: newID(), ConversationID: r.PathValue("conversationID"), Body: request.Body, Author: conversations.Author{Kind: "user", ID: "local-user", Name: "You"}, Mentions: request.Mentions, OccurredAt: time.Now().UTC()})
	if errors.Is(err, conversations.ErrNotFound) {
		http.NotFound(w, r)
		return
	}
	if err != nil {
		http.Error(w, "unable to create message", 500)
		return
	}
	if created {
		a.live.Publish(message)
		if a.router != nil {
			_ = a.router.Route(r.Context(), workspaceID, idempotencyKey, message)
		}
	}
	w.Header().Set("Content-Type", "application/json")
	if created {
		w.WriteHeader(http.StatusCreated)
	} else {
		w.WriteHeader(http.StatusOK)
	}
	_ = json.NewEncoder(w).Encode(message)
}
func validMentions(body string, mentions []conversations.Mention) bool {
	for _, mention := range mentions {
		if mention.Start < 0 || mention.End <= mention.Start || mention.End > len(body) {
			return false
		}
	}
	return true
}
