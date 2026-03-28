package ui_test

import (
	"path/filepath"
	"testing"

	"tankbattle/testing/testkit"
)

func TestDebugUISnapshots(t *testing.T) {
	session := testkit.StartSession(t, testkit.LaunchOptions{})

	cases := []struct {
		name      string
		scene     string
		goldenRel string
	}{
		{name: "menu-default", scene: "scene.menu.default", goldenRel: "menu/menu-default.png"},
		{name: "menu-hard", scene: "scene.menu.hard", goldenRel: "menu/menu-hard.png"},
		{name: "menu-resume", scene: "scene.menu.resume", goldenRel: "menu/menu-resume.png"},
		{name: "hud-playing", scene: "scene.hud.playing", goldenRel: "hud/hud-playing.png"},
		{name: "hud-shield", scene: "scene.hud.shield", goldenRel: "hud/hud-shield.png"},
		{name: "hud-history", scene: "scene.hud.history", goldenRel: "hud/hud-history.png"},
		{name: "pause-panel", scene: "scene.pause", goldenRel: "panels/pause-panel.png"},
		{name: "victory-panel", scene: "scene.victory", goldenRel: "panels/victory-panel.png"},
		{name: "defeat-panel", scene: "scene.defeat", goldenRel: "panels/defeat-panel.png"},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			if _, err := session.Client.Actions(tc.scene); err != nil {
				t.Fatalf("set scene %q: %v", tc.scene, err)
			}
			actualPath, err := session.Client.Snapshot(session.SnapshotDir, tc.name)
			if err != nil {
				t.Fatalf("export snapshot %q: %v", tc.name, err)
			}
			if filepath.Ext(actualPath) != ".png" {
				t.Fatalf("snapshot path should use png extension, got %q", actualPath)
			}
			testkit.AssertMatchesGolden(t, session.RootDir, actualPath, tc.goldenRel)
		})
	}
}
