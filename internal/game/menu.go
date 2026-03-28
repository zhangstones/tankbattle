package game

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

type menuAction int

const (
	menuNavUp menuAction = iota
	menuNavDown
	menuDec
	menuInc
	menuStart
	menuSetEasy
	menuSetNormal
	menuSetHard
)

func (g *game) updateMenu() {
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowUp) || inpututil.IsKeyJustPressed(ebiten.KeyW) {
		g.applyMenuAction(menuNavUp)
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowDown) || inpututil.IsKeyJustPressed(ebiten.KeyS) {
		g.applyMenuAction(menuNavDown)
	}
	if inpututil.IsKeyJustPressed(ebiten.Key1) {
		g.applyMenuAction(menuSetEasy)
	}
	if inpututil.IsKeyJustPressed(ebiten.Key2) {
		g.applyMenuAction(menuSetNormal)
	}
	if inpututil.IsKeyJustPressed(ebiten.Key3) {
		g.applyMenuAction(menuSetHard)
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) || inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		g.applyMenuAction(menuStart)
	}

	if g.menuIndex == 0 || g.menuIndex == 1 || g.menuIndex == 2 || g.menuIndex == 3 {
		if inpututil.IsKeyJustPressed(ebiten.KeyArrowLeft) || inpututil.IsKeyJustPressed(ebiten.KeyA) {
			g.applyMenuAction(menuDec)
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyArrowRight) || inpututil.IsKeyJustPressed(ebiten.KeyD) {
			g.applyMenuAction(menuInc)
		}
	}
}

func (g *game) applyMenuAction(action menuAction) {
	switch action {
	case menuNavUp:
		g.menuIndex--
		if g.menuIndex < 0 {
			g.menuIndex = menuItemCount - 1
		}
		g.playSFX(sfxMenuMove)
	case menuNavDown:
		g.menuIndex++
		if g.menuIndex >= menuItemCount {
			g.menuIndex = 0
		}
		g.playSFX(sfxMenuMove)
	case menuDec:
		handled := false
		if g.menuIndex == 0 && g.difficulty > diffEasy {
			g.difficulty--
			g.totalWaves = g.maxWaveByDifficulty()
			g.enemyBase = g.enemyBaseByDifficulty()
			g.markMenuRequireRestart()
			g.playSFX(sfxMenuMove)
			handled = true
		}
		if g.menuIndex == 1 && g.totalWaves > matchWaveMin {
			g.totalWaves--
			g.markMenuRequireRestart()
			g.playSFX(sfxMenuMove)
			handled = true
		}
		if g.menuIndex == 2 {
			g.toggleSoundEnabled()
			g.playSFX(sfxMenuConfirm)
			handled = true
		}
		if g.menuIndex == 3 {
			changed := g.adjustSoundVolume(-25)
			if changed {
				g.playSFX(sfxMenuMove)
				handled = true
			}
		}
		if !handled {
			g.playSFX(sfxMenuBlocked)
		}
	case menuInc:
		handled := false
		if g.menuIndex == 0 && g.difficulty < diffHard {
			g.difficulty++
			g.totalWaves = g.maxWaveByDifficulty()
			g.enemyBase = g.enemyBaseByDifficulty()
			g.markMenuRequireRestart()
			g.playSFX(sfxMenuMove)
			handled = true
		}
		if g.menuIndex == 1 && g.totalWaves < matchWaveMax {
			g.totalWaves++
			g.markMenuRequireRestart()
			g.playSFX(sfxMenuMove)
			handled = true
		}
		if g.menuIndex == 2 {
			g.toggleSoundEnabled()
			g.playSFX(sfxMenuConfirm)
			handled = true
		}
		if g.menuIndex == 3 {
			changed := g.adjustSoundVolume(25)
			if changed {
				g.playSFX(sfxMenuMove)
				handled = true
			}
		}
		if !handled {
			g.playSFX(sfxMenuBlocked)
		}
	case menuSetEasy:
		g.difficulty = diffEasy
		g.totalWaves = g.maxWaveByDifficulty()
		g.enemyBase = g.enemyBaseByDifficulty()
		g.markMenuRequireRestart()
		g.playSFX(sfxMenuConfirm)
	case menuSetNormal:
		g.difficulty = diffNormal
		g.totalWaves = g.maxWaveByDifficulty()
		g.enemyBase = g.enemyBaseByDifficulty()
		g.markMenuRequireRestart()
		g.playSFX(sfxMenuConfirm)
	case menuSetHard:
		g.difficulty = diffHard
		g.totalWaves = g.maxWaveByDifficulty()
		g.enemyBase = g.enemyBaseByDifficulty()
		g.markMenuRequireRestart()
		g.playSFX(sfxMenuConfirm)
	case menuStart:
		g.playSFX(sfxMenuConfirm)
		g.startMatch()
	}
}
