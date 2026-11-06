package main

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	formadb "github.com/thanoslabs/forma/db"
	"github.com/thanoslabs/forma/internal/adapters/agents"
	"github.com/thanoslabs/forma/internal/adapters/agents/mock"
	"github.com/thanoslabs/forma/internal/adapters/agents/paseo"
	"github.com/thanoslabs/forma/internal/config"
	"github.com/thanoslabs/forma/internal/httpapi"
	"github.com/thanoslabs/forma/internal/projections"
	"github.com/thanoslabs/forma/internal/realtime"
	"github.com/thanoslabs/forma/internal/secrets"
	"github.com/thanoslabs/forma/internal/store/postgres"
	"github.com/thanoslabs/forma/internal/turns"
	"github.com/thanoslabs/forma/internal/workspaces"
)

func main() {
	settings, err := config.Load(os.Getenv)
	if err != nil {
		log.Fatal(err)
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	pool, err := pgxpool.New(ctx, settings.DatabaseURL)
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
	var credentialCipher *secrets.Cipher
	if encodedKey := settings.CredentialKey; encodedKey != "" {
		credentialCipher, err = secrets.FromBase64(encodedKey)
		if err != nil {
			log.Fatal(err)
		}
	}
	store := postgres.NewWithCipher(pool, credentialCipher)
	live := realtime.New()
	agentAdapters := []agents.Adapter{mock.New()}
	if settings.PaseoURL != "" {
		provider, err := paseo.New(paseo.Config{URL: settings.PaseoURL, Password: settings.PaseoPassword})
		if err != nil {
			log.Fatal(err)
		}
		agentAdapters = append(agentAdapters, provider)
	}
	registry := turns.NewRegistry(agentAdapters...)
	registry.RegisterFactory("paseo", func(_ context.Context, connection workspaces.Connection) (agents.Adapter, error) {
		return paseo.New(paseo.Config{URL: connection.EndpointURL, Password: connection.Credential})
	})
	router := turns.NewRouter(store, store, registry, live)
	agentReader := projections.NewAgentReader(store, registry)
	httpapi.New(store, store, live, router).WithAgentProjections(agentReader).WithConnectionTester(agentReader).WithAgentDetails(agentReader).Register(mux)
	if settings.EnableDemoAPI {
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
	handler := httpapi.RequireBearer(mux, settings.APIToken)
	if settings.Production {
		handler = httpapi.ProductionGuards(handler, settings.TrustedOrigins)
	}
	server := &http.Server{Addr: settings.ListenAddress, Handler: handler, ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 30 * time.Second, IdleTimeout: 60 * time.Second, MaxHeaderBytes: 1 << 20}
	serverErrors := make(chan error, 1)
	go func() { serverErrors <- server.ListenAndServe() }()
	log.Printf("Forma listening on %s", settings.ListenAddress)
	select {
	case err := <-serverErrors:
		if !errors.Is(err, http.ErrServerClosed) {
			log.Fatal(err)
		}
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			log.Printf("graceful shutdown failed: %v", err)
		}
	}
}
