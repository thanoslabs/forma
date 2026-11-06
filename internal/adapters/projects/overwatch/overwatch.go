// Package overwatch adapts Overwatch's versioned REST API to Forma projects.
package overwatch

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/thanoslabs/forma/internal/adapters/projects"
)

type Config struct {
	URL   string
	Token string
	HTTP  *http.Client
}

type Provider struct {
	baseURL string
	token   string
	http    *http.Client
}

func New(config Config) (*Provider, error) {
	parsed, err := url.Parse(strings.TrimRight(config.URL, "/"))
	if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Host == "" {
		return nil, errors.New("Overwatch URL must be an absolute http:// or https:// URL")
	}
	client := config.HTTP
	if client == nil {
		client = &http.Client{Timeout: 10 * time.Second}
	}
	return &Provider{baseURL: strings.TrimRight(parsed.String(), "/") + "/api/v1", token: config.Token, http: client}, nil
}

func (*Provider) ID() string { return "overwatch" }
func (*Provider) Capabilities() []projects.Capability {
	return []projects.Capability{projects.Read, projects.Claim}
}

func (p *Provider) ListProjects(ctx context.Context) ([]projects.Project, error) {
	var response []struct {
		ID       string `json:"id"`
		FullName string `json:"full_name"`
	}
	if err := p.get(ctx, "/repositories?limit=1000", &response); err != nil {
		return nil, err
	}
	result := make([]projects.Project, 0, len(response))
	for _, repository := range response {
		result = append(result, projects.Project{ID: repository.ID, Name: repository.FullName})
	}
	return result, nil
}

func (p *Provider) ListWorkItems(ctx context.Context, projectID string) ([]projects.WorkItem, error) {
	var response struct {
		Issues []struct {
			ID        string `json:"id"`
			Title     string `json:"title"`
			State     string `json:"state"`
			URL       string `json:"url"`
			UpdatedAt string `json:"updated_at"`
		} `json:"issues"`
	}
	path := "/issues?repository_id=" + url.QueryEscape(projectID) + "&limit=100"
	if err := p.get(ctx, path, &response); err != nil {
		return nil, err
	}
	result := make([]projects.WorkItem, 0, len(response.Issues))
	for _, issue := range response.Issues {
		updatedAt, _ := time.Parse(time.RFC3339Nano, issue.UpdatedAt)
		result = append(result, projects.WorkItem{ID: issue.ID, ProjectID: projectID, Title: issue.Title, Status: issue.State, URL: issue.URL, UpdatedAt: updatedAt})
	}
	return result, nil
}

func (p *Provider) get(ctx context.Context, path string, destination any) error {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, p.baseURL+path, nil)
	if err != nil {
		return err
	}
	request.Header.Set("Accept", "application/json")
	if p.token != "" {
		request.Header.Set("Authorization", "Bearer "+p.token)
	}
	response, err := p.http.Do(request)
	if err != nil {
		return fmt.Errorf("call Overwatch: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return fmt.Errorf("Overwatch returned %s", response.Status)
	}
	if err := json.NewDecoder(response.Body).Decode(destination); err != nil {
		return fmt.Errorf("decode Overwatch response: %w", err)
	}
	return nil
}
