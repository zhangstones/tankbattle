package game

import (
	"path/filepath"
	"testing"
	"time"
)

func TestSnapshotPathValidation(t *testing.T) {
	path, err := snapshotPath(t.TempDir(), "menu")
	if err != nil {
		t.Fatalf("snapshotPath should accept bare filename: %v", err)
	}
	if filepath.Ext(path) != ".png" {
		t.Fatalf("snapshotPath should append png extension, got %q", path)
	}
	if _, err := snapshotPath("", "menu.png"); err == nil {
		t.Fatalf("snapshotPath should reject empty dir")
	}
	if _, err := snapshotPath(t.TempDir(), "nested\\menu.png"); err == nil {
		t.Fatalf("snapshotPath should reject path separators in name")
	}
	if _, err := snapshotPath(t.TempDir(), "menu.jpg"); err == nil {
		t.Fatalf("snapshotPath should reject non-png extensions")
	}
}

func TestDebugActionSequenceUpdatesMenuState(t *testing.T) {
	seed := int64(42)
	g := newGameWithOptions(newGameOptions{
		loadUserSettings: false,
		persistUserData:  false,
		randomSeed:       &seed,
	})
	actions := []string{
		"menu.down",
		"menu.down",
		"menu.right",
		"menu.up",
		"menu.left",
	}
	for _, action := range actions {
		if err := g.executeDebugAction(action); err != nil {
			t.Fatalf("executeDebugAction(%q) failed: %v", action, err)
		}
	}
	if g.menuIndex != 1 {
		t.Fatalf("expected menu index 1 after action sequence, got %d", g.menuIndex)
	}
	if g.soundEnabled {
		t.Fatalf("expected sound to be toggled off by debug action")
	}
}

func TestDebugControllerProcessesRequests(t *testing.T) {
	controller := NewDebugController()
	seed := int64(7)
	g := newGameWithOptions(newGameOptions{
		loadUserSettings: false,
		persistUserData:  false,
		debug:            controller,
		randomSeed:       &seed,
	})

	errCh := make(chan error, 1)
	go func() {
		errCh <- controller.ExecuteActions("menu.down", "menu.right")
	}()

	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if err := g.processDebugRequests(); err != nil {
			t.Fatalf("processDebugRequests failed: %v", err)
		}
		select {
		case err := <-errCh:
			if err != nil {
				t.Fatalf("controller action roundtrip failed: %v", err)
			}
			if g.menuIndex != 1 {
				t.Fatalf("expected menu index 1, got %d", g.menuIndex)
			}
			if g.totalWaves != 5 {
				t.Fatalf("expected total waves to increase to 5, got %d", g.totalWaves)
			}
			return
		default:
			time.Sleep(10 * time.Millisecond)
		}
	}
	t.Fatalf("controller action roundtrip timed out")
}

func TestDebugStateReportsCurrentValues(t *testing.T) {
	seed := int64(9)
	g := newGameWithOptions(newGameOptions{
		loadUserSettings: false,
		persistUserData:  false,
		randomSeed:       &seed,
	})
	g.startMatch()
	g.paused = true
	g.score = 180
	g.msg = "Wave 1 incoming"
	state := g.debugState()
	if state.GameState != "playing" {
		t.Fatalf("expected playing debug state, got %q", state.GameState)
	}
	if state.Difficulty != "normal" {
		t.Fatalf("expected normal difficulty, got %q", state.Difficulty)
	}
	if !state.Paused {
		t.Fatalf("expected paused state")
	}
	if state.Score != 180 {
		t.Fatalf("expected score 180, got %d", state.Score)
	}
}

func TestDebugGameUsesDeterministicSeed(t *testing.T) {
	seed := int64(20260328)
	g1 := newGameWithOptions(newGameOptions{
		loadUserSettings: false,
		persistUserData:  false,
		randomSeed:       &seed,
	})
	g1.startMatch()

	g2 := newGameWithOptions(newGameOptions{
		loadUserSettings: false,
		persistUserData:  false,
		randomSeed:       &seed,
	})
	g2.startMatch()

	if len(g1.walls) != len(g2.walls) {
		t.Fatalf("deterministic seed should keep wall count stable: %d vs %d", len(g1.walls), len(g2.walls))
	}
	for i := range g1.walls {
		if g1.walls[i].box != g2.walls[i].box {
			t.Fatalf("deterministic seed should keep wall layout stable at %d: %+v vs %+v", i, g1.walls[i].box, g2.walls[i].box)
		}
	}
	if len(g1.enemies) != len(g2.enemies) {
		t.Fatalf("deterministic seed should keep enemy count stable: %d vs %d", len(g1.enemies), len(g2.enemies))
	}
	for i := range g1.enemies {
		if g1.enemies[i].x != g2.enemies[i].x || g1.enemies[i].y != g2.enemies[i].y || g1.enemies[i].dir != g2.enemies[i].dir {
			t.Fatalf("deterministic seed should keep spawn stable at %d", i)
		}
	}
}

func TestUnsupportedDebugActionFails(t *testing.T) {
	g := newGameWithOptions(newGameOptions{
		loadUserSettings: false,
		persistUserData:  false,
	})
	if err := g.executeDebugAction("menu.nope"); err == nil {
		t.Fatalf("unsupported debug action should fail")
	}
}

func TestDebugScenesProvideStableFunctionalStates(t *testing.T) {
	seed := int64(20260328)
	g := newGameWithOptions(newGameOptions{
		loadUserSettings: false,
		persistUserData:  false,
		randomSeed:       &seed,
	})

	if err := g.executeDebugAction("scene.hud.progressed"); err != nil {
		t.Fatalf("scene.hud.progressed failed: %v", err)
	}
	if g.state != statePlaying || g.wave != 3 || g.score != 275 {
		t.Fatalf("progressed scene mismatch: state=%v wave=%d score=%d", g.state, g.wave, g.score)
	}

	if err := g.executeDebugAction("scene.hud.shield"); err != nil {
		t.Fatalf("scene.hud.shield failed: %v", err)
	}
	if g.state != statePlaying || g.shieldTick == 0 || g.rapidTick == 0 {
		t.Fatalf("shield scene should enable active buffs: state=%v shield=%d rapid=%d", g.state, g.shieldTick, g.rapidTick)
	}

	if err := g.executeDebugAction("scene.victory"); err != nil {
		t.Fatalf("scene.victory failed: %v", err)
	}
	if g.state != stateEnded || !g.win {
		t.Fatalf("victory scene mismatch: state=%v win=%v", g.state, g.win)
	}
}
