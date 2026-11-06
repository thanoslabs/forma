package httpapi

import (
	"errors"
	"math"
	"net/http"

	"github.com/thanoslabs/forma/internal/workspaces"
	"github.com/thanoslabs/forma/internal/worlds"
)

func (a *API) getWorld(w http.ResponseWriter, r *http.Request) {
	if a.worlds == nil {
		http.Error(w, "world configuration is unavailable", http.StatusServiceUnavailable)
		return
	}
	config, err := a.worlds.GetWorld(r.Context(), r.PathValue("workspaceID"))
	if errors.Is(err, workspaces.ErrNotFound) {
		http.NotFound(w, r)
		return
	}
	if err != nil {
		http.Error(w, "unable to load world configuration", http.StatusInternalServerError)
		return
	}
	respond(w, config)
}

func (a *API) putWorld(w http.ResponseWriter, r *http.Request) {
	if a.worlds == nil {
		http.Error(w, "world configuration is unavailable", http.StatusServiceUnavailable)
		return
	}
	var config worlds.Configuration
	if !decodeJSON(w, r, &config) {
		return
	}
	if !validWorld(config) {
		http.Error(w, "world version 1 requires exactly three in-bounds desk positions", http.StatusBadRequest)
		return
	}
	config.WorkspaceID = ""
	config.UpdatedAt = config.UpdatedAt.UTC()
	saved, err := a.worlds.PutWorld(r.Context(), r.PathValue("workspaceID"), config)
	if errors.Is(err, workspaces.ErrNotFound) {
		http.NotFound(w, r)
		return
	}
	if err != nil {
		http.Error(w, "unable to save world configuration", http.StatusInternalServerError)
		return
	}
	respond(w, saved)
}

func validWorld(config worlds.Configuration) bool {
	if config.Version != 1 || len(config.Desks) != 3 {
		return false
	}
	for _, point := range config.Desks {
		if math.IsNaN(point.X) || math.IsNaN(point.Y) || math.IsInf(point.X, 0) || math.IsInf(point.Y, 0) || point.X < 110 || point.X > 760 || point.Y < 120 || point.Y > 300 {
			return false
		}
	}
	return true
}
