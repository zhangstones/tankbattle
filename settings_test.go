package tankbattle

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestSettingsRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "settings.json")
	in := userSettings{
		SoundEnabled: false,
		SoundVolume:  55,
	}
	if err := saveSettingsAt(path, in); err != nil {
		t.Fatalf("save settings failed: %v", err)
	}
	out, err := loadSettingsAt(path)
	if err != nil {
		t.Fatalf("load settings failed: %v", err)
	}
	if out.SoundEnabled != in.SoundEnabled {
		t.Fatalf("sound enabled mismatch: got %v want %v", out.SoundEnabled, in.SoundEnabled)
	}
	if out.SoundVolume != in.SoundVolume {
		t.Fatalf("sound volume mismatch: got %d want %d", out.SoundVolume, in.SoundVolume)
	}
}

func TestSettingsClampVolume(t *testing.T) {
	path := filepath.Join(t.TempDir(), "settings.json")
	if err := saveSettingsAt(path, userSettings{SoundEnabled: true, SoundVolume: 200}); err != nil {
		t.Fatalf("save settings failed: %v", err)
	}
	out, err := loadSettingsAt(path)
	if err != nil {
		t.Fatalf("load settings failed: %v", err)
	}
	if out.SoundVolume != 100 {
		t.Fatalf("volume should be clamped to 100, got %d", out.SoundVolume)
	}
}

func TestSettingsPathUsesTankbattleHomeDir(t *testing.T) {
	got := filepath.Clean(settingsPath())
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		if !strings.Contains(got, filepath.Join(".tankbattle", "settings.json")) {
			t.Fatalf("settings path should use .tankbattle dir, got %q", got)
		}
		return
	}
	wantSuffix := filepath.Join(home, ".tankbattle", "settings.json")
	if got != filepath.Clean(wantSuffix) {
		t.Fatalf("settings path mismatch: got %q want %q", got, wantSuffix)
	}
}

func TestSanitizeScoreHistorySortedAndLimited(t *testing.T) {
	entries := make([]scoreEntry, 0, scoreHistoryLimit+5)
	now := time.Now().UTC()
	for i := 0; i < scoreHistoryLimit+5; i++ {
		entries = append(entries, scoreEntry{
			Score: i,
			At:    now.Add(time.Duration(i) * time.Minute).Format(time.RFC3339),
		})
	}
	entries = append(entries, scoreEntry{Score: -1, At: "bad"})
	got := sanitizeScoreHistory(entries)
	if len(got) != scoreHistoryLimit {
		t.Fatalf("history should be truncated to %d, got %d", scoreHistoryLimit, len(got))
	}
	if got[0].Score != scoreHistoryLimit+4 {
		t.Fatalf("history should be sorted desc by score, top=%d", got[0].Score)
	}
	for i, e := range got {
		if e.Score < 0 {
			t.Fatalf("negative score should be removed at index %d", i)
		}
		if e.At != "" {
			if _, err := time.Parse(time.RFC3339, e.At); err != nil {
				t.Fatalf("invalid timestamp at index %d: %s (%v)", i, fmt.Sprintf("%q", e.At), err)
			}
		}
	}
}
