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

func TestMenuDifficultyAndEnemyBaseBounds(t *testing.T) {
	g := newGame()
	g.menuIndex = 0
	g.difficulty = diffNormal
	g.applyMenuAction(menuInc)
	if g.difficulty != diffHard {
		t.Fatalf("expected difficulty to increase to hard")
	}
	g.applyMenuAction(menuInc)
	if g.difficulty != diffHard {
		t.Fatalf("difficulty should stay at hard upper bound")
	}

	g.menuIndex = 1
	g.enemyBase = enemyBaseMin
	g.applyMenuAction(menuDec)
	if g.enemyBase != enemyBaseMin {
		t.Fatalf("enemyBase should stay at lower bound %d", enemyBaseMin)
	}
	g.enemyBase = enemyBaseMax
	g.applyMenuAction(menuInc)
	if g.enemyBase != enemyBaseMax {
		t.Fatalf("enemyBase should stay at upper bound %d", enemyBaseMax)
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
	g.applyMenuAction(menuSetNormal)
	if g.difficulty != diffNormal {
		t.Fatalf("menuSetNormal failed")
	}
	g.applyMenuAction(menuSetHard)
	if g.difficulty != diffHard {
		t.Fatalf("menuSetHard failed")
	}
}

func TestApplyMenuEnemyShortcutBounds(t *testing.T) {
	g := newGame()
	g.enemyBase = enemyBaseMin
	g.applyMenuAction(menuEnemyDown)
	if g.enemyBase != enemyBaseMin {
		t.Fatalf("enemy lower bound broken")
	}
	g.enemyBase = enemyBaseMax
	g.applyMenuAction(menuEnemyUp)
	if g.enemyBase != enemyBaseMax {
		t.Fatalf("enemy upper bound broken")
	}
}

func TestMenuSoundToggle(t *testing.T) {
	g := newGame()
	if !g.soundEnabled {
		t.Fatalf("sound should be enabled by default")
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
