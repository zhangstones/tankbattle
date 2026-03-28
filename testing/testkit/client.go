package testkit

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
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

type Client struct {
	BaseURL    string
	HTTPClient *http.Client
}

func NewClient(baseURL string) *Client {
	return &Client{
		BaseURL: strings.TrimRight(baseURL, "/"),
		HTTPClient: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

func (c *Client) State() (DebugState, error) {
	var state DebugState
	if err := c.doJSON(http.MethodGet, "/debug/state", nil, &state); err != nil {
		return DebugState{}, err
	}
	return state, nil
}

func (c *Client) Actions(actions ...string) (DebugState, error) {
	var state DebugState
	payload := map[string][]string{"actions": actions}
	if err := c.doJSON(http.MethodPost, "/debug/actions", payload, &state); err != nil {
		return DebugState{}, err
	}
	return state, nil
}

func (c *Client) Snapshot(dir, name string) (string, error) {
	resp := struct {
		Path string `json:"path"`
	}{}
	payload := map[string]string{
		"dir":  dir,
		"name": name,
	}
	if err := c.doJSON(http.MethodPost, "/debug/snapshot", payload, &resp); err != nil {
		return "", err
	}
	return resp.Path, nil
}

func (c *Client) WaitForReady(ctx context.Context) error {
	for {
		if _, err := c.State(); err == nil {
			return nil
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			time.Sleep(100 * time.Millisecond)
		}
	}
}

func (c *Client) WaitForState(ctx context.Context, predicate func(DebugState) bool) (DebugState, error) {
	for {
		state, err := c.State()
		if err == nil && predicate(state) {
			return state, nil
		}
		select {
		case <-ctx.Done():
			if err != nil {
				return DebugState{}, err
			}
			return DebugState{}, ctx.Err()
		default:
			time.Sleep(100 * time.Millisecond)
		}
	}
}

func GoldenPath(rootDir, rel string) string {
	return filepath.Join(rootDir, "testing", "testdata", "golden", filepath.FromSlash(rel))
}

func UpdateGoldenEnabled() bool {
	return os.Getenv("TANKBATTLE_UPDATE_GOLDEN") == "1"
}

func (c *Client) doJSON(method, path string, payload any, out any) error {
	var bodyReader *bytes.Reader
	if payload == nil {
		bodyReader = bytes.NewReader(nil)
	} else {
		raw, err := json.Marshal(payload)
		if err != nil {
			return err
		}
		bodyReader = bytes.NewReader(raw)
	}
	req, err := http.NewRequest(method, c.BaseURL+path, bodyReader)
	if err != nil {
		return err
	}
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var msg map[string]string
		_ = json.NewDecoder(resp.Body).Decode(&msg)
		if errMsg := msg["error"]; errMsg != "" {
			return fmt.Errorf("debug api %s %s failed: %s", method, path, errMsg)
		}
		return fmt.Errorf("debug api %s %s failed: status=%d", method, path, resp.StatusCode)
	}
	if out == nil {
		return nil
	}
	return json.NewDecoder(resp.Body).Decode(out)
}
