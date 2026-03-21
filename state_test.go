package tankbattle

import "testing"

func TestFunctionalWaveProgressionAndSpawn(t *testing.T) {
	g := newGame()
	g.startMatch()
	g.maxWave = 3
	g.wave = 1
	g.enemies = nil
	g.waveDelay = 0

	if err := g.Update(); err != nil {
		t.Fatalf("Update error: %v", err)
	}
	if g.wave != 2 || g.waveDelay == 0 {
		t.Fatalf("expected prepare next wave, got wave=%d delay=%d", g.wave, g.waveDelay)
	}

	for g.waveDelay > 0 {
		g.enemies = nil
		_ = g.Update()
	}
	if len(g.enemies) == 0 {
		t.Fatalf("expected next wave enemies spawned")
	}
}

func TestFunctionalVictoryTransition(t *testing.T) {
	g := newGame()
	g.startMatch()
	g.wave = g.maxWave
	g.enemies = nil
	g.waveDelay = 0

	if err := g.Update(); err != nil {
		t.Fatalf("Update error: %v", err)
	}
	if g.state != stateEnded || !g.win {
		t.Fatalf("expected victory end state")
	}
}

func TestUpdateDefeatWhenFortressDestroyed(t *testing.T) {
	g := newPlayingGameForTest()
	g.fort.hp = 0
	if err := g.Update(); err != nil {
		t.Fatalf("Update error: %v", err)
	}
	if g.state != stateEnded || g.win {
		t.Fatalf("expected defeat end state")
	}
	if g.fort.hp != 0 || g.player.hp != 0 || g.player.turretHP != 0 {
		t.Fatalf("defeat should clamp fortress and tank energies to zero")
	}
}

func TestUpdateDefeatWhenPlayerDestroyed(t *testing.T) {
	g := newPlayingGameForTest()
	g.player.hp = 0
	if err := g.Update(); err != nil {
		t.Fatalf("Update error: %v", err)
	}
	if g.state != stateEnded || g.win {
		t.Fatalf("expected defeat end state")
	}
	if g.fort.hp != 0 || g.player.hp != 0 || g.player.turretHP != 0 {
		t.Fatalf("defeat should clamp fortress and tank energies to zero")
	}
}

func TestUpdateWaveDelayDoesNotSpawnEarly(t *testing.T) {
	g := newPlayingGameForTest()
	g.enemies = nil
	g.wave = 1
	g.waveDelay = 2
	_ = g.Update()
	if len(g.enemies) != 0 {
		t.Fatalf("should not spawn while waveDelay > 0")
	}
}

func TestRestartIfAllowedOnlyOutsideMenu(t *testing.T) {
	g := newGame()
	g.state = stateMenu
	if g.restartIfAllowed() {
		t.Fatalf("restart should be blocked in menu state")
	}

	g.state = stateEnded
	g.score = 200
	if !g.restartIfAllowed() {
		t.Fatalf("restart should be allowed outside menu state")
	}
	if g.state != statePlaying || g.score != 0 || g.wave != 1 {
		t.Fatalf("restart should reset state and start match")
	}
}

func TestReturnToMenuClearsPause(t *testing.T) {
	g := newPlayingGameForTest()
	g.paused = true
	g.returnToMenu()
	if g.state != stateMenu {
		t.Fatalf("returnToMenu should switch state to menu")
	}
	if g.paused {
		t.Fatalf("returnToMenu should clear pause")
	}
}

func TestTogglePauseSetsExpectedMessage(t *testing.T) {
	g := newPlayingGameForTest()
	g.togglePause()
	if !g.paused || g.msg != "Paused" {
		t.Fatalf("togglePause should pause and set paused message")
	}
	g.togglePause()
	if g.paused || g.msg != "Resume" {
		t.Fatalf("togglePause should resume and set resume message")
	}
}
