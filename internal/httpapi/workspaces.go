package httpapi

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/thanoslabs/forma/internal/conversations"
	"github.com/thanoslabs/forma/internal/projections"
	"github.com/thanoslabs/forma/internal/realtime"
	"github.com/thanoslabs/forma/internal/workspaces"
)

type MessageRouter interface {
	Route(context.Context, string, string, conversations.Message) error
}

type AgentProjectionReader interface {
	ListAgents(context.Context, string) ([]projections.Agent, error)
}

type ConnectionTester interface {
	CheckConnection(context.Context, string, string) (projections.ConnectionCheck, error)
}

type API struct {
	workspaces       workspaces.Repository
	conversations    conversations.Repository
	live             *realtime.Broker
	router           MessageRouter
	agents           AgentProjectionReader
	connectionTester ConnectionTester
}

const writeTimeout = 10 * time.Second

func New(workspaceRepository workspaces.Repository, conversationRepository conversations.Repository, live *realtime.Broker, routers ...MessageRouter) *API {
	if live == nil {
		live = realtime.New()
	}
	api := &API{workspaces: workspaceRepository, conversations: conversationRepository, live: live}
	if len(routers) > 0 {
		api.router = routers[0]
	}
	return api
}

func (a *API) WithAgentProjections(reader AgentProjectionReader) *API {
	a.agents = reader
	return a
}

func (a *API) WithConnectionTester(tester ConnectionTester) *API {
	a.connectionTester = tester
	return a
}

func (a *API) Register(mux *http.ServeMux) {
	a.registerConversations(mux)
	mux.HandleFunc("GET /api/v1/workspaces", a.listWorkspaces)
	mux.HandleFunc("POST /api/v1/workspaces", a.createWorkspace)
	mux.HandleFunc("GET /api/v1/workspaces/{workspaceID}/connections", a.listConnections)
	mux.HandleFunc("POST /api/v1/workspaces/{workspaceID}/connections", a.createConnection)
	mux.HandleFunc("POST /api/v1/workspaces/{workspaceID}/connections/{connectionID}/test", a.testConnection)
	mux.HandleFunc("GET /api/v1/workspaces/{workspaceID}/agents", a.listAgents)
}

func (a *API) testConnection(w http.ResponseWriter, r *http.Request) {
	if a.connectionTester == nil {
		http.Error(w, "connection testing is unavailable", http.StatusServiceUnavailable)
		return
	}
	check, err := a.connectionTester.CheckConnection(r.Context(), r.PathValue("workspaceID"), r.PathValue("connectionID"))
	if errors.Is(err, workspaces.ErrNotFound) {
		http.NotFound(w, r)
		return
	}
	if err != nil {
		http.Error(w, "provider connection failed: "+err.Error(), http.StatusBadGateway)
		return
	}
	respond(w, check)
}

func (a *API) listAgents(w http.ResponseWriter, r *http.Request) {
	if a.agents == nil {
		http.Error(w, "agent projections are unavailable", http.StatusServiceUnavailable)
		return
	}
	items, err := a.agents.ListAgents(r.Context(), r.PathValue("workspaceID"))
	if err != nil && len(items) == 0 {
		http.Error(w, "unable to load agent projections", http.StatusBadGateway)
		return
	}
	if err != nil {
		w.Header().Set("Warning", `199 forma "some agent connections could not be read"`)
	}
	respond(w, items)
}

func (a *API) listWorkspaces(w http.ResponseWriter, r *http.Request) {
	items, err := a.workspaces.ListWorkspaces(r.Context())
	if err != nil {
		http.Error(w, "unable to list workspaces", http.StatusInternalServerError)
		return
	}
	respond(w, items)
}

func (a *API) createWorkspace(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Name string `json:"name"`
	}
	if !decodeJSON(w, r, &request) {
		return
	}
	request.Name = strings.TrimSpace(request.Name)
	if request.Name == "" || len(request.Name) > 120 {
		http.Error(w, "name must contain between 1 and 120 bytes", http.StatusBadRequest)
		return
	}
	workspace, err := a.workspaces.CreateWorkspace(r.Context(), workspaces.Workspace{ID: newID(), Name: request.Name})
	if err != nil {
		http.Error(w, "unable to create workspace", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Location", "/api/v1/workspaces/"+workspace.ID)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(workspace)
}

func (a *API) listConnections(w http.ResponseWriter, r *http.Request) {
	items, err := a.workspaces.ListConnections(r.Context(), r.PathValue("workspaceID"))
	if err != nil {
		http.Error(w, "unable to list connections", http.StatusInternalServerError)
		return
	}
	respond(w, items)
}

func (a *API) createConnection(w http.ResponseWriter, r *http.Request) {
	var request struct {
		AdapterID    string `json:"adapterId"`
		ProviderType string `json:"providerType"`
		DisplayName  string `json:"displayName"`
		EndpointURL  string `json:"endpointUrl"`
		Credential   string `json:"credential"`
	}
	if !decodeJSON(w, r, &request) {
		return
	}
	request.AdapterID = strings.TrimSpace(request.AdapterID)
	request.ProviderType = strings.TrimSpace(request.ProviderType)
	request.DisplayName = strings.TrimSpace(request.DisplayName)
	request.EndpointURL = strings.TrimSpace(request.EndpointURL)
	if request.AdapterID == "" || len(request.AdapterID) > 120 || request.DisplayName == "" || len(request.DisplayName) > 120 || len(request.EndpointURL) > 2048 || len(request.Credential) > 16384 || !validProviderType(request.ProviderType) {
		http.Error(w, "invalid connection metadata", http.StatusBadRequest)
		return
	}
	connection, err := a.workspaces.CreateConnection(r.Context(), workspaces.Connection{ID: newID(), WorkspaceID: r.PathValue("workspaceID"), AdapterID: request.AdapterID, ProviderType: request.ProviderType, DisplayName: request.DisplayName, EndpointURL: request.EndpointURL, Credential: request.Credential})
	if errors.Is(err, workspaces.ErrNotFound) {
		http.NotFound(w, r)
		return
	}
	if err != nil {
		http.Error(w, "unable to create connection", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Location", "/api/v1/workspaces/"+connection.WorkspaceID+"/connections/"+connection.ID)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(connection)
}

func decodeJSON(w http.ResponseWriter, r *http.Request, destination any) bool {
	if !strings.HasPrefix(r.Header.Get("Content-Type"), "application/json") {
		http.Error(w, "expected application/json", http.StatusUnsupportedMediaType)
		return false
	}
	r.Body = http.MaxBytesReader(w, r.Body, 16*1024)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(destination); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return false
	}
	return true
}

func validProviderType(value string) bool {
	return value == "agent" || value == "project" || value == "chat"
}

func newID() string {
	var value [16]byte
	if _, err := rand.Read(value[:]); err != nil {
		panic(err)
	}
	value[6] = value[6]&0x0f | 0x40
	value[8] = value[8]&0x3f | 0x80
	encoded := hex.EncodeToString(value[:])
	return encoded[:8] + "-" + encoded[8:12] + "-" + encoded[12:16] + "-" + encoded[16:20] + "-" + encoded[20:]
}
