package postgres

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	formadb "github.com/thanoslabs/forma/db"
	"github.com/thanoslabs/forma/internal/conversations"
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
	store := New(pool)
	workspaceID := newUUID()
	t.Cleanup(func() { _, _ = pool.Exec(context.Background(), "DELETE FROM workspaces WHERE id=$1", workspaceID) })
	workspace, err := store.CreateWorkspace(ctx, workspaces.Workspace{ID: workspaceID, Name: "Integration"})
	if err != nil || workspace.CreatedAt.IsZero() {
		t.Fatalf("workspace: %#v %v", workspace, err)
	}
	conversationID := newUUID()
	conversation, err := store.CreateConversation(ctx, conversations.Conversation{ID: conversationID, WorkspaceID: workspaceID, Title: "Office"})
	if err != nil || conversation.CreatedAt.IsZero() {
		t.Fatalf("conversation: %#v %v", conversation, err)
	}
	message := conversations.Message{ID: newUUID(), ConversationID: conversationID, Body: "hello @Atlas", Author: conversations.Author{Kind: "user", ID: "integration-user", Name: "Integration User"}, Mentions: []conversations.Mention{{Kind: "agent", ID: "atlas", Name: "Atlas", Text: "@Atlas", Start: 6, End: 12}}, OccurredAt: time.Now().UTC()}
	if _, err := store.CreateMessage(ctx, workspaceID, message); err != nil {
		t.Fatal(err)
	}
	messages, err := store.ListMessages(ctx, workspaceID, conversationID)
	if err != nil || len(messages) != 1 || messages[0].Author.ID != "integration-user" || len(messages[0].Mentions) != 1 || messages[0].Mentions[0].ID != "atlas" {
		t.Fatalf("messages: %#v %v", messages, err)
	}
}
