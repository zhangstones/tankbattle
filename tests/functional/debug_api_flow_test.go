package functional_test

import (
	"testing"

	"tankbattle/testkit"
)

func TestDebugAPIMenuConfigurationFlow(t *testing.T) {
	session := testkit.StartSession(t, testkit.LaunchOptions{})

	state := mustActions(t, session, "scene.menu.default")
	assertState(t, state.GameState == "menu", "expected menu state, got %q", state.GameState)
	assertState(t, state.Difficulty == "normal", "expected normal difficulty, got %q", state.Difficulty)
	assertState(t, state.TotalWaves == 4, "expected 4 total waves, got %d", state.TotalWaves)
	assertState(t, state.SoundEnabled, "expected sound enabled by default")
	assertState(t, state.SoundVolume == 75, "expected default volume 75, got %d", state.SoundVolume)

	state = mustActions(t, session, "menu.hard")
	assertState(t, state.Difficulty == "hard", "expected hard difficulty, got %q", state.Difficulty)
	assertState(t, state.TotalWaves == 5, "expected hard preset to set 5 waves, got %d", state.TotalWaves)

	state = mustActions(t, session, "menu.down", "menu.down", "menu.right")
	assertState(t, state.MenuIndex == 2, "expected menu index 2 on sound row, got %d", state.MenuIndex)
	assertState(t, !state.SoundEnabled, "expected sound toggle to disable audio")

	state = mustActions(t, session, "menu.down", "menu.left")
	assertState(t, state.MenuIndex == 3, "expected menu index 3 on volume row, got %d", state.MenuIndex)
	assertState(t, state.SoundVolume == 50, "expected volume to decrease to 50, got %d", state.SoundVolume)
}

func TestDebugAPIPauseHistoryAndOutcomeFlow(t *testing.T) {
	session := testkit.StartSession(t, testkit.LaunchOptions{})

	state := mustActions(t, session, "scene.hud.progressed")
	assertState(t, state.GameState == "playing", "expected playing state, got %q", state.GameState)
	assertState(t, state.Wave == 3, "expected wave 3 scene, got %d", state.Wave)
	assertState(t, state.Score == 275, "expected score 275, got %d", state.Score)

	state = mustActions(t, session, "game.pause")
	assertState(t, state.Paused, "expected paused state after game.pause")

	state = mustActions(t, session, "game.resume")
	assertState(t, !state.Paused, "expected resumed state after game.resume")

	state = mustActions(t, session, "game.toggle_history")
	assertState(t, state.ShowHistory, "expected history panel to be visible")

	state = mustActions(t, session, "game.toggle_history")
	assertState(t, !state.ShowHistory, "expected history panel to be hidden")

	state = mustActions(t, session, "scene.victory")
	assertState(t, state.GameState == "ended", "expected ended state for victory, got %q", state.GameState)
	assertState(t, state.Win, "expected victory scene to report win=true")

	state = mustActions(t, session, "scene.defeat")
	assertState(t, state.GameState == "ended", "expected ended state for defeat, got %q", state.GameState)
	assertState(t, !state.Win, "expected defeat scene to report win=false")
}

func TestDebugAPIMenuResumeVsRestart(t *testing.T) {
	session := testkit.StartSession(t, testkit.LaunchOptions{})

	state := mustActions(t, session, "scene.hud.progressed", "game.enter_menu")
	assertState(t, state.GameState == "menu", "expected menu after entering config, got %q", state.GameState)
	assertState(t, state.MenuResumeAvailable, "expected menu resume to be available")
	assertState(t, !state.MenuRequireRestart, "unexpected restart requirement before edits")

	state = mustActions(t, session, "menu.down", "menu.down", "menu.right", "game.leave_menu")
	assertState(t, state.GameState == "playing", "expected to resume playing after audio-only edit, got %q", state.GameState)
	assertState(t, state.Wave == 3, "expected resumed run to keep wave 3, got %d", state.Wave)
	assertState(t, state.Score == 275, "expected resumed run to keep score 275, got %d", state.Score)
	assertState(t, !state.MenuResumeAvailable, "menu resume flag should be cleared after leaving menu")

	state = mustActions(t, session, "scene.hud.progressed", "game.enter_menu", "menu.hard", "game.leave_menu")
	assertState(t, state.GameState == "playing", "expected playing state after restart-required menu leave, got %q", state.GameState)
	assertState(t, state.Difficulty == "hard", "expected hard difficulty to persist after restart, got %q", state.Difficulty)
	assertState(t, state.Wave == 1, "expected restart-required leave to reset to wave 1, got %d", state.Wave)
	assertState(t, state.Score == 0, "expected restart-required leave to reset score, got %d", state.Score)
	assertState(t, !state.MenuResumeAvailable, "menu resume flag should be cleared after restart")
}

func mustActions(t *testing.T, session *testkit.Session, actions ...string) testkit.DebugState {
	t.Helper()
	state, err := session.Client.Actions(actions...)
	if err != nil {
		t.Fatalf("debug actions %v failed: %v", actions, err)
	}
	return state
}

func assertState(t *testing.T, cond bool, format string, args ...any) {
	t.Helper()
	if !cond {
		t.Fatalf(format, args...)
	}
}
