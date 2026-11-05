package main

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	formadb "github.com/thanoslabs/forma/db"
	"github.com/thanoslabs/forma/internal/adapters/agents/mock"
	"github.com/thanoslabs/forma/internal/httpapi"
	"github.com/thanoslabs/forma/internal/projections"
	"github.com/thanoslabs/forma/internal/realtime"
	"github.com/thanoslabs/forma/internal/store/postgres"
	"github.com/thanoslabs/forma/internal/turns"
)

func main() {
	ctx := context.Background()
	url := os.Getenv("DATABASE_URL")
	if url == "" {
		log.Fatal("DATABASE_URL is required")
	}
	pool, err := pgxpool.New(ctx, url)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()
	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := pool.Ping(pingCtx); err != nil {
		log.Fatal(err)
	}
	if err := formadb.MigrateUp(ctx, pool); err != nil {
		log.Fatal(err)
	}

	mux := http.NewServeMux()
	store := postgres.New(pool)
	live := realtime.New()
	registry := turns.NewRegistry(mock.New())
	router := turns.NewRouter(store, store, registry, live)
	httpapi.New(store, store, live, router).WithAgentProjections(projections.NewAgentReader(store, registry)).Register(mux)
	if os.Getenv("FORMA_ENABLE_DEMO_API") == "true" {
		httpapi.NewMockAPI().Register(mux)
	}
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})
	mux.HandleFunc("GET /readyz", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()
		if err := pool.Ping(ctx); err != nil {
			http.Error(w, "database unavailable", http.StatusServiceUnavailable)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "ready"})
	})
	addr := os.Getenv("FORMA_LISTEN_ADDR")
	if addr == "" {
		addr = ":8080"
	}
	server := &http.Server{Addr: addr, Handler: mux, ReadHeaderTimeout: 5 * time.Second}
	log.Printf("Forma listening on %s", addr)
	if err := server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		log.Fatal(err)
	}
}
