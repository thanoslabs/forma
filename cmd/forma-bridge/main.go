package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/thanoslabs/forma/internal/bridgeapp"
	"github.com/thanoslabs/forma/pkg/bridge"
	"github.com/thanoslabs/forma/pkg/bridge/filestore"
	"github.com/thanoslabs/forma/pkg/bridge/herdr"
	"github.com/thanoslabs/forma/pkg/bridge/overwatch"
)

func main() {
	config, err := bridgeapp.LoadConfig()
	if err != nil {
		log.Fatal(err)
	}
	store, err := filestore.Open(config.StatePath)
	if err != nil {
		log.Fatal(err)
	}
	work, err := overwatch.New(overwatch.Config{URL: config.OverwatchURL, Token: config.OverwatchToken, ConnectionID: config.OverwatchConnectionID, AgentName: config.AgentName, LeaseSeconds: config.LeaseSeconds})
	if err != nil {
		log.Fatal(err)
	}
	runtime, err := herdr.New(herdr.Config{Binary: config.HerdrBinary, ConnectionID: config.HerdrConnectionID, Machine: config.HerdrMachine, WorkingDir: config.WorkingDirectory, Command: config.AgentCommand})
	if err != nil {
		log.Fatal(err)
	}
	coordinator, err := bridge.New(work, runtime, store, newID)
	if err != nil {
		log.Fatal(err)
	}
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	supervisor := bridgeapp.NewSupervisor(coordinator, store, config.PollInterval, config.HeartbeatInterval, config.MaxConcurrent, logger)
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	mux := http.NewServeMux()
	bridgeapp.NewObservationAPI(store, config.ObserverToken).Register(mux)
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})
	mux.HandleFunc("GET /readyz", func(w http.ResponseWriter, r *http.Request) {
		if _, err := store.ListActive(r.Context()); err != nil {
			http.Error(w, "store unavailable", 503)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"ready"}`))
	})
	server := &http.Server{Addr: config.ListenAddress, Handler: mux, ReadHeaderTimeout: 5 * time.Second}
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("probe server failed", "error", err)
			stop()
		}
	}()
	logger.Info("bridge started", "listen", config.ListenAddress)
	_ = supervisor.Run(ctx)
	shutdown, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = server.Shutdown(shutdown)
}
func newID() string {
	var value [16]byte
	if _, err := rand.Read(value[:]); err != nil {
		panic(err)
	}
	return hex.EncodeToString(value[:])
}
