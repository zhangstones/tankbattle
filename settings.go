package tankbattle

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"time"
)

type scoreEntry struct {
	Score int    `json:"score"`
	At    string `json:"at"`
}

type userSettings struct {
	SoundEnabled bool         `json:"sound_enabled"`
	SoundVolume  int          `json:"sound_volume"`
	ScoreHistory []scoreEntry `json:"score_history,omitempty"`
}

func defaultSettings() userSettings {
	return userSettings{
		SoundEnabled: true,
		SoundVolume:  75,
	}
}

func settingsPath() string {
	return filepath.Join(settingsDir(), "settings.json")
}

func settingsDir() string {
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		return filepath.Join(".", ".tankbattle")
	}
	return filepath.Join(home, ".tankbattle")
}

func legacySettingsPath() string {
	return filepath.Join(".", "settings.json")
}

func sanitizeScoreHistory(entries []scoreEntry) []scoreEntry {
	clean := make([]scoreEntry, 0, len(entries))
	for _, e := range entries {
		if e.Score < 0 {
			continue
		}
		if _, err := time.Parse(time.RFC3339, e.At); err != nil {
			e.At = ""
		}
		clean = append(clean, e)
	}
	sort.SliceStable(clean, func(i, j int) bool {
		if clean[i].Score == clean[j].Score {
			return clean[i].At > clean[j].At
		}
		return clean[i].Score > clean[j].Score
	})
	if len(clean) > scoreHistoryLimit {
		clean = clean[:scoreHistoryLimit]
	}
	return clean
}

func loadSettingsAt(path string) (userSettings, error) {
	cfg := defaultSettings()
	raw, err := os.ReadFile(path)
	if err != nil {
		return cfg, err
	}
	if err := json.Unmarshal(raw, &cfg); err != nil {
		return defaultSettings(), err
	}
	cfg.SoundVolume = clampInt(cfg.SoundVolume, 0, 100)
	cfg.ScoreHistory = sanitizeScoreHistory(cfg.ScoreHistory)
	return cfg, nil
}

func saveSettingsAt(path string, cfg userSettings) error {
	cfg.SoundVolume = clampInt(cfg.SoundVolume, 0, 100)
	cfg.ScoreHistory = sanitizeScoreHistory(cfg.ScoreHistory)
	raw, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, raw, 0o644)
}

func (g *game) loadUserSettings() {
	cfg, err := loadSettingsAt(settingsPath())
	if err != nil {
		legacy, legacyErr := loadSettingsAt(legacySettingsPath())
		if legacyErr == nil {
			cfg = legacy
			_ = saveSettingsAt(settingsPath(), cfg)
		} else {
			cfg = defaultSettings()
		}
	}
	g.soundEnabled = cfg.SoundEnabled
	g.soundVolume = cfg.SoundVolume
	g.scoreHistory = sanitizeScoreHistory(cfg.ScoreHistory)
	g.rankScroll = 0
	if g.audio != nil {
		g.audio.SetEnabled(g.soundEnabled)
		g.audio.SetSFXVolume(float64(g.soundVolume) / 100.0)
	}
}

func (g *game) saveUserSettings() {
	if g == nil {
		return
	}
	_ = saveSettingsAt(settingsPath(), userSettings{
		SoundEnabled: g.soundEnabled,
		SoundVolume:  g.soundVolume,
		ScoreHistory: g.scoreHistory,
	})
}
