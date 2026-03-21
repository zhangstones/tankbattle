package tankbattle

import (
	"path/filepath"
	"testing"
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
