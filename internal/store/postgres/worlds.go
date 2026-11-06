package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/thanoslabs/forma/internal/workspaces"
	"github.com/thanoslabs/forma/internal/worlds"
)

func (s *Store) GetWorld(ctx context.Context, workspaceID string) (worlds.Configuration, error) {
	var config worlds.Configuration
	var encoded []byte
	var updatedAt time.Time
	err := s.pool.QueryRow(ctx, "SELECT config, updated_at FROM workspace_world_configs WHERE workspace_id=$1", workspaceID).Scan(&encoded, &updatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return worlds.Configuration{}, workspaces.ErrNotFound
	}
	if err != nil {
		return config, err
	}
	if err := json.Unmarshal(encoded, &config); err != nil {
		return worlds.Configuration{}, err
	}
	config.WorkspaceID = workspaceID
	config.UpdatedAt = updatedAt
	return config, nil
}

func (s *Store) PutWorld(ctx context.Context, workspaceID string, config worlds.Configuration) (worlds.Configuration, error) {
	encoded, err := json.Marshal(config)
	if err != nil {
		return worlds.Configuration{}, err
	}
	err = s.pool.QueryRow(ctx, `INSERT INTO workspace_world_configs(workspace_id, config) VALUES($1,$2)
 ON CONFLICT(workspace_id) DO UPDATE SET config=EXCLUDED.config, updated_at=now()
 RETURNING updated_at`, workspaceID, encoded).Scan(&config.UpdatedAt)
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23503" {
		return worlds.Configuration{}, workspaces.ErrNotFound
	}
	config.WorkspaceID = workspaceID
	return config, err
}
