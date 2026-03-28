package debugapi

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

type State struct {
	GameState           string `json:"game_state"`
	MenuIndex           int    `json:"menu_index"`
	Difficulty          string `json:"difficulty"`
	TotalWaves          int    `json:"total_waves"`
	SoundEnabled        bool   `json:"sound_enabled"`
	SoundVolume         int    `json:"sound_volume"`
	Paused              bool   `json:"paused"`
	ShowHistory         bool   `json:"show_history"`
	Wave                int    `json:"wave"`
	MaxWave             int    `json:"max_wave"`
	Score               int    `json:"score"`
	EnemyCount          int    `json:"enemy_count"`
	Win                 bool   `json:"win"`
	MenuResumeAvailable bool   `json:"menu_resume_available"`
	MenuRequireRestart  bool   `json:"menu_require_restart"`
	Message             string `json:"message"`
}

type RequestKind string

const (
	RequestActions  RequestKind = "actions"
	RequestSnapshot RequestKind = "snapshot"
	RequestState    RequestKind = "state"
)

type Request struct {
	Kind    RequestKind
	Actions []string
	Dir     string
	Name    string
	resp    chan Response
}

type Response struct {
	Path  string
	State State
	Err   error
}

func (r Request) Reply(resp Response) {
	if r.resp != nil {
		r.resp <- resp
	}
}

type Controller struct {
	requests chan Request

	mu     sync.RWMutex
	server *http.Server
	addr   string
}

type debugActionsPayload struct {
	Actions []string `json:"actions"`
}

type debugSnapshotPayload struct {
	Dir  string `json:"dir"`
	Name string `json:"name"`
}

func New() *Controller {
	return &Controller{
		requests: make(chan Request, 16),
	}
}

func (c *Controller) ExecuteActions(actions ...string) error {
	if len(actions) == 0 {
		return nil
	}
	_, err := c.roundTrip(Request{
		Kind:    RequestActions,
		Actions: append([]string(nil), actions...),
	})
	return err
}

func (c *Controller) ExportSnapshot(dir, name string) (string, error) {
	resp, err := c.roundTrip(Request{
		Kind: RequestSnapshot,
		Dir:  dir,
		Name: name,
	})
	if err != nil {
		return "", err
	}
	return resp.Path, nil
}

func (c *Controller) State() (State, error) {
	resp, err := c.roundTrip(Request{Kind: RequestState})
	if err != nil {
		return State{}, err
	}
	return resp.State, nil
}

func (c *Controller) StartHTTP(addr string) error {
	if c == nil {
		return fmt.Errorf("debug controller is nil")
	}
	addr = strings.TrimSpace(addr)
	if addr == "" {
		return fmt.Errorf("debug api addr is empty")
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/debug/state", c.handleState)
	mux.HandleFunc("/debug/actions", c.handleActions)
	mux.HandleFunc("/debug/snapshot", c.handleSnapshot)

	server := &http.Server{
		Addr:    addr,
		Handler: mux,
	}
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	c.mu.Lock()
	c.server = server
	c.addr = ln.Addr().String()
	c.mu.Unlock()

	go func() {
		_ = server.Serve(ln)
	}()
	return nil
}

func (c *Controller) Close() error {
	if c == nil {
		return nil
	}
	c.mu.RLock()
	server := c.server
	c.mu.RUnlock()
	if server == nil {
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	return server.Shutdown(ctx)
}

func (c *Controller) Addr() string {
	if c == nil {
		return ""
	}
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.addr
}

func (c *Controller) Dequeue() (Request, bool) {
	if c == nil {
		return Request{}, false
	}
	select {
	case req := <-c.requests:
		return req, true
	default:
		return Request{}, false
	}
}

func (c *Controller) roundTrip(req Request) (Response, error) {
	if c == nil {
		return Response{}, fmt.Errorf("debug controller is nil")
	}
	req.resp = make(chan Response, 1)
	c.requests <- req
	resp := <-req.resp
	if resp.Err != nil {
		return Response{}, resp.Err
	}
	return resp, nil
}

func (c *Controller) handleState(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeDebugError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	state, err := c.State()
	if err != nil {
		writeDebugError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeDebugJSON(w, http.StatusOK, state)
}

func (c *Controller) handleActions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeDebugError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	var payload debugActionsPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeDebugError(w, http.StatusBadRequest, "invalid json body")
		return
	}
	if len(payload.Actions) == 0 {
		writeDebugError(w, http.StatusBadRequest, "actions are required")
		return
	}
	if err := c.ExecuteActions(payload.Actions...); err != nil {
		writeDebugError(w, http.StatusBadRequest, err.Error())
		return
	}
	state, err := c.State()
	if err != nil {
		writeDebugError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeDebugJSON(w, http.StatusOK, state)
}

func (c *Controller) handleSnapshot(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeDebugError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	var payload debugSnapshotPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeDebugError(w, http.StatusBadRequest, "invalid json body")
		return
	}
	path, err := c.ExportSnapshot(payload.Dir, payload.Name)
	if err != nil {
		writeDebugError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeDebugJSON(w, http.StatusOK, map[string]string{"path": path})
}

func writeDebugJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

func writeDebugError(w http.ResponseWriter, code int, msg string) {
	writeDebugJSON(w, code, map[string]string{"error": msg})
}
