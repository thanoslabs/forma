package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/thanoslabs/forma/internal/secrets"
	"github.com/thanoslabs/forma/internal/workspaces"
)

type Store struct {
	pool   *pgxpool.Pool
	cipher *secrets.Cipher
}

func New(pool *pgxpool.Pool) *Store { return &Store{pool: pool} }
func NewWithCipher(pool *pgxpool.Pool, cipher *secrets.Cipher) *Store {
	return &Store{pool: pool, cipher: cipher}
}

func (s *Store) ListWorkspaces(ctx context.Context) ([]workspaces.Workspace, error) {
	rows, err := s.pool.Query(ctx, "SELECT id, name, created_at FROM workspaces ORDER BY created_at, id")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := make([]workspaces.Workspace, 0)
	for rows.Next() {
		var workspace workspaces.Workspace
		if err := rows.Scan(&workspace.ID, &workspace.Name, &workspace.CreatedAt); err != nil {
			return nil, err
		}
		result = append(result, workspace)
	}
	return result, rows.Err()
}

func (s *Store) CreateWorkspace(ctx context.Context, workspace workspaces.Workspace) (workspaces.Workspace, error) {
	err := s.pool.QueryRow(ctx, "INSERT INTO workspaces(id, name) VALUES($1, $2) RETURNING created_at", workspace.ID, workspace.Name).Scan(&workspace.CreatedAt)
	return workspace, err
}

func (s *Store) ListConnections(ctx context.Context, workspaceID string) ([]workspaces.Connection, error) {
	rows, err := s.pool.Query(ctx, `SELECT id, workspace_id, adapter_id, provider_type, display_name, COALESCE(endpoint_url,''), created_at
 FROM connections WHERE workspace_id=$1 ORDER BY created_at, id`, workspaceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := make([]workspaces.Connection, 0)
	for rows.Next() {
		var connection workspaces.Connection
		if err := rows.Scan(&connection.ID, &connection.WorkspaceID, &connection.AdapterID, &connection.ProviderType, &connection.DisplayName, &connection.EndpointURL, &connection.CreatedAt); err != nil {
			return nil, err
		}
		result = append(result, connection)
	}
	return result, rows.Err()
}

func (s *Store) GetConnection(ctx context.Context, workspaceID, connectionID string) (workspaces.Connection, error) {
	var connection workspaces.Connection
	var ciphertext []byte
	err := s.pool.QueryRow(ctx, `SELECT id, workspace_id, adapter_id, provider_type, display_name, COALESCE(endpoint_url,''), credential_ciphertext, created_at FROM connections WHERE workspace_id=$1 AND id=$2`, workspaceID, connectionID).Scan(&connection.ID, &connection.WorkspaceID, &connection.AdapterID, &connection.ProviderType, &connection.DisplayName, &connection.EndpointURL, &ciphertext, &connection.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return workspaces.Connection{}, workspaces.ErrNotFound
	}
	if err != nil {
		return connection, err
	}
	if len(ciphertext) > 0 {
		if s.cipher == nil {
			return workspaces.Connection{}, errors.New("connection credential cipher is unavailable")
		}
		plaintext, err := s.cipher.Open(ciphertext, []byte(connection.ID))
		if err != nil {
			return workspaces.Connection{}, err
		}
		connection.Credential = string(plaintext)
	}
	return connection, nil
}

func (s *Store) CreateConnection(ctx context.Context, connection workspaces.Connection) (workspaces.Connection, error) {
	var ciphertext []byte
	var err error
	if connection.Credential != "" {
		if s.cipher == nil {
			return workspaces.Connection{}, errors.New("connection credential cipher is unavailable")
		}
		ciphertext, err = s.cipher.Seal([]byte(connection.Credential), []byte(connection.ID))
		if err != nil {
			return workspaces.Connection{}, err
		}
	}
	err = s.pool.QueryRow(ctx, `INSERT INTO connections(id, workspace_id, adapter_id, provider_type, display_name, endpoint_url, credential_ciphertext)
 VALUES($1, $2, $3, $4, $5, NULLIF($6,''), $7) RETURNING created_at`, connection.ID, connection.WorkspaceID, connection.AdapterID, connection.ProviderType, connection.DisplayName, connection.EndpointURL, ciphertext).Scan(&connection.CreatedAt)
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23503" {
		return workspaces.Connection{}, workspaces.ErrNotFound
	}
	return connection, err
}
