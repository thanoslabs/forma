// Package projects defines provider-neutral project management contracts.
package projects

import (
	"context"
	"errors"
	"time"
)

var ErrUnsupported = errors.New("project capability unsupported")

type Capability string

const (
	Read      Capability = "read"
	Create    Capability = "create"
	Update    Capability = "update"
	Assign    Capability = "assign"
	Claim     Capability = "claim"
	Subscribe Capability = "subscribe"
)

type Project struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	URL  string `json:"url,omitempty"`
}
type WorkItem struct {
	ID        string    `json:"id"`
	ProjectID string    `json:"projectId"`
	Title     string    `json:"title"`
	Status    string    `json:"status"`
	URL       string    `json:"url,omitempty"`
	UpdatedAt time.Time `json:"updatedAt"`
}
type Adapter interface {
	ID() string
	Capabilities() []Capability
	ListProjects(ctx context.Context) ([]Project, error)
	ListWorkItems(ctx context.Context, projectID string) ([]WorkItem, error)
}
type ClaimRequest struct {
	ItemID         string
	WorkerID       string
	IdempotencyKey string
}
type ClaimResult struct {
	ClaimID   string
	ExpiresAt time.Time
}
type Claimer interface {
	Claim(ctx context.Context, req ClaimRequest) (ClaimResult, error)
	Heartbeat(ctx context.Context, claimID string) (ClaimResult, error)
	Release(ctx context.Context, claimID string) error
}
