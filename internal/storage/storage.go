package storage

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"time"
)

const ScoreHistoryLimit = 100

type ScoreEntry struct {
	Score       int    `json:"score"`
	At          string `json:"at"`
	DurationSec int    `json:"duration_sec"`
}

type UserSettings struct {
	SoundEnabled bool `json:"sound_enabled"`
	SoundVolume  int  `json:"sound_volume"`
}

func DefaultSettings() UserSettings {
	return UserSettings{
		SoundEnabled: true,
		SoundVolume:  75,
	}
}

func SettingsPath() string {
	return filepath.Join(SettingsDir(), "settings.json")
}

func HistoryPath() string {
	return filepath.Join(SettingsDir(), "history.json")
}

func SettingsDir() string {
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		return filepath.Join(".", ".tankbattle")
	}
	return filepath.Join(home, ".tankbattle")
}

func LegacySettingsPath() string {
	return filepath.Join(".", "settings.json")
}

func SanitizeScoreHistory(entries []ScoreEntry) []ScoreEntry {
	clean := make([]ScoreEntry, 0, len(entries))
	for _, e := range entries {
		if e.Score < 0 {
			continue
		}
		if e.DurationSec < 0 {
			e.DurationSec = 0
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
	if len(clean) > ScoreHistoryLimit {
		clean = clean[:ScoreHistoryLimit]
	}
	return clean
}

func LoadSettingsAt(path string) (UserSettings, error) {
	cfg := DefaultSettings()
	raw, err := os.ReadFile(path)
	if err != nil {
		return cfg, err
	}
	if err := json.Unmarshal(raw, &cfg); err != nil {
		return DefaultSettings(), err
	}
	cfg.SoundVolume = clampInt(cfg.SoundVolume, 0, 100)
	return cfg, nil
}

func SaveSettingsAt(path string, cfg UserSettings) error {
	cfg.SoundVolume = clampInt(cfg.SoundVolume, 0, 100)
	raw, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, raw, 0o644)
}

func LoadHistoryAt(path string) ([]ScoreEntry, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var entries []ScoreEntry
	if err := json.Unmarshal(raw, &entries); err != nil {
		return nil, err
	}
	return SanitizeScoreHistory(entries), nil
}

func SaveHistoryAt(path string, entries []ScoreEntry) error {
	entries = SanitizeScoreHistory(entries)
	raw, err := json.MarshalIndent(entries, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, raw, 0o644)
}

func clampInt(v, min, max int) int {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}
