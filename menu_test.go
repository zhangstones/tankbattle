package tankbattle

import "testing"

func TestMenuStartFromAnySelection(t *testing.T) {
	g := newGame()
	g.menuIndex = 0
	g.applyMenuAction(menuStart)
	if g.state != statePlaying {
		t.Fatalf("expected menuStart to enter playing state")
	}
}

func TestMenuDifficultyAndTotalWaveDefaults(t *testing.T) {
	g := newGame()
	g.menuIndex = 0
	g.difficulty = diffNormal
	g.applyMenuAction(menuInc)
	if g.difficulty != diffHard {
		t.Fatalf("expected difficulty to increase to hard")
	}
	if g.totalWaves != 5 {
		t.Fatalf("difficulty change should reset total waves to hard default 5, got %d", g.totalWaves)
	}
	g.applyMenuAction(menuInc)
	if g.difficulty != diffHard {
		t.Fatalf("difficulty should stay at hard upper bound")
	}

	if g.enemyBase != 4 {
		t.Fatalf("difficulty change should apply hard enemy base, got %d", g.enemyBase)
	}
}

func TestMenuTotalWaveBounds(t *testing.T) {
	g := newGame()
	g.menuIndex = 1
	g.totalWaves = matchWaveMin
	g.applyMenuAction(menuDec)
	if g.totalWaves != matchWaveMin {
		t.Fatalf("total waves should stay at lower bound %d", matchWaveMin)
	}
	g.totalWaves = matchWaveMax
	g.applyMenuAction(menuInc)
	if g.totalWaves != matchWaveMax {
		t.Fatalf("total waves should stay at upper bound %d", matchWaveMax)
	}
}

func TestMenuNavigationWrap(t *testing.T) {
	g := newGame()
	g.menuIndex = 0
	g.applyMenuAction(menuNavUp)
	if g.menuIndex != menuItemCount-1 {
		t.Fatalf("expected wrap up to %d, got %d", menuItemCount-1, g.menuIndex)
	}
	g.applyMenuAction(menuNavDown)
	if g.menuIndex != 0 {
		t.Fatalf("expected wrap down to 0, got %d", g.menuIndex)
	}
}

func TestApplyMenuSetDifficultyActions(t *testing.T) {
	g := newGame()
	g.applyMenuAction(menuSetEasy)
	if g.difficulty != diffEasy {
		t.Fatalf("menuSetEasy failed")
	}
	if g.totalWaves != 3 {
		t.Fatalf("easy should reset total waves to 3, got %d", g.totalWaves)
	}
	g.applyMenuAction(menuSetNormal)
	if g.difficulty != diffNormal {
		t.Fatalf("menuSetNormal failed")
	}
	if g.totalWaves != 4 {
		t.Fatalf("normal should reset total waves to 4, got %d", g.totalWaves)
	}
	g.applyMenuAction(menuSetHard)
	if g.difficulty != diffHard {
		t.Fatalf("menuSetHard failed")
	}
	if g.totalWaves != 5 {
		t.Fatalf("hard should reset total waves to 5, got %d", g.totalWaves)
	}
}

func TestMenuSoundToggle(t *testing.T) {
	g := newGame()
	g.soundEnabled = true
	if g.audio != nil {
		g.audio.SetEnabled(true)
	}
	g.menuIndex = 2
	g.applyMenuAction(menuInc)
	if g.soundEnabled {
		t.Fatalf("sound should toggle off")
	}
	if g.audio == nil || g.audio.Enabled() {
		t.Fatalf("audio manager should sync disabled state")
	}
	g.applyMenuAction(menuDec)
	if !g.soundEnabled {
		t.Fatalf("sound should toggle on")
	}
	if g.audio == nil || !g.audio.Enabled() {
		t.Fatalf("audio manager should sync enabled state")
	}
}

func TestMenuSoundVolumeBounds(t *testing.T) {
	g := newGame()
	g.menuIndex = 3
	g.soundVolume = 100
	g.applyMenuAction(menuInc)
	if g.soundVolume != 100 {
		t.Fatalf("volume should stay capped at 100, got %d", g.soundVolume)
	}
	g.soundVolume = 0
	g.applyMenuAction(menuDec)
	if g.soundVolume != 0 {
		t.Fatalf("volume should stay at lower bound 0, got %d", g.soundVolume)
	}
	g.soundVolume = 50
	g.applyMenuAction(menuInc)
	if g.soundVolume != 75 {
		t.Fatalf("volume should increase by 25, got %d", g.soundVolume)
	}
	g.applyMenuAction(menuDec)
	if g.soundVolume != 50 {
		t.Fatalf("volume should decrease by 25, got %d", g.soundVolume)
	}
}

func TestMenuBlockedSFXAtBounds(t *testing.T) {
	g := newGame()
	mock := &mockSFXPlayer{enabled: true}
	g.audio = mock

	g.menuIndex = 0
	g.difficulty = diffHard
	g.applyMenuAction(menuInc)
	if last, ok := mock.last(); !ok || last != sfxMenuBlocked {
		t.Fatalf("expected blocked sfx on difficulty upper bound, got %v (ok=%v)", last, ok)
	}

	g.menuIndex = 3
	g.soundVolume = 0
	g.applyMenuAction(menuDec)
	if last, ok := mock.last(); !ok || last != sfxMenuBlocked {
		t.Fatalf("expected blocked sfx on volume lower bound, got %v (ok=%v)", last, ok)
	}
}

func TestAudioMenuChangesDoNotRequireRestart(t *testing.T) {
	g := newPlayingGameForTest()
	g.enterMenuForConfig()
	g.menuIndex = 2
	g.applyMenuAction(menuInc)
	if g.menuRequireRestart {
		t.Fatalf("sound toggle should not require restart")
	}
	g.menuIndex = 3
	g.soundVolume = 50
	g.applyMenuAction(menuInc)
	if g.menuRequireRestart {
		t.Fatalf("sound volume change should not require restart")
	}
}

func TestDifficultyOrWaveMenuChangesRequireRestart(t *testing.T) {
	g := newPlayingGameForTest()
	g.enterMenuForConfig()
	g.menuIndex = 0
	g.difficulty = diffNormal
	g.applyMenuAction(menuInc)
	if !g.menuRequireRestart {
		t.Fatalf("difficulty change should require restart")
	}

	g2 := newPlayingGameForTest()
	g2.enterMenuForConfig()
	g2.menuIndex = 1
	g2.totalWaves = matchWaveMin + 1
	g2.applyMenuAction(menuDec)
	if !g2.menuRequireRestart {
		t.Fatalf("total waves change should require restart")
	}
}
