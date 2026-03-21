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
	initialPlayerHP := g.player.hp
	initialTurretHP := g.player.turretHP
	g.fort.hp = 0
	if err := g.Update(); err != nil {
		t.Fatalf("Update error: %v", err)
	}
	if g.state != stateEnded || g.win {
		t.Fatalf("expected defeat end state")
	}
	if g.fort.hp != 0 {
		t.Fatalf("fortress hp should clamp to zero on fortress defeat")
	}
	if g.player.hp != initialPlayerHP || g.player.turretHP != initialTurretHP {
		t.Fatalf("tank energy should remain unchanged on fortress-only defeat")
	}
}

func TestUpdateDefeatWhenPlayerDestroyed(t *testing.T) {
	g := newPlayingGameForTest()
	initialFortHP := g.fort.hp
	g.player.hp = 0
	if err := g.Update(); err != nil {
		t.Fatalf("Update error: %v", err)
	}
	if g.state != stateEnded || g.win {
		t.Fatalf("expected defeat end state")
	}
	if g.player.hp != 0 || g.player.turretHP != 0 {
		t.Fatalf("tank energy should clamp to zero on player defeat")
	}
	if g.fort.hp != initialFortHP {
		t.Fatalf("fortress hp should remain unchanged on player-only defeat")
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
	g.showHistory = true
	g.returnToMenu()
	if g.state != stateMenu {
		t.Fatalf("returnToMenu should switch state to menu")
	}
	if g.paused {
		t.Fatalf("returnToMenu should clear pause")
	}
	if g.showHistory {
		t.Fatalf("returnToMenu should hide history panel")
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

func TestToggleHistoryViewDoesNotShowMessage(t *testing.T) {
	g := newPlayingGameForTest()
	g.msg = "keep"
	g.msgTick = 12
	g.showHistory = false
	g.toggleHistoryView()
	if !g.showHistory {
		t.Fatalf("history panel should be enabled")
	}
	if g.msg != "keep" || g.msgTick != 12 {
		t.Fatalf("toggleHistoryView should not change message state")
	}
	g.toggleHistoryView()
	if g.showHistory {
		t.Fatalf("history panel should be disabled")
	}
	if g.msg != "keep" || g.msgTick != 12 {
		t.Fatalf("toggleHistoryView should not change message state when disabling")
	}
}

func TestUpdateInMenuAdvancesAudioFrame(t *testing.T) {
	g := newGame()
	g.state = stateMenu
	before := g.audioFrame
	if err := g.Update(); err != nil {
		t.Fatalf("Update error: %v", err)
	}
	if g.audioFrame <= before {
		t.Fatalf("audio frame should advance in menu update, before=%d after=%d", before, g.audioFrame)
	}
}

func TestStartMatchDoesNotResetAudioFrame(t *testing.T) {
	g := newGame()
	g.audioFrame = 123
	g.startMatch()
	if g.audioFrame != 123 {
		t.Fatalf("startMatch should not reset audio frame, got %d", g.audioFrame)
	}
}

func TestPlaySFXUsesAudioFrame(t *testing.T) {
	g := newGame()
	mock := &mockSFXPlayer{enabled: true}
	g.audio = mock
	g.audioFrame = 77
	g.playSFX(sfxMenuMove)
	if frame, ok := mock.lastFrame(); !ok || frame != 77 {
		t.Fatalf("playSFX should pass audio frame, got %d (ok=%v)", frame, ok)
	}
}
