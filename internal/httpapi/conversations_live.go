package httpapi

import (
	"context"
	"net/http"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
)

func (a *API) streamConversationMessages(w http.ResponseWriter, r *http.Request) {
	conversationID := r.PathValue("conversationID")
	updates, unsubscribe := a.live.Subscribe(conversationID)
	defer unsubscribe()
	history, err := a.conversations.ListMessages(r.Context(), r.PathValue("workspaceID"), conversationID)
	if err != nil {
		http.Error(w, "unable to replay messages", http.StatusInternalServerError)
		return
	}
	connection, err := websocket.Accept(w, r, nil)
	if err != nil {
		return
	}
	defer connection.CloseNow()
	seen := make(map[string]bool, len(history))
	for _, message := range history {
		seen[message.ID] = true
		if err := wsjson.Write(r.Context(), connection, message); err != nil {
			return
		}
	}
	for {
		select {
		case <-r.Context().Done():
			return
		case message, ok := <-updates:
			if !ok {
				_ = connection.Close(websocket.StatusTryAgainLater, "subscriber fell behind")
				return
			}
			if seen[message.ID] {
				continue
			}
			seen[message.ID] = true
			writeContext, cancel := context.WithTimeout(r.Context(), writeTimeout)
			err := wsjson.Write(writeContext, connection, message)
			cancel()
			if err != nil {
				return
			}
		}
	}
}
