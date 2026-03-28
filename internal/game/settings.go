package game

import gamestorage "tankbattle/internal/storage"

func defaultSettings() userSettings {
	return gamestorage.DefaultSettings()
}

func settingsPath() string {
	return gamestorage.SettingsPath()
}

func historyPath() string {
	return gamestorage.HistoryPath()
}

func settingsDir() string {
	return gamestorage.SettingsDir()
}

func legacySettingsPath() string {
	return gamestorage.LegacySettingsPath()
}

func sanitizeScoreHistory(entries []scoreEntry) []scoreEntry {
	return gamestorage.SanitizeScoreHistory(entries)
}

func loadSettingsAt(path string) (userSettings, error) {
	return gamestorage.LoadSettingsAt(path)
}

func saveSettingsAt(path string, cfg userSettings) error {
	return gamestorage.SaveSettingsAt(path, cfg)
}

func loadHistoryAt(path string) ([]scoreEntry, error) {
	return gamestorage.LoadHistoryAt(path)
}

func saveHistoryAt(path string, entries []scoreEntry) error {
	return gamestorage.SaveHistoryAt(path, entries)
}

func (g *game) loadUserSettings() {
	if g == nil || !g.persistUserData {
		g.scoreHistory = nil
		g.rankScroll = 0
		if g.audio != nil {
			g.audio.SetEnabled(g.soundEnabled)
			g.audio.SetSFXVolume(float64(g.soundVolume) / 100.0)
		}
		return
	}
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
	g.loadUserHistory()
	if g.audio != nil {
		g.audio.SetEnabled(g.soundEnabled)
		g.audio.SetSFXVolume(float64(g.soundVolume) / 100.0)
	}
}

func (g *game) loadUserHistory() {
	if g == nil || !g.persistUserData {
		g.scoreHistory = nil
		g.rankScroll = 0
		return
	}
	entries, err := loadHistoryAt(historyPath())
	if err == nil {
		g.scoreHistory = entries
		g.rankScroll = 0
		return
	}
	g.scoreHistory = entries
	g.rankScroll = 0
}

func (g *game) saveUserSettings() {
	if g == nil || !g.persistUserData {
		return
	}
	_ = saveSettingsAt(settingsPath(), userSettings{
		SoundEnabled: g.soundEnabled,
		SoundVolume:  g.soundVolume,
	})
}

func (g *game) saveUserHistory() {
	if g == nil || !g.persistUserData {
		return
	}
	_ = saveHistoryAt(historyPath(), g.scoreHistory)
}
