package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
	"github.com/thanoslabs/forma/internal/projections"
)

type agentSnapshot struct {
	Type        string              `json:"type"`
	WorkspaceID string              `json:"workspaceId"`
	Agents      []projections.Agent `json:"agents"`
	ObservedAt  time.Time           `json:"observedAt"`
}

func (a *API) streamAgents(w http.ResponseWriter, r *http.Request) {
	if a.agents == nil {
		http.Error(w, "agent projections are unavailable", http.StatusServiceUnavailable)
		return
	}
	connection, err := websocket.Accept(w, r, nil)
	if err != nil {
		return
	}
	defer connection.CloseNow()
	workspaceID := r.PathValue("workspaceID")
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	var previous string
	for {
		agents, readErr := a.agents.ListAgents(r.Context(), workspaceID)
		if readErr != nil && len(agents) == 0 {
			_ = connection.Close(websocket.StatusTryAgainLater, "agent provider unavailable")
			return
		}
		encoded, _ := json.Marshal(agents)
		if current := string(encoded); current != previous {
			writeContext, cancel := context.WithTimeout(r.Context(), writeTimeout)
			err = wsjson.Write(writeContext, connection, agentSnapshot{Type: "agents.snapshot", WorkspaceID: workspaceID, Agents: agents, ObservedAt: time.Now().UTC()})
			cancel()
			if err != nil {
				return
			}
			previous = current
		}
		select {
		case <-r.Context().Done():
			return
		case <-ticker.C:
		}
	}
}
