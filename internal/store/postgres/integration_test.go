package postgres

import (
	"bytes"
	"context"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	formadb "github.com/thanoslabs/forma/db"
	"github.com/thanoslabs/forma/internal/conversations"
	"github.com/thanoslabs/forma/internal/secrets"
	"github.com/thanoslabs/forma/internal/workspaces"
)

func TestWorkspaceConversationRoundTrip(t *testing.T) {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		t.Skip("DATABASE_URL is not set")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	if err := formadb.MigrateUp(ctx, pool); err != nil {
		t.Fatal(err)
	}
	cipher, err := secrets.New(bytes.Repeat([]byte{3}, 32))
	if err != nil {
		t.Fatal(err)
	}
	store := NewWithCipher(pool, cipher)
	workspaceID := newUUID()
	t.Cleanup(func() { _, _ = pool.Exec(context.Background(), "DELETE FROM workspaces WHERE id=$1", workspaceID) })
	workspace, err := store.CreateWorkspace(ctx, workspaces.Workspace{ID: workspaceID, Name: "Integration"})
	if err != nil || workspace.CreatedAt.IsZero() {
		t.Fatalf("workspace: %#v %v", workspace, err)
	}
	connectionID := newUUID()
	if _, err := store.CreateConnection(ctx, workspaces.Connection{ID: connectionID, WorkspaceID: workspaceID, AdapterID: "mock", ProviderType: "agent", DisplayName: "Mock", EndpointURL: "wss://example.test", Credential: "provider-secret"}); err != nil {
		t.Fatal(err)
	}
	connection, err := store.GetConnection(ctx, workspaceID, connectionID)
	if err != nil || connection.Credential != "provider-secret" {
		t.Fatalf("connection=%#v err=%v", connection, err)
	}
	var storedCredential string
	if err := pool.QueryRow(ctx, "SELECT credential_ciphertext::text FROM connections WHERE id=$1", connectionID).Scan(&storedCredential); err != nil || storedCredential == "provider-secret" {
		t.Fatalf("credential was not encrypted: %q %v", storedCredential, err)
	}
	conversationID := newUUID()
	conversation, err := store.CreateConversation(ctx, conversations.Conversation{ID: conversationID, WorkspaceID: workspaceID, Title: "Office"})
	if err != nil || conversation.CreatedAt.IsZero() {
		t.Fatalf("conversation: %#v %v", conversation, err)
	}
	message := conversations.Message{ID: newUUID(), ConversationID: conversationID, Body: "hello @Atlas", Author: conversations.Author{Kind: "user", ID: "integration-user", Name: "Integration User"}, Mentions: []conversations.Mention{{Kind: "agent", ID: "atlas", Name: "Atlas", Text: "@Atlas", Start: 6, End: 12}}, OccurredAt: time.Now().UTC()}
	if _, created, err := store.CreateMessage(ctx, workspaceID, "integration-message", message); err != nil || !created {
		t.Fatal(err)
	}
	messages, err := store.ListMessages(ctx, workspaceID, conversationID)
	if err != nil || len(messages) != 1 || messages[0].Author.ID != "integration-user" || len(messages[0].Mentions) != 1 || messages[0].Mentions[0].ID != "atlas" {
		t.Fatalf("messages: %#v %v", messages, err)
	}
}
