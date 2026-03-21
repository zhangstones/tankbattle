package tankbattle

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type userSettings struct {
	SoundEnabled bool `json:"sound_enabled"`
	SoundVolume  int  `json:"sound_volume"`
}

func defaultSettings() userSettings {
	return userSettings{
		SoundEnabled: true,
		SoundVolume:  75,
	}
}

func settingsPath() string {
	return filepath.Join(".", "settings.json")
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
	return cfg, nil
}

func saveSettingsAt(path string, cfg userSettings) error {
	cfg.SoundVolume = clampInt(cfg.SoundVolume, 0, 100)
	raw, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, raw, 0o644)
}

func (g *game) loadUserSettings() {
	cfg, err := loadSettingsAt(settingsPath())
	if err != nil {
		cfg = defaultSettings()
	}
	g.soundEnabled = cfg.SoundEnabled
	g.soundVolume = cfg.SoundVolume
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
	})
}
