// Package paseo adapts the Paseo daemon WebSocket RPC protocol to Forma agents.
package paseo

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/coder/websocket"
	"github.com/thanoslabs/forma/internal/adapters/agents"
)

type Config struct {
	URL      string
	Password string
	Timeout  time.Duration
}

type Provider struct{ config Config }

func New(config Config) (*Provider, error) {
	parsed, err := url.Parse(config.URL)
	if err != nil || (parsed.Scheme != "ws" && parsed.Scheme != "wss") || parsed.Host == "" {
		return nil, errors.New("Paseo URL must be an absolute ws:// or wss:// URL")
	}
	if config.Timeout <= 0 {
		config.Timeout = 10 * time.Second
	}
	return &Provider{config: config}, nil
}

func (*Provider) ID() string { return "paseo" }
func (*Provider) Capabilities() []agents.Capability {
	return []agents.Capability{agents.Observe, agents.Input}
}
func (*Provider) ListSessions(context.Context, string) ([]agents.Session, error) {
	return nil, agents.ErrUnsupported
}

func (p *Provider) ListAgents(ctx context.Context) ([]agents.Agent, error) {
	requestID := newID()
	var response struct {
		Type    string `json:"type"`
		Payload struct {
			Entries []struct {
				Agent struct {
					ID        string  `json:"id"`
					Title     *string `json:"title"`
					Status    string  `json:"status"`
					UpdatedAt string  `json:"updatedAt"`
				} `json:"agent"`
			} `json:"entries"`
		} `json:"payload"`
	}
	if err := p.call(ctx, requestID, map[string]any{"type": "fetch_agents_request", "requestId": requestID, "page": map[string]any{"limit": 200}}, "fetch_agents_response", &response); err != nil {
		return nil, err
	}
	result := make([]agents.Agent, 0, len(response.Payload.Entries))
	for _, entry := range response.Payload.Entries {
		name := entry.Agent.ID
		if entry.Agent.Title != nil && strings.TrimSpace(*entry.Agent.Title) != "" {
			name = *entry.Agent.Title
		}
		lastSeen, _ := time.Parse(time.RFC3339Nano, entry.Agent.UpdatedAt)
		result = append(result, agents.Agent{ID: entry.Agent.ID, Name: name, Status: mapStatus(entry.Agent.Status), LastSeen: lastSeen})
	}
	return result, nil
}

func (p *Provider) SendInputAsync(ctx context.Context, agentID, body, idempotencyKey string) error {
	if strings.TrimSpace(agentID) == "" || strings.TrimSpace(body) == "" || strings.TrimSpace(idempotencyKey) == "" {
		return errors.New("agent ID, body, and idempotency key are required")
	}
	requestID := newID()
	var response struct {
		Payload struct {
			Accepted bool    `json:"accepted"`
			Error    *string `json:"error"`
		} `json:"payload"`
	}
	request := map[string]any{"type": "send_agent_message_request", "requestId": requestID, "agentId": agentID, "text": body, "messageId": idempotencyKey}
	if err := p.call(ctx, requestID, request, "send_agent_message_response", &response); err != nil {
		return err
	}
	if !response.Payload.Accepted {
		if response.Payload.Error != nil {
			return errors.New(*response.Payload.Error)
		}
		return errors.New("Paseo rejected the agent message")
	}
	return nil
}

func (p *Provider) call(parent context.Context, requestID string, request any, responseType string, destination any) error {
	ctx, cancel := context.WithTimeout(parent, p.config.Timeout)
	defer cancel()
	headers := http.Header{}
	if p.config.Password != "" {
		headers.Set("Authorization", "Bearer "+p.config.Password)
	}
	connection, _, err := websocket.Dial(ctx, p.config.URL, &websocket.DialOptions{HTTPHeader: headers})
	if err != nil {
		return fmt.Errorf("connect to Paseo: %w", err)
	}
	defer connection.Close(websocket.StatusNormalClosure, "request complete")
	connection.SetReadLimit(4 << 20)
	hello := map[string]any{"type": "hello", "clientId": "forma-" + newID(), "clientType": "cli", "protocolVersion": 1, "capabilities": map[string]any{}}
	if p.config.Password != "" {
		hello["auth"] = map[string]string{"kind": "password", "password": p.config.Password}
	}
	if err := writeJSON(ctx, connection, hello); err != nil {
		return err
	}
	if err := writeJSON(ctx, connection, map[string]any{"type": "session", "message": request}); err != nil {
		return err
	}
	for {
		_, payload, err := connection.Read(ctx)
		if err != nil {
			return fmt.Errorf("read Paseo response: %w", err)
		}
		var envelope struct {
			Type    string          `json:"type"`
			Message json.RawMessage `json:"message"`
			Reason  string          `json:"reason"`
		}
		if err := json.Unmarshal(payload, &envelope); err != nil {
			continue
		}
		if envelope.Type == "ping" {
			if err := writeJSON(ctx, connection, map[string]string{"type": "pong"}); err != nil {
				return err
			}
			continue
		}
		if envelope.Type == "hello.rejected" {
			return fmt.Errorf("Paseo authentication rejected: %s", envelope.Reason)
		}
		if envelope.Type != "session" {
			continue
		}
		var identity struct {
			Type    string `json:"type"`
			Payload struct {
				RequestID string `json:"requestId"`
			} `json:"payload"`
		}
		if json.Unmarshal(envelope.Message, &identity) != nil || identity.Payload.RequestID != requestID {
			continue
		}
		if identity.Type == "rpc_error" {
			return errors.New("Paseo RPC request failed")
		}
		if identity.Type != responseType {
			continue
		}
		return json.Unmarshal(envelope.Message, destination)
	}
}

func writeJSON(ctx context.Context, connection *websocket.Conn, value any) error {
	payload, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return connection.Write(ctx, websocket.MessageText, payload)
}

func mapStatus(status string) agents.Status {
	switch status {
	case "idle":
		return agents.Idle
	case "running", "initializing":
		return agents.Running
	case "error":
		return agents.Failed
	case "closed":
		return agents.Completed
	default:
		return agents.Unknown
	}
}

func newID() string {
	var value [16]byte
	if _, err := rand.Read(value[:]); err != nil {
		panic(err)
	}
	return hex.EncodeToString(value[:])
}
