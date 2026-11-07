// Package overwatch implements the bridge work-provider port using Overwatch's
// versioned REST API.
package overwatch

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/thanoslabs/forma/pkg/bridge"
	"github.com/thanoslabs/forma/pkg/orchestration"
)

type Config struct {
	URL           string
	Token         string
	ConnectionID  string
	AgentName     string
	Kind          string
	RepositoryIDs []string
	Labels        []string
	LeaseSeconds  int
	HTTP          *http.Client
}

type Provider struct {
	baseURL      string
	token        string
	connectionID string
	agentName    string
	kind         string
	repositories []string
	labels       []string
	leaseSeconds int
	http         *http.Client
}

type agentTask struct {
	ID             string `json:"id"`
	ItemID         string `json:"item_id"`
	Repository     string `json:"repository"`
	Title          string `json:"title"`
	URL            string `json:"url"`
	LeaseExpiresAt string `json:"lease_expires_at"`
}

func New(config Config) (*Provider, error) {
	parsed, err := url.Parse(strings.TrimRight(config.URL, "/"))
	if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Host == "" {
		return nil, errors.New("Overwatch URL must be absolute HTTP or HTTPS")
	}
	if strings.TrimSpace(config.ConnectionID) == "" || strings.TrimSpace(config.AgentName) == "" {
		return nil, errors.New("connection ID and agent name are required")
	}
	if config.Kind == "" {
		config.Kind = "any"
	}
	if config.Kind != "any" && config.Kind != "issue" && config.Kind != "pull_request" {
		return nil, errors.New("kind must be any, issue or pull_request")
	}
	if config.LeaseSeconds == 0 {
		config.LeaseSeconds = 900
	}
	if config.LeaseSeconds < 60 || config.LeaseSeconds > 3600 {
		return nil, errors.New("lease seconds must be between 60 and 3600")
	}
	client := config.HTTP
	if client == nil {
		client = &http.Client{Timeout: 15 * time.Second}
	}
	return &Provider{baseURL: strings.TrimRight(parsed.String(), "/") + "/api/v1", token: config.Token, connectionID: config.ConnectionID, agentName: config.AgentName, kind: config.Kind, repositories: append([]string(nil), config.RepositoryIDs...), labels: append([]string(nil), config.Labels...), leaseSeconds: config.LeaseSeconds, http: client}, nil
}

func (p *Provider) ClaimNext(ctx context.Context, executionID string) (bridge.ClaimedWork, error) {
	body := map[string]any{"kind": p.kind, "repository_ids": p.repositories, "labels": p.labels, "lease_seconds": p.leaseSeconds, "message": "execution " + executionID + " claimed by " + p.agentName}
	var task agentTask
	status, err := p.request(ctx, http.MethodPost, "/agent/claim-next", body, &task)
	if status == http.StatusNoContent {
		return bridge.ClaimedWork{}, bridge.ErrNoWork
	}
	if err != nil {
		return bridge.ClaimedWork{}, err
	}
	claim, err := p.claimRef(task)
	if err != nil {
		return bridge.ClaimedWork{}, err
	}
	prompt := task.Title + "\n\n" + task.URL
	return bridge.ClaimedWork{Work: orchestration.WorkRef{ConnectionID: p.connectionID, ProjectID: task.Repository, ItemID: task.ItemID}, Claim: claim, Prompt: prompt}, nil
}

func (p *Provider) Heartbeat(ctx context.Context, _ orchestration.WorkRef, claim orchestration.ClaimRef) (orchestration.ClaimRef, error) {
	var task agentTask
	_, err := p.request(ctx, http.MethodPost, "/agent-tasks/"+url.PathEscape(claim.ID)+"/heartbeat", map[string]int{"lease_seconds": p.leaseSeconds}, &task)
	if err != nil {
		return claim, err
	}
	return p.claimRef(task)
}

func (p *Provider) Release(ctx context.Context, _ orchestration.WorkRef, claim orchestration.ClaimRef, reason string) error {
	_, err := p.request(ctx, http.MethodPost, "/agent-tasks/"+url.PathEscape(claim.ID)+"/release", map[string]string{"message": reason}, &agentTask{})
	return err
}

func (p *Provider) Report(ctx context.Context, _ orchestration.WorkRef, claim orchestration.ClaimRef, update bridge.WorkUpdate) error {
	status := string(update.State)
	if update.State == orchestration.StateRunning {
		status = "in_progress"
	}
	if update.State == orchestration.StateReview {
		status = "needs_review"
	}
	if status != "in_progress" && status != "blocked" && status != "needs_review" && status != "failed" && status != "completed" {
		return fmt.Errorf("execution state %s cannot be reported to Overwatch", update.State)
	}
	body := map[string]any{"status": status, "message": update.Message}
	if update.Progress > 0 {
		body["progress"] = update.Progress
	}
	if update.ResultURL != "" {
		body["result_url"] = update.ResultURL
	}
	_, err := p.request(ctx, http.MethodPatch, "/agent-tasks/"+url.PathEscape(claim.ID), body, &agentTask{})
	return err
}

func (p *Provider) claimRef(task agentTask) (orchestration.ClaimRef, error) {
	if strings.TrimSpace(task.ID) == "" {
		return orchestration.ClaimRef{}, errors.New("Overwatch returned a task without an ID")
	}
	expiresAt, err := time.Parse(time.RFC3339Nano, task.LeaseExpiresAt)
	if err != nil {
		return orchestration.ClaimRef{}, fmt.Errorf("parse Overwatch lease expiry: %w", err)
	}
	return orchestration.ClaimRef{ID: task.ID, ExpiresAt: expiresAt}, nil
}

func (p *Provider) request(ctx context.Context, method, path string, input, output any) (int, error) {
	encoded, err := json.Marshal(input)
	if err != nil {
		return 0, err
	}
	request, err := http.NewRequestWithContext(ctx, method, p.baseURL+path, bytes.NewReader(encoded))
	if err != nil {
		return 0, err
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Accept", "application/json")
	if p.token != "" {
		request.Header.Set("Authorization", "Bearer "+p.token)
	}
	response, err := p.http.Do(request)
	if err != nil {
		return 0, fmt.Errorf("call Overwatch: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode == http.StatusNoContent {
		return response.StatusCode, nil
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		message, _ := io.ReadAll(io.LimitReader(response.Body, 4096))
		return response.StatusCode, fmt.Errorf("Overwatch returned %s: %s", response.Status, strings.TrimSpace(string(message)))
	}
	if err := json.NewDecoder(response.Body).Decode(output); err != nil {
		return response.StatusCode, fmt.Errorf("decode Overwatch response: %w", err)
	}
	return response.StatusCode, nil
}
