// Package worlds owns workspace-specific virtual office configuration.
package worlds

import (
	"context"
	"time"
)

type Point struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

type Configuration struct {
	Version     int       `json:"version"`
	Desks       []Point   `json:"desks"`
	WorkspaceID string    `json:"workspaceId,omitempty"`
	UpdatedAt   time.Time `json:"updatedAt,omitempty"`
}

type Repository interface {
	GetWorld(context.Context, string) (Configuration, error)
	PutWorld(context.Context, string, Configuration) (Configuration, error)
}
