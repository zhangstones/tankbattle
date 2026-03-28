package tankbattle

import (
	"context"
	"encoding/json"
	"fmt"
	"image"
	"image/png"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
)

type DebugState struct {
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

type DebugController struct {
	requests chan debugRequest

	mu     sync.RWMutex
	server *http.Server
	addr   string
}

type debugRequestKind string

const (
	debugRequestActions  debugRequestKind = "actions"
	debugRequestSnapshot debugRequestKind = "snapshot"
	debugRequestState    debugRequestKind = "state"
)

type debugRequest struct {
	kind    debugRequestKind
	actions []string
	dir     string
	name    string
	resp    chan debugResponse
}

type debugResponse struct {
	path  string
	state DebugState
	err   error
}

type debugActionsPayload struct {
	Actions []string `json:"actions"`
}

type debugSnapshotPayload struct {
	Dir  string `json:"dir"`
	Name string `json:"name"`
}

func NewDebugController() *DebugController {
	return &DebugController{
		requests: make(chan debugRequest, 16),
	}
}

func (c *DebugController) ExecuteActions(actions ...string) error {
	if len(actions) == 0 {
		return nil
	}
	_, err := c.roundTrip(debugRequest{
		kind:    debugRequestActions,
		actions: append([]string(nil), actions...),
	})
	return err
}

func (c *DebugController) ExportSnapshot(dir, name string) (string, error) {
	resp, err := c.roundTrip(debugRequest{
		kind: debugRequestSnapshot,
		dir:  dir,
		name: name,
	})
	if err != nil {
		return "", err
	}
	return resp.path, nil
}

func (c *DebugController) State() (DebugState, error) {
	resp, err := c.roundTrip(debugRequest{kind: debugRequestState})
	if err != nil {
		return DebugState{}, err
	}
	return resp.state, nil
}

func (c *DebugController) StartHTTP(addr string) error {
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

func (c *DebugController) Close() error {
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

func (c *DebugController) Addr() string {
	if c == nil {
		return ""
	}
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.addr
}

func (c *DebugController) roundTrip(req debugRequest) (debugResponse, error) {
	if c == nil {
		return debugResponse{}, fmt.Errorf("debug controller is nil")
	}
	req.resp = make(chan debugResponse, 1)
	c.requests <- req
	resp := <-req.resp
	if resp.err != nil {
		return debugResponse{}, resp.err
	}
	return resp, nil
}

func (c *DebugController) handleState(w http.ResponseWriter, r *http.Request) {
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

func (c *DebugController) handleActions(w http.ResponseWriter, r *http.Request) {
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

func (c *DebugController) handleSnapshot(w http.ResponseWriter, r *http.Request) {
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

func (g *game) processDebugRequests() error {
	if g == nil || g.debug == nil {
		return nil
	}
	for {
		select {
		case req := <-g.debug.requests:
			req.resp <- g.handleDebugRequest(req)
		default:
			return nil
		}
	}
}

func (g *game) handleDebugRequest(req debugRequest) debugResponse {
	switch req.kind {
	case debugRequestActions:
		for _, action := range req.actions {
			if err := g.executeDebugAction(action); err != nil {
				return debugResponse{err: err}
			}
		}
		return debugResponse{}
	case debugRequestSnapshot:
		path, err := g.exportSnapshot(req.dir, req.name)
		if err != nil {
			return debugResponse{err: err}
		}
		return debugResponse{path: path}
	case debugRequestState:
		return debugResponse{state: g.debugState()}
	default:
		return debugResponse{err: fmt.Errorf("unsupported debug request kind: %s", req.kind)}
	}
}

func (g *game) executeDebugAction(action string) error {
	action = strings.ToLower(strings.TrimSpace(action))
	if action != "" && !strings.HasPrefix(action, "scene.") {
		g.debugFreeze = false
	}
	switch action {
	case "menu.up":
		return g.executeMenuDebugAction(menuNavUp)
	case "menu.down":
		return g.executeMenuDebugAction(menuNavDown)
	case "menu.left", "menu.decrease":
		return g.executeMenuDebugAction(menuDec)
	case "menu.right", "menu.increase":
		return g.executeMenuDebugAction(menuInc)
	case "menu.start":
		return g.executeMenuDebugAction(menuStart)
	case "menu.easy", "menu.set_easy":
		return g.executeMenuDebugAction(menuSetEasy)
	case "menu.normal", "menu.set_normal":
		return g.executeMenuDebugAction(menuSetNormal)
	case "menu.hard", "menu.set_hard":
		return g.executeMenuDebugAction(menuSetHard)
	case "game.enter_menu":
		if g.state != stateMenu {
			g.enterMenuForConfig()
		}
		return nil
	case "game.leave_menu":
		if g.state != stateMenu {
			return fmt.Errorf("game.leave_menu requires menu state")
		}
		g.leaveMenuByToggle()
		return nil
	case "game.start_match":
		g.startMatch()
		return nil
	case "game.restart":
		if !g.restartIfAllowed() {
			return fmt.Errorf("game.restart is not available in menu state")
		}
		return nil
	case "game.pause":
		if g.state != statePlaying || g.paused {
			return fmt.Errorf("game.pause requires active playing state")
		}
		g.togglePause()
		return nil
	case "game.resume":
		if g.state != statePlaying || !g.paused {
			return fmt.Errorf("game.resume requires paused playing state")
		}
		g.togglePause()
		return nil
	case "game.toggle_history":
		g.toggleHistoryView()
		return nil
	case "scene.menu.default":
		g.setDebugScene("menu.default")
		return nil
	case "scene.menu.hard":
		g.setDebugScene("menu.hard")
		return nil
	case "scene.menu.resume":
		g.setDebugScene("menu.resume")
		return nil
	case "scene.hud.playing":
		g.setDebugScene("hud.playing")
		return nil
	case "scene.hud.progressed":
		g.setDebugScene("hud.progressed")
		return nil
	case "scene.hud.shield":
		g.setDebugScene("hud.shield")
		return nil
	case "scene.hud.history":
		g.setDebugScene("hud.history")
		return nil
	case "scene.pause":
		g.setDebugScene("pause")
		return nil
	case "scene.victory":
		g.setDebugScene("victory")
		return nil
	case "scene.defeat":
		g.setDebugScene("defeat")
		return nil
	default:
		return fmt.Errorf("unsupported debug action: %s", action)
	}
}

func (g *game) executeMenuDebugAction(action menuAction) error {
	if g.state != stateMenu {
		return fmt.Errorf("menu action %v requires menu state", action)
	}
	g.applyMenuAction(action)
	return nil
}

func (g *game) exportSnapshot(dir, name string) (string, error) {
	fullPath, err := snapshotPath(dir, name)
	if err != nil {
		return "", err
	}
	img := ebiten.NewImage(screenW, screenH)
	g.Draw(img)

	pixels := make([]byte, 4*screenW*screenH)
	img.ReadPixels(pixels)
	rgba := image.NewRGBA(image.Rect(0, 0, screenW, screenH))
	copy(rgba.Pix, pixels)

	if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
		return "", err
	}
	f, err := os.Create(fullPath)
	if err != nil {
		return "", err
	}
	defer f.Close()
	if err := png.Encode(f, rgba); err != nil {
		return "", err
	}
	return fullPath, nil
}

func snapshotPath(dir, name string) (string, error) {
	dir = strings.TrimSpace(dir)
	name = strings.TrimSpace(name)
	if dir == "" {
		return "", fmt.Errorf("snapshot dir is required")
	}
	if name == "" {
		return "", fmt.Errorf("snapshot name is required")
	}
	if filepath.Base(name) != name {
		return "", fmt.Errorf("snapshot name must not contain path separators")
	}
	ext := strings.ToLower(filepath.Ext(name))
	if ext == "" {
		name += ".png"
	} else if ext != ".png" {
		return "", fmt.Errorf("snapshot name must use .png extension")
	}
	return filepath.Join(dir, name), nil
}

func (g *game) debugState() DebugState {
	if g == nil {
		return DebugState{}
	}
	return DebugState{
		GameState:           debugStateName(g.state),
		MenuIndex:           g.menuIndex,
		Difficulty:          debugDifficultyName(g.difficulty),
		TotalWaves:          g.totalWaves,
		SoundEnabled:        g.soundEnabled,
		SoundVolume:         g.soundVolume,
		Paused:              g.paused,
		ShowHistory:         g.showHistory,
		Wave:                g.wave,
		MaxWave:             g.maxWave,
		Score:               g.score,
		EnemyCount:          len(g.enemies),
		Win:                 g.win,
		MenuResumeAvailable: g.menuResumeAvailable,
		MenuRequireRestart:  g.menuRequireRestart,
		Message:             g.msg,
	}
}

func (g *game) setDebugScene(name string) {
	switch name {
	case "menu.default":
		g.resetDebugScene()
	case "menu.hard":
		g.resetDebugScene()
		g.difficulty = diffHard
		g.totalWaves = g.maxWaveByDifficulty()
		g.enemyBase = g.enemyBaseByDifficulty()
	case "menu.resume":
		g.resetDebugScene()
		g.startMatch()
		g.enterMenuForConfig()
	case "hud.playing":
		g.resetDebugScene()
		g.startMatch()
		g.msg = ""
		g.msgTick = 0
	case "hud.progressed":
		g.resetDebugScene()
		g.startMatch()
		g.msg = ""
		g.msgTick = 0
		g.wave = 3
		g.score = 275
		g.enemyKills = 7
	case "hud.shield":
		g.resetDebugScene()
		g.startMatch()
		g.msg = ""
		g.msgTick = 0
		g.wave = 2
		g.score = 480
		g.enemyKills = 11
		g.shieldTick = 300
		g.rapidTick = 180
		g.player.hp = 4
		g.player.turretHP = 3
	case "hud.history":
		g.resetDebugScene()
		g.startMatch()
		g.msg = ""
		g.msgTick = 0
		g.score = 360
		g.showHistory = true
		g.scoreHistory = debugHistoryEntries()
	case "pause":
		g.resetDebugScene()
		g.startMatch()
		g.msg = ""
		g.msgTick = 0
		g.paused = true
	case "victory":
		g.resetDebugScene()
		g.startMatch()
		g.state = stateEnded
		g.win = true
		g.score = 640
		g.msg = ""
		g.msgTick = 0
	case "defeat":
		g.resetDebugScene()
		g.startMatch()
		g.state = stateEnded
		g.win = false
		g.player.hp = 0
		g.fort.hp = 0
		g.msg = ""
		g.msgTick = 0
	}
	g.debugFreeze = true
}

func (g *game) resetDebugScene() {
	if g == nil {
		return
	}
	seed := int64(20260328)
	fresh := newGameWithOptions(newGameOptions{
		loadUserSettings: false,
		persistUserData:  false,
		audio:            g.audio,
		debug:            g.debug,
		randomSeed:       &seed,
	})
	*g = *fresh
}

func debugHistoryEntries() []scoreEntry {
	return []scoreEntry{
		{Score: 900, At: "2026-03-22T10:00:00Z", DurationSec: 420},
		{Score: 750, At: "2026-03-21T10:00:00Z", DurationSec: 398},
		{Score: 630, At: "2026-03-20T10:00:00Z", DurationSec: 365},
		{Score: 540, At: "2026-03-19T10:00:00Z", DurationSec: 344},
		{Score: 460, At: "2026-03-18T10:00:00Z", DurationSec: 321},
		{Score: 420, At: "2026-03-17T10:00:00Z", DurationSec: 302},
		{Score: 390, At: "2026-03-16T10:00:00Z", DurationSec: 288},
		{Score: 370, At: "2026-03-15T10:00:00Z", DurationSec: 276},
		{Score: 355, At: "2026-03-14T10:00:00Z", DurationSec: 264},
		{Score: 340, At: "2026-03-13T10:00:00Z", DurationSec: 251},
	}
}

func debugStateName(state gameState) string {
	switch state {
	case stateMenu:
		return "menu"
	case statePlaying:
		return "playing"
	case stateEnded:
		return "ended"
	default:
		return "unknown"
	}
}

func debugDifficultyName(d difficulty) string {
	switch d {
	case diffEasy:
		return "easy"
	case diffHard:
		return "hard"
	default:
		return "normal"
	}
}
